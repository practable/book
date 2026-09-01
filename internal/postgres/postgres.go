package postgres

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/practable/book/internal/interval"
	"github.com/practable/book/internal/store"
	"gopkg.in/yaml.v2"
)

const migrationLock int64 = 0x626f6f6b5f6d6967
const maintenanceLock int64 = 0x626f6f6b5f6d6169

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Repository struct {
	pool             *pgxpool.Pool
	operationTimeout time.Duration
}

func Open(ctx context.Context, databaseURL string, maxConnections int32, operationTimeout time.Duration) (*Repository, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}
	if maxConnections > 0 {
		config.MaxConns = maxConnections
	}
	config.ConnConfig.RuntimeParams["search_path"] = "public,pg_catalog"
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET TIME ZONE 'UTC'")
		return err
	}
	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	if operationTimeout <= 0 {
		operationTimeout = 30 * time.Second
	}
	repository := &Repository{pool: pool, operationTimeout: operationTimeout}
	if err := repository.Migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return repository, nil
}

func (r *Repository) Close() { r.pool.Close() }

func (r *Repository) Migrate(ctx context.Context) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLock); err != nil {
		return err
	}
	defer conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", migrationLock) //nolint:errcheck
	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS public.schema_migrations (
		version bigint PRIMARY KEY, name text NOT NULL, checksum text NOT NULL,
		applied_at timestamptz NOT NULL DEFAULT clock_timestamp())`); err != nil {
		return err
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".sql") || strings.HasSuffix(name, ".down.sql") {
			continue
		}
		var version int64
		if _, err := fmt.Sscanf(name, "%d_", &version); err != nil {
			return fmt.Errorf("invalid migration filename %q", name)
		}
		body, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(body)
		checksum := hex.EncodeToString(digest[:])
		var existing string
		err = conn.QueryRow(ctx, "SELECT checksum FROM public.schema_migrations WHERE version=$1", version).Scan(&existing)
		if err == nil {
			if existing != checksum {
				return fmt.Errorf("migration %d checksum differs from the applied migration", version)
			}
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, string(body)); err == nil {
			_, err = tx.Exec(ctx, "INSERT INTO public.schema_migrations(version,name,checksum) VALUES($1,$2,$3)", version, name, checksum)
		}
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			_ = tx.Rollback(ctx)
		}
		if err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}
	return nil
}

func (r *Repository) Load(ctx context.Context) (store.PersistentState, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	return loadState(ctx, r.pool)
}

func (r *Repository) ActiveManifestVersion(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	var version int64
	err := r.pool.QueryRow(ctx, "SELECT version FROM public.active_manifest WHERE singleton").Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return version, err
}

type stateQueryer interface {
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

func loadState(ctx context.Context, queryer stateQueryer) (store.PersistentState, error) {
	rows, err := queryer.Query(ctx, bookingSelect+" WHERE NOT superseded ORDER BY row_id")
	if err != nil {
		return store.PersistentState{}, err
	}
	defer rows.Close()
	state := store.PersistentState{
		Groups:               make(map[string][]string),
		ResourceAvailability: make(map[string]store.AvailabilityStatus),
		SlotAvailability:     make(map[string]store.AvailabilityStatus),
	}
	for rows.Next() {
		persisted, err := scanBooking(rows)
		if err != nil {
			return store.PersistentState{}, err
		}
		state.Bookings = append(state.Bookings, persisted)
	}
	if err := rows.Err(); err != nil {
		return store.PersistentState{}, err
	}
	state.Maintenance.Message = "Welcome to the interval booking store"
	err = queryer.QueryRow(ctx, "SELECT booking_creation_paused,welcome_message FROM public.service_state WHERE singleton").Scan(&state.Maintenance.Locked, &state.Maintenance.Message)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return store.PersistentState{}, err
	}
	groupRows, err := queryer.Query(ctx, "SELECT user_name, group_name FROM public.user_groups ORDER BY user_name, group_name")
	if err != nil {
		return store.PersistentState{}, err
	}
	defer groupRows.Close()
	for groupRows.Next() {
		var user, group string
		if err := groupRows.Scan(&user, &group); err != nil {
			return store.PersistentState{}, err
		}
		state.Groups[user] = append(state.Groups[user], group)
	}
	if err := groupRows.Err(); err != nil {
		return store.PersistentState{}, err
	}
	if err := loadAvailability(ctx, queryer, "public.resource_availability", "resource_name", state.ResourceAvailability); err != nil {
		return store.PersistentState{}, err
	}
	if err := loadAvailability(ctx, queryer, "public.slot_availability", "slot_name", state.SlotAvailability); err != nil {
		return store.PersistentState{}, err
	}
	var persisted store.PersistentManifest
	var document string
	err = queryer.QueryRow(ctx, `SELECT m.version,m.document,m.checksum,m.activated_at
		FROM public.active_manifest a JOIN public.manifest_versions m ON m.version=a.version
		WHERE a.singleton`).Scan(&persisted.Version, &document, &persisted.Checksum, &persisted.ActivatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return state, nil
	}
	if err != nil {
		return store.PersistentState{}, err
	}
	digest := sha256.Sum256([]byte(document))
	if hex.EncodeToString(digest[:]) != persisted.Checksum {
		return store.PersistentState{}, errors.New("active manifest checksum mismatch")
	}
	if err := yaml.Unmarshal([]byte(document), &persisted.Manifest); err != nil {
		return store.PersistentState{}, fmt.Errorf("decode active manifest: %w", err)
	}
	state.Manifest = &persisted
	return state, nil
}

func loadAvailability(ctx context.Context, queryer stateQueryer, table, nameColumn string, target map[string]store.AvailabilityStatus) error {
	rows, err := queryer.Query(ctx, "SELECT "+nameColumn+",available,reason FROM "+table+" ORDER BY "+nameColumn)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var name, reason string
		var available bool
		if err := rows.Scan(&name, &available, &reason); err != nil {
			return err
		}
		target[name] = store.AvailabilityStatus{Available: available, Reason: reason}
	}
	return rows.Err()
}

type rowScanner interface{ Scan(...interface{}) error }

func scanBooking(row rowScanner) (store.PersistentBooking, error) {
	var revision int64
	var name, collection, user, policy, slot, resource, cancelledBy, cancelledByText, startedAtText string
	var constrained, unfulfilled, started, cancelled, maintenance bool
	var startsNS, endsNS, chargeNS int64
	var startedNS, cancelledNS *int64
	if err := row.Scan(&revision, &name, &collection, &user, &policy, &slot, &resource, &constrained,
		&startsNS, &endsNS, &startedNS, &cancelledNS, &cancelledBy, &unfulfilled, &chargeNS,
		&started, &startedAtText, &cancelled, &cancelledByText, &maintenance); err != nil {
		return store.PersistentBooking{}, err
	}
	booking := store.Booking{Name: name, User: user, Policy: policy, Slot: slot,
		Unfulfilled: unfulfilled, Maintenance: maintenance, When: interval.Interval{Start: fromNS(startsNS), End: fromNS(endsNS)}}
	booking.Started = started
	booking.StartedAt = startedAtText
	if booking.StartedAt == "" && startedNS != nil {
		booking.StartedAt = fromNS(*startedNS).Format(time.RFC3339Nano)
	}
	booking.Cancelled = cancelled
	if cancelledByText != "" {
		booking.CancelledBy = cancelledByText
	}
	if cancelledNS != nil {
		booking.CancelledAt = fromNS(*cancelledNS)
		booking.CancelledBy = cancelledBy
		booking.UsageCharged = time.Duration(chargeNS)
	}
	return store.PersistentBooking{Booking: booking, Revision: revision, Resource: resource,
		ResourceConstrained: constrained, Current: collection == "live", UsageCharge: time.Duration(chargeNS)}, nil
}

func (r *Repository) CreateBooking(ctx context.Context, request store.CreateBookingRequest) (store.PersistentBooking, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.PersistentBooking{}, false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if err := lockCreate(ctx, tx, request.Booking.User, request.Booking.Policy, request.Resource); err != nil {
		return store.PersistentBooking{}, false, err
	}
	if err := assertManifestVersion(ctx, tx, request.ManifestVersion); err != nil {
		return store.PersistentBooking{}, false, err
	}
	if !request.Maintenance {
		if err := assertAvailable(ctx, tx, request.Resource, request.Booking.Slot); err != nil {
			return store.PersistentBooking{}, false, err
		}
	}

	row := tx.QueryRow(ctx, bookingSelect+` WHERE collection='live' AND NOT superseded
		AND user_name=$1 AND policy_name=$2 AND slot_name=$3 AND starts_ns=$4 AND ends_ns=$5`,
		request.Booking.User, request.Booking.Policy, request.Booking.Slot,
		request.Booking.When.Start.UnixNano(), request.Booking.When.End.UnixNano())
	if existing, scanErr := scanBooking(row); scanErr == nil {
		if err := tx.Commit(ctx); err != nil {
			return store.PersistentBooking{}, false, err
		}
		return existing, false, nil
	} else if !errors.Is(scanErr, pgx.ErrNoRows) {
		return store.PersistentBooking{}, false, scanErr
	}

	var sameName int
	err = tx.QueryRow(ctx, "SELECT 1 FROM public.bookings WHERE name=$1 AND collection='live' AND NOT superseded", request.Booking.Name).Scan(&sameName)
	if err == nil {
		return store.PersistentBooking{}, false, store.ErrBookingIDConflict
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return store.PersistentBooking{}, false, err
	}

	if request.EnforceMaxBookings {
		var count int64
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM public.bookings WHERE NOT superseded
			AND collection='live' AND user_name=$1 AND policy_name=$2 AND ends_ns >= $3`,
			request.Booking.User, request.Booking.Policy, request.Now.UnixNano()).Scan(&count); err != nil {
			return store.PersistentBooking{}, false, err
		}
		if count >= request.MaxBookings {
			return store.PersistentBooking{}, false, store.ErrMaxBookings
		}
	}
	if request.EnforceMaxUsage {
		var usage int64
		if err := tx.QueryRow(ctx, `SELECT coalesce(sum(usage_charge_ns),0) FROM public.bookings
			WHERE NOT superseded AND user_name=$1 AND policy_name=$2`, request.Booking.User, request.Booking.Policy).Scan(&usage); err != nil {
			return store.PersistentBooking{}, false, err
		}
		if time.Duration(usage)+request.Booking.When.End.Sub(request.Booking.When.Start) > request.MaxUsage {
			return store.PersistentBooking{}, false, store.ErrMaxUsage
		}
	}

	rowID, err := insertBooking(ctx, tx, request, "live", "created")
	if err != nil {
		return store.PersistentBooking{}, false, mapWriteError(err)
	}
	if err := insertEvent(ctx, tx, rowID, request.Booking.Name, "created", request.Now, request.Booking.User); err != nil {
		return store.PersistentBooking{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.PersistentBooking{}, false, mapWriteError(err)
	}
	return store.PersistentBooking{Booking: request.Booking, Revision: rowID, Resource: request.Resource,
		ResourceConstrained: request.ResourceConstrained, Current: true,
		UsageCharge: request.Booking.When.End.Sub(request.Booking.When.Start)}, true, nil
}

