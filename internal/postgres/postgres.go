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
)

const migrationLock int64 = 0x626f6f6b5f6d6967
const maintenanceLock int64 = 0x626f6f6b5f6d6169

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Repository struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string, maxConnections int32) (*Repository, error) {
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
	repository := &Repository{pool: pool}
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
	rows, err := r.pool.Query(ctx, `SELECT name, collection, user_name, policy_name, slot_name,
		resource_name, resource_constrained, starts_ns, ends_ns, started_at_ns,
		cancelled_at_ns, cancelled_by, unfulfilled, usage_charge_ns,
		started, started_at_text, cancelled, cancelled_by_text
		FROM public.bookings WHERE NOT superseded ORDER BY row_id`)
	if err != nil {
		return store.PersistentState{}, err
	}
	defer rows.Close()
	state := store.PersistentState{Groups: make(map[string][]string)}
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
	groupRows, err := r.pool.Query(ctx, "SELECT user_name, group_name FROM public.user_groups ORDER BY user_name, group_name")
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
	return state, groupRows.Err()
}

type rowScanner interface{ Scan(...interface{}) error }

func scanBooking(row rowScanner) (store.PersistentBooking, error) {
	var name, collection, user, policy, slot, resource, cancelledBy, cancelledByText, startedAtText string
	var constrained, unfulfilled, started, cancelled bool
	var startsNS, endsNS, chargeNS int64
	var startedNS, cancelledNS *int64
	if err := row.Scan(&name, &collection, &user, &policy, &slot, &resource, &constrained,
		&startsNS, &endsNS, &startedNS, &cancelledNS, &cancelledBy, &unfulfilled, &chargeNS,
		&started, &startedAtText, &cancelled, &cancelledByText); err != nil {
		return store.PersistentBooking{}, err
	}
	booking := store.Booking{Name: name, User: user, Policy: policy, Slot: slot,
		Unfulfilled: unfulfilled, When: interval.Interval{Start: fromNS(startsNS), End: fromNS(endsNS)}}
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
	return store.PersistentBooking{Booking: booking, Resource: resource,
		ResourceConstrained: constrained, Current: collection == "live", UsageCharge: time.Duration(chargeNS)}, nil
}

func (r *Repository) CreateBooking(ctx context.Context, request store.CreateBookingRequest) (store.PersistentBooking, bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.PersistentBooking{}, false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if err := lockCreate(ctx, tx, request.Booking.User, request.Booking.Policy, request.Resource); err != nil {
		return store.PersistentBooking{}, false, err
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
	return store.PersistentBooking{Booking: request.Booking, Resource: request.Resource,
		ResourceConstrained: request.ResourceConstrained, Current: true,
		UsageCharge: request.Booking.When.End.Sub(request.Booking.When.Start)}, true, nil
}

const bookingSelect = `SELECT name, collection, user_name, policy_name, slot_name,
	resource_name, resource_constrained, starts_ns, ends_ns, started_at_ns,
	cancelled_at_ns, cancelled_by, unfulfilled, usage_charge_ns,
	started, started_at_text, cancelled, cancelled_by_text FROM public.bookings`

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
		(name,collection,user_name,policy_name,slot_name,resource_name,resource_constrained,
		 starts_at,ends_at,starts_ns,ends_ns,unfulfilled,usage_charge_ns,
		 started,started_at_text,started_at,started_at_ns,cancelled,cancelled_at,cancelled_at_ns,cancelled_by,cancelled_by_text)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22) RETURNING row_id`,
		b.Name, collection, b.User, b.Policy, b.Slot, request.Resource, request.ResourceConstrained,
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

func (r *Repository) CancelBooking(ctx context.Context, name string, at time.Time, actor string, charge time.Duration) (store.PersistentBooking, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.PersistentBooking{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock_shared($1)", maintenanceLock); err != nil {
		return store.PersistentBooking{}, err
	}
	var rowID int64
	var user, policy string
	if err := tx.QueryRow(ctx, `SELECT row_id,user_name,policy_name FROM public.bookings
		WHERE name=$1 AND collection='live' AND NOT superseded FOR UPDATE`, name).Scan(&rowID, &user, &policy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.PersistentBooking{}, store.ErrPersistentNotFound
		}
		return store.PersistentBooking{}, err
	}
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", userPolicyLock(user, policy)); err != nil {
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

func (r *Repository) StartBooking(ctx context.Context, name string, at time.Time) (store.PersistentBooking, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.PersistentBooking{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	var rowID int64
	var existing *int64
	var alreadyStarted bool
	if err := tx.QueryRow(ctx, `SELECT row_id,started,started_at_ns FROM public.bookings
		WHERE name=$1 AND collection='live' AND NOT superseded FOR UPDATE`, name).Scan(&rowID, &alreadyStarted, &existing); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.PersistentBooking{}, store.ErrPersistentNotFound
		}
		return store.PersistentBooking{}, err
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
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock_shared($1)", maintenanceLock); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `UPDATE public.bookings SET collection='history',updated_at=clock_timestamp()
		WHERE collection='live' AND NOT superseded AND ends_ns < $1 RETURNING row_id,name`, now.UnixNano())
	if err != nil {
		return err
	}
	type expired struct {
		id   int64
		name string
	}
	var values []expired
	for rows.Next() {
		var value expired
		if err := rows.Scan(&value.id, &value.name); err != nil {
			rows.Close()
			return err
		}
		values = append(values, value)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, value := range values {
		if err := insertEvent(ctx, tx, value.id, value.name, "expired", now, "system"); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) ReplaceBookings(ctx context.Context, requests []store.CreateBookingRequest) error {
	return r.replace(ctx, "live", requests, nil)
}

func (r *Repository) ReplaceOldBookings(ctx context.Context, bookings []store.PersistentBooking) error {
	requests := make([]store.CreateBookingRequest, 0, len(bookings))
	for _, value := range bookings {
		requests = append(requests, store.CreateBookingRequest{Booking: value.Booking, Resource: value.Resource, ResourceConstrained: value.ResourceConstrained})
	}
	return r.replace(ctx, "history", requests, bookings)
}

func (r *Repository) replace(ctx context.Context, collection string, requests []store.CreateBookingRequest, old []store.PersistentBooking) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", maintenanceLock); err != nil {
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

func (r *Repository) GrantGroup(ctx context.Context, user, group string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO public.user_groups(user_name,group_name) VALUES($1,$2)
		ON CONFLICT (user_name,group_name) DO NOTHING`, user, group)
	return err
}
func (r *Repository) RevokeGroup(ctx context.Context, user, group string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM public.user_groups WHERE user_name=$1 AND group_name=$2", user, group)
	return err
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