const bookingSelect = `SELECT row_id, name, collection, user_name, policy_name, slot_name,
	resource_name, resource_constrained, starts_ns, ends_ns, started_at_ns,
	cancelled_at_ns, cancelled_by, unfulfilled, usage_charge_ns,
	started, started_at_text, cancelled, cancelled_by_text, maintenance FROM public.bookings`

func (r *Repository) GetBooking(ctx context.Context, name string) (store.PersistentBooking, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	persisted, err := scanBooking(r.pool.QueryRow(ctx, bookingSelect+` WHERE name=$1
		AND collection='live' AND NOT superseded`, name))
	if errors.Is(err, pgx.ErrNoRows) {
		return store.PersistentBooking{}, store.ErrPersistentNotFound
	}
	return persisted, err
}

func bookingMatchesRequest(persisted store.PersistentBooking, request store.CreateBookingRequest) bool {
	b := persisted.Booking
	r := request.Booking
	return b.Name == r.Name && b.User == r.User && b.Policy == r.Policy && b.Slot == r.Slot &&
		b.When.Start.Equal(r.When.Start) && b.When.End.Equal(r.When.End) && b.Maintenance == r.Maintenance &&
		!b.Started && !b.Cancelled && !b.Unfulfilled &&
		persisted.Resource == request.Resource && persisted.ResourceConstrained == request.ResourceConstrained
}

func lockDomainKeys(ctx context.Context, tx pgx.Tx, keys ...string) error {
	sort.Strings(keys)
	for i, key := range keys {
		if i > 0 && key == keys[i-1] {
			continue
		}
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", key); err != nil {
			return err
		}
	}
	return nil
}

func replacementRetry(ctx context.Context, tx pgx.Tx, originalName string, expectedRevision int64, request store.CreateBookingRequest) (store.PersistentBooking, bool, error) {
	var newRevision int64
	err := tx.QueryRow(ctx, `SELECT r.new_booking_row_id FROM public.booking_replacements r
		JOIN public.bookings old ON old.row_id=r.old_booking_row_id
		WHERE r.old_booking_row_id=$1 AND old.name=$2`, expectedRevision, originalName).Scan(&newRevision)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.PersistentBooking{}, false, store.ErrBookingRevision
	}
	if err != nil {
		return store.PersistentBooking{}, false, err
	}
	persisted, err := scanBooking(tx.QueryRow(ctx, bookingSelect+" WHERE row_id=$1", newRevision))
	if err != nil {
		return store.PersistentBooking{}, false, err
	}
	if !bookingMatchesRequest(persisted, request) {
		return store.PersistentBooking{}, false, store.ErrBookingRevision
	}
	return persisted, false, nil
}

func (r *Repository) ReplaceBooking(ctx context.Context, originalName string, expectedRevision int64, request store.CreateBookingRequest) (store.PersistentBooking, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	if request.Booking.Started || request.Booking.Cancelled || request.Booking.Unfulfilled ||
		request.Booking.StartedAt != "" || !request.Booking.CancelledAt.IsZero() || request.Booking.CancelledBy != "" {
		return store.PersistentBooking{}, false, errors.New("replacement must be an unstarted, uncancelled booking")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.PersistentBooking{}, false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock_shared($1)", maintenanceLock); err != nil {
		return store.PersistentBooking{}, false, err
	}
	if err := assertManifestVersion(ctx, tx, request.ManifestVersion); err != nil {
		return store.PersistentBooking{}, false, err
	}
	if !request.Maintenance {
		if err := assertAvailable(ctx, tx, request.Resource, request.Booking.Slot); err != nil {
			return store.PersistentBooking{}, false, err
		}
	}
	var oldUser, oldPolicy, oldResource string
	err = tx.QueryRow(ctx, `SELECT user_name,policy_name,resource_name FROM public.bookings
		WHERE row_id=$1 AND name=$2 AND collection='live' AND NOT superseded`, expectedRevision, originalName).
		Scan(&oldUser, &oldPolicy, &oldResource)
	if errors.Is(err, pgx.ErrNoRows) {
		persisted, fresh, retryErr := replacementRetry(ctx, tx, originalName, expectedRevision, request)
		if retryErr != nil {
			return store.PersistentBooking{}, false, retryErr
		}
		if err := tx.Commit(ctx); err != nil {
			return store.PersistentBooking{}, false, err
		}
		return persisted, fresh, nil
	}
	if err != nil {
		return store.PersistentBooking{}, false, err
	}
	if err := lockDomainKeys(ctx, tx,
		"resource:"+oldResource, "resource:"+request.Resource,
		userPolicyLock(oldUser, oldPolicy), userPolicyLock(request.Booking.User, request.Booking.Policy)); err != nil {
		return store.PersistentBooking{}, false, err
	}
	old, err := scanBooking(tx.QueryRow(ctx, bookingSelect+` WHERE row_id=$1 AND name=$2
		AND collection='live' AND NOT superseded FOR UPDATE`, expectedRevision, originalName))
	if errors.Is(err, pgx.ErrNoRows) {
		return store.PersistentBooking{}, false, store.ErrBookingRevision
	}
	if err != nil {
		return store.PersistentBooking{}, false, err
	}
	if old.Booking.Started {
		return store.PersistentBooking{}, false, store.ErrBookingStarted
	}
	if _, err := tx.Exec(ctx, "UPDATE public.bookings SET superseded=true,updated_at=clock_timestamp() WHERE row_id=$1", expectedRevision); err != nil {
		return store.PersistentBooking{}, false, err
	}
	if request.EnforceMaxBookings {
		var count int64
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM public.bookings WHERE NOT superseded
			AND collection='live' AND user_name=$1 AND policy_name=$2 AND ends_ns >= $3`,
			request.Booking.User, request.Booking.Policy, request.Now.UnixNano()).Scan(&count); err != nil {
			return store.PersistentBooking{}, false, err
		}
		if count >= request.MaxBookings {
			return store.PersistentBooking{}, false, store.ErrMaxBookings
		}
	}
	if request.EnforceMaxUsage {
		var usage int64
		if err := tx.QueryRow(ctx, `SELECT coalesce(sum(usage_charge_ns),0) FROM public.bookings
			WHERE NOT superseded AND user_name=$1 AND policy_name=$2`, request.Booking.User, request.Booking.Policy).Scan(&usage); err != nil {
			return store.PersistentBooking{}, false, err
		}
		if time.Duration(usage)+request.Booking.When.End.Sub(request.Booking.When.Start) > request.MaxUsage {
			return store.PersistentBooking{}, false, store.ErrMaxUsage
		}
	}
	newRevision, err := insertBooking(ctx, tx, request, "live", "created")
	if err != nil {
		return store.PersistentBooking{}, false, mapWriteError(err)
	}
	when := request.Now.UTC()
	if err := insertEvent(ctx, tx, expectedRevision, originalName, "superseded", when, "admin-edit"); err != nil {
		return store.PersistentBooking{}, false, err
	}
	if err := insertEvent(ctx, tx, newRevision, request.Booking.Name, "created", when, "admin-edit"); err != nil {
		return store.PersistentBooking{}, false, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO public.booking_replacements
		(old_booking_row_id,new_booking_row_id,replaced_at,replaced_at_ns,actor)
		VALUES($1,$2,$3,$4,'admin-edit')`, expectedRevision, newRevision, when, when.UnixNano()); err != nil {
		return store.PersistentBooking{}, false, err
	}
	persisted, err := scanBooking(tx.QueryRow(ctx, bookingSelect+" WHERE row_id=$1", newRevision))
	if err != nil {
		return store.PersistentBooking{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.PersistentBooking{}, false, mapWriteError(err)
	}
	return persisted, true, nil
}

func lockCreate(ctx context.Context, tx pgx.Tx, user, policy, resource string) error {
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock_shared($1)", maintenanceLock); err != nil {
		return err
	}
	keys := []string{"resource:" + resource, userPolicyLock(user, policy)}
	sort.Strings(keys)
	for _, key := range keys {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", key); err != nil {
			return err
		}
	}
	return nil
}

func assertManifestVersion(ctx context.Context, tx pgx.Tx, expected int64) error {
	var active int64
	err := tx.QueryRow(ctx, "SELECT version FROM public.active_manifest WHERE singleton").Scan(&active)
	if errors.Is(err, pgx.ErrNoRows) {
		active = 0
	} else if err != nil {
		return err
	}
	if active != expected {
		return store.ErrStaleManifest
	}
	return nil
}

func assertAvailable(ctx context.Context, tx pgx.Tx, resource, slot string) error {
	for _, target := range []struct {
		table, column, name string
	}{
		{"public.resource_availability", "resource_name", resource},
		{"public.slot_availability", "slot_name", slot},
	} {
		var available bool
		var reason string
		err := tx.QueryRow(ctx, "SELECT available,reason FROM "+target.table+" WHERE "+target.column+"=$1", target.name).Scan(&available, &reason)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		if !available {
			if reason == "" {
				return errors.New("unavailable")
			}
			return errors.New("unavailable because " + reason)
		}
	}
	return nil
}

func insertBooking(ctx context.Context, tx pgx.Tx, request store.CreateBookingRequest, collection, event string) (int64, error) {
	b := request.Booking
	charge := b.When.End.Sub(b.When.Start)
	var rowID int64
	var cancelledAt interface{}
	var cancelledNS interface{}
	databaseCancelledBy := b.CancelledBy
	if !b.CancelledAt.IsZero() {
		cancelledAt, cancelledNS = b.CancelledAt.UTC(), b.CancelledAt.UnixNano()
	} else {
		databaseCancelledBy = ""
	}
	var startedAt interface{}
	var startedNS interface{}
	if parsed, err := time.Parse(time.RFC3339Nano, b.StartedAt); err == nil {
		startedAt, startedNS = parsed.UTC(), parsed.UnixNano()
	}
	err := tx.QueryRow(ctx, `INSERT INTO public.bookings
		(name,collection,user_name,policy_name,slot_name,resource_name,resource_constrained,maintenance,
		 starts_at,ends_at,starts_ns,ends_ns,unfulfilled,usage_charge_ns,
		 started,started_at_text,started_at,started_at_ns,cancelled,cancelled_at,cancelled_at_ns,cancelled_by,cancelled_by_text)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23) RETURNING row_id`,
		b.Name, collection, b.User, b.Policy, b.Slot, request.Resource, request.ResourceConstrained, b.Maintenance,
		b.When.Start.UTC(), b.When.End.UTC(), b.When.Start.UnixNano(), b.When.End.UnixNano(), b.Unfulfilled, charge.Nanoseconds(),
		b.Started, b.StartedAt, startedAt, startedNS, b.Cancelled, cancelledAt, cancelledNS, databaseCancelledBy, b.CancelledBy).Scan(&rowID)
	return rowID, err
}

func insertEvent(ctx context.Context, tx pgx.Tx, rowID int64, name, kind string, at time.Time, actor string) error {
	_, err := tx.Exec(ctx, `INSERT INTO public.booking_events
		(booking_row_id,booking_name,event_type,occurred_at,occurred_at_ns,actor)
		VALUES($1,$2,$3,$4,$5,$6)`, rowID, name, kind, at.UTC(), at.UnixNano(), actor)
	return err
}

func (r *Repository) CancelBooking(ctx context.Context, name string, at time.Time, actor string, charge time.Duration, manifestVersion int64) (store.PersistentBooking, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.PersistentBooking{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock_shared($1)", maintenanceLock); err != nil {
		return store.PersistentBooking{}, err
	}
	if err := assertManifestVersion(ctx, tx, manifestVersion); err != nil {
		return store.PersistentBooking{}, err
	}
	var user, policy string
	if err := tx.QueryRow(ctx, `SELECT user_name,policy_name FROM public.bookings
		WHERE name=$1 AND collection='live' AND NOT superseded`, name).Scan(&user, &policy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.PersistentBooking{}, store.ErrPersistentNotFound
		}
		return store.PersistentBooking{}, err
	}
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", userPolicyLock(user, policy)); err != nil {
		return store.PersistentBooking{}, err
	}
	var rowID int64
	var started bool
	var endsAt time.Time
	if err := tx.QueryRow(ctx, `SELECT row_id,started,ends_at FROM public.bookings
		WHERE name=$1 AND collection='live' AND NOT superseded FOR UPDATE`, name).Scan(&rowID, &started, &endsAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.PersistentBooking{}, store.ErrPersistentNotFound
		}
		return store.PersistentBooking{}, err
	}
	if started {
		if _, err := tx.Exec(ctx, `INSERT INTO public.relay_revocations
			(booking_row_id, booking_name, expires_at, expires_at_ns, revoked_by)
			VALUES($1,$2,$3,$4,$5) ON CONFLICT (booking_row_id) DO NOTHING`,
			rowID, name, endsAt.UTC(), endsAt.UnixNano(), actor); err != nil {
			return store.PersistentBooking{}, err
		}
	}
	if err := supersedeHistoryName(ctx, tx, name, at, actor); err != nil {
		return store.PersistentBooking{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE public.bookings SET collection='history', cancelled=true, cancelled_at=$2,
		cancelled_at_ns=$3,cancelled_by=$4,cancelled_by_text=$4,usage_charge_ns=$5,updated_at=clock_timestamp() WHERE row_id=$1`,
		rowID, at.UTC(), at.UnixNano(), actor, charge.Nanoseconds()); err != nil {
		return store.PersistentBooking{}, mapWriteError(err)
	}
	if err := insertEvent(ctx, tx, rowID, name, "cancelled", at, actor); err != nil {
		return store.PersistentBooking{}, err
	}
	row := tx.QueryRow(ctx, bookingSelect+" WHERE row_id=$1", rowID)
	persisted, err := scanBooking(row)
	if err != nil {
		return store.PersistentBooking{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.PersistentBooking{}, mapWriteError(err)
	}
	return persisted, nil
}

func (r *Repository) StartBooking(ctx context.Context, name string, at time.Time, manifestVersion int64) (store.PersistentBooking, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.PersistentBooking{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock_shared($1)", maintenanceLock); err != nil {
		return store.PersistentBooking{}, err
	}
	if err := assertManifestVersion(ctx, tx, manifestVersion); err != nil {
		return store.PersistentBooking{}, err
	}
	var rowID int64
	var resource, slot string
	var maintenance bool
	var existing *int64
	var alreadyStarted bool
	if err := tx.QueryRow(ctx, `SELECT row_id,resource_name,slot_name,started,started_at_ns,maintenance FROM public.bookings
		WHERE name=$1 AND collection='live' AND NOT superseded FOR UPDATE`, name).Scan(&rowID, &resource, &slot, &alreadyStarted, &existing, &maintenance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.PersistentBooking{}, store.ErrPersistentNotFound
		}
		return store.PersistentBooking{}, err
	}
	if !maintenance {
		if err := assertAvailable(ctx, tx, resource, slot); err != nil {
			return store.PersistentBooking{}, err
		}
	}
	if !alreadyStarted {
		if _, err := tx.Exec(ctx, "UPDATE public.bookings SET started=true,started_at_text=$2,started_at=$3,started_at_ns=$4,updated_at=clock_timestamp() WHERE row_id=$1", rowID, at.UTC().Format(time.RFC3339Nano), at.UTC(), at.UnixNano()); err != nil {
			return store.PersistentBooking{}, err
		}
		if err := insertEvent(ctx, tx, rowID, name, "started", at, "user"); err != nil {
			return store.PersistentBooking{}, err
		}
	}
	persisted, err := scanBooking(tx.QueryRow(ctx, bookingSelect+" WHERE row_id=$1", rowID))
	if err != nil {
		return store.PersistentBooking{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.PersistentBooking{}, err
	}
	return persisted, nil
}

func (r *Repository) ExpireBookings(ctx context.Context, now time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock_shared($1)", maintenanceLock); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT row_id,name,user_name,policy_name FROM public.bookings
		WHERE collection='live' AND NOT superseded AND ends_ns < $1`, now.UnixNano())
	if err != nil {
		return err
	}
	type expired struct {
		id         int64
		name, user string
		policy     string
	}
	var values []expired
	for rows.Next() {
		var value expired
		if err := rows.Scan(&value.id, &value.name, &value.user, &value.policy); err != nil {
			rows.Close()
			return err
		}
		values = append(values, value)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	lockKeys := make([]string, 0, len(values))
	for _, value := range values {
		lockKeys = append(lockKeys, userPolicyLock(value.user, value.policy))
	}
	sort.Strings(lockKeys)
	for i, key := range lockKeys {
		if i > 0 && key == lockKeys[i-1] {
			continue
		}
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", key); err != nil {
			return err
		}
	}
	for _, value := range values {
		var exists int
		if err := tx.QueryRow(ctx, `SELECT 1 FROM public.bookings
			WHERE row_id=$1 AND collection='live' AND NOT superseded FOR UPDATE`, value.id).Scan(&exists); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return err
		}
		if err := supersedeHistoryName(ctx, tx, value.name, now, "system"); err != nil {
			return err
		}
		result, err := tx.Exec(ctx, `UPDATE public.bookings SET collection='history',updated_at=clock_timestamp()
			WHERE row_id=$1 AND collection='live' AND NOT superseded`, value.id)
		if err != nil {
			return err
		}
		if result.RowsAffected() == 0 {
			continue
		}
		if err := insertEvent(ctx, tx, value.id, value.name, "expired", now, "system"); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func supersedeHistoryName(ctx context.Context, tx pgx.Tx, name string, at time.Time, actor string) error {
	rows, err := tx.Query(ctx, `UPDATE public.bookings SET superseded=true,updated_at=clock_timestamp()
		WHERE name=$1 AND collection='history' AND NOT superseded RETURNING row_id`, name)
	if err != nil {
		return err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, id := range ids {
		if err := insertEvent(ctx, tx, id, name, "superseded", at, actor); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ReplaceBookings(ctx context.Context, requests []store.CreateBookingRequest, manifestVersion int64) error {
	return r.replace(ctx, "live", requests, nil, manifestVersion)
}

func (r *Repository) ReplaceOldBookings(ctx context.Context, bookings []store.PersistentBooking, manifestVersion int64) error {
	requests := make([]store.CreateBookingRequest, 0, len(bookings))
	for _, value := range bookings {
		requests = append(requests, store.CreateBookingRequest{Booking: value.Booking, Resource: value.Resource, ResourceConstrained: value.ResourceConstrained})
	}
	return r.replace(ctx, "history", requests, bookings, manifestVersion)
}

func (r *Repository) replace(ctx context.Context, collection string, requests []store.CreateBookingRequest, old []store.PersistentBooking, manifestVersion int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", maintenanceLock); err != nil {
		return err
	}
	if err := assertManifestVersion(ctx, tx, manifestVersion); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, "UPDATE public.bookings SET superseded=true,updated_at=clock_timestamp() WHERE collection=$1 AND NOT superseded RETURNING row_id,name", collection)
	if err != nil {
		return err
	}
	type superseded struct {
		id   int64
		name string
	}
	var displaced []superseded
	for rows.Next() {
		var value superseded
		if err := rows.Scan(&value.id, &value.name); err != nil {
			rows.Close()
			return err
		}
		displaced = append(displaced, value)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	when := time.Now().UTC()
	for _, value := range displaced {
		if err := insertEvent(ctx, tx, value.id, value.name, "superseded", when, "admin"); err != nil {
			return err
		}
	}
	for i, request := range requests {
		rowID, err := insertBooking(ctx, tx, request, collection, "imported")
		if err != nil {
			return mapWriteError(err)
		}
		charge := request.Booking.When.End.Sub(request.Booking.When.Start)
		if collection == "history" && old != nil {
			charge = old[i].UsageCharge
		}
		if _, err := tx.Exec(ctx, "UPDATE public.bookings SET usage_charge_ns=$2 WHERE row_id=$1", rowID, charge.Nanoseconds()); err != nil {
			return err
		}
		if err := insertEvent(ctx, tx, rowID, request.Booking.Name, "imported", when, "admin"); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) ReplaceManifest(ctx context.Context, manifest store.Manifest, validate store.ManifestValidator) (store.PersistentManifest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	document, err := yaml.Marshal(manifest)
	if err != nil {
		return store.PersistentManifest{}, fmt.Errorf("encode manifest: %w", err)
	}
	digest := sha256.Sum256(document)
	checksum := hex.EncodeToString(digest[:])
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.PersistentManifest{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", maintenanceLock); err != nil {
		return store.PersistentManifest{}, err
	}
	state, err := loadState(ctx, tx)
	if err != nil {
		return store.PersistentManifest{}, err
	}
	if validate != nil {
		if err := validate(state); err != nil {
			return store.PersistentManifest{}, err
		}
	}
	if state.Manifest != nil && state.Manifest.Checksum == checksum {
		if err := tx.Commit(ctx); err != nil {
			return store.PersistentManifest{}, err
		}
		return *state.Manifest, nil
	}
	var persisted store.PersistentManifest
	persisted.Manifest = manifest
	persisted.Checksum = checksum
	if err := tx.QueryRow(ctx, `INSERT INTO public.manifest_versions(document,checksum)
		VALUES($1,$2) RETURNING version,activated_at`, string(document), checksum).
		Scan(&persisted.Version, &persisted.ActivatedAt); err != nil {
		return store.PersistentManifest{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO public.active_manifest(singleton,version) VALUES(true,$1)
		ON CONFLICT (singleton) DO UPDATE SET version=EXCLUDED.version`, persisted.Version); err != nil {
		return store.PersistentManifest{}, err
	}
	if _, err := tx.Exec(ctx, "SELECT pg_notify('book_manifest_changed',$1)", fmt.Sprint(persisted.Version)); err != nil {
		return store.PersistentManifest{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.PersistentManifest{}, err
	}
	return persisted, nil
}

func (r *Repository) SetResourceAvailability(ctx context.Context, resource string, available bool, reason string, manifestVersion int64) error {
	return r.setAvailability(ctx, "public.resource_availability", "resource_name", resource, available, reason, manifestVersion)
}

func (r *Repository) SetMaintenance(ctx context.Context, paused bool, message *string) (store.PersistentMaintenance, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return store.PersistentMaintenance{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", maintenanceLock); err != nil {
		return store.PersistentMaintenance{}, err
	}
	var result store.PersistentMaintenance
	if message == nil {
		err = tx.QueryRow(ctx, `UPDATE public.service_state SET booking_creation_paused=$1,updated_at=clock_timestamp(),updated_at_ns=$2 WHERE singleton RETURNING booking_creation_paused,welcome_message`, paused, time.Now().UTC().UnixNano()).Scan(&result.Locked, &result.Message)
	} else {
		err = tx.QueryRow(ctx, `UPDATE public.service_state SET booking_creation_paused=$1,welcome_message=$2,updated_at=clock_timestamp(),updated_at_ns=$3 WHERE singleton RETURNING booking_creation_paused,welcome_message`, paused, *message, time.Now().UTC().UnixNano()).Scan(&result.Locked, &result.Message)
	}
	if err != nil {
		return store.PersistentMaintenance{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return store.PersistentMaintenance{}, err
	}
	return result, nil
}

func (r *Repository) SetSlotAvailability(ctx context.Context, slot string, available bool, reason string, manifestVersion int64) error {
	return r.setAvailability(ctx, "public.slot_availability", "slot_name", slot, available, reason, manifestVersion)
}

func (r *Repository) setAvailability(ctx context.Context, table, nameColumn, name string, available bool, reason string, manifestVersion int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", maintenanceLock); err != nil {
		return err
	}
	if err := assertManifestVersion(ctx, tx, manifestVersion); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "INSERT INTO "+table+"("+nameColumn+",available,reason,updated_at_ns) VALUES($1,$2,$3,$4) ON CONFLICT ("+nameColumn+") DO UPDATE SET available=EXCLUDED.available,reason=EXCLUDED.reason,updated_at=clock_timestamp(),updated_at_ns=EXCLUDED.updated_at_ns", name, available, reason, time.Now().UTC().UnixNano())
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) GrantGroup(ctx context.Context, user, group string, manifestVersion int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock_shared($1)", maintenanceLock); err != nil {
		return err
	}
	if err := assertManifestVersion(ctx, tx, manifestVersion); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO public.user_groups(user_name,group_name) VALUES($1,$2)
		ON CONFLICT (user_name,group_name) DO NOTHING`, user, group); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *Repository) RevokeGroup(ctx context.Context, user, group string, manifestVersion int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock_shared($1)", maintenanceLock); err != nil {
		return err
	}
	if err := assertManifestVersion(ctx, tx, manifestVersion); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM public.user_groups WHERE user_name=$1 AND group_name=$2", user, group); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23P01":
			return store.ErrBookingConflict
		case "23505":
			return store.ErrBookingIDConflict
		}
	}
	return err
}

func fromNS(value int64) time.Time { return time.Unix(0, value).UTC() }

func userPolicyLock(user, policy string) string {
	return fmt.Sprintf("user-policy:%d:%s:%d:%s", len(user), user, len(policy), policy)
}

var _ store.BookingRepository = (*Repository)(nil)
