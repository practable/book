package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/practable/book/internal/interval"
	"github.com/practable/book/internal/store"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func integrationRepository(t *testing.T) *Repository {
	t.Helper()
	url := os.Getenv("BOOK_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("BOOK_TEST_DATABASE_URL is not set")
	}
	repository, err := Open(context.Background(), url, 8, 10*time.Second)
	require.NoError(t, err)
	_, err = repository.pool.Exec(context.Background(),
		"TRUNCATE public.booking_events, public.bookings, public.user_groups, public.active_manifest, public.manifest_versions RESTART IDENTITY")
	require.NoError(t, err)
	t.Cleanup(repository.Close)
	return repository
}

func TestMigrationsFromEmptyDatabase(t *testing.T) {
	adminURL := os.Getenv("BOOK_TEST_ADMIN_DATABASE_URL")
	if adminURL == "" {
		t.Skip("BOOK_TEST_ADMIN_DATABASE_URL is not set")
	}
	ctx := context.Background()
	admin, err := pgxpool.Connect(ctx, adminURL)
	require.NoError(t, err)
	t.Cleanup(admin.Close)
	name := fmt.Sprintf("book_migration_%d", time.Now().UnixNano())
	identifier := pgx.Identifier{name}.Sanitize()
	_, err = admin.Exec(ctx, "CREATE DATABASE "+identifier)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = admin.Exec(context.Background(),
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1", name)
		_, _ = admin.Exec(context.Background(), "DROP DATABASE IF EXISTS "+identifier)
	})
	databaseURL, err := url.Parse(adminURL)
	require.NoError(t, err)
	databaseURL.Path = "/" + name
	repository, err := Open(ctx, databaseURL.String(), 2, 10*time.Second)
	require.NoError(t, err)
	var versions int
	require.NoError(t, repository.pool.QueryRow(ctx,
		"SELECT count(*) FROM public.schema_migrations").Scan(&versions))
	require.Equal(t, 3, versions)
	var constraintExists bool
	require.NoError(t, repository.pool.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM pg_constraint WHERE conname='bookings_no_resource_overlap')`).Scan(&constraintExists))
	require.True(t, constraintExists)
	repository.Close()
}

func request(name, user, slot, resource string, start, end time.Time) store.CreateBookingRequest {
	return store.CreateBookingRequest{
		Booking: store.Booking{Name: name, User: user, Policy: "policy", Slot: slot,
			When: interval.Interval{Start: start, End: end}},
		Resource: resource, ResourceConstrained: true, Now: start.Add(-time.Hour),
	}
}

func TestMigrationAndRestartRecovery(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	var versions int
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.schema_migrations").Scan(&versions))
	require.GreaterOrEqual(t, versions, 1)

	start := time.Date(2026, 9, 1, 10, 0, 0, 123, time.UTC)
	created, fresh, err := repository.CreateBooking(ctx, request("restart", "opaque-user", "slot-a", "resource-a", start, start.Add(time.Hour)))
	require.NoError(t, err)
	require.True(t, fresh)
	_, err = repository.StartBooking(ctx, created.Booking.Name, start.Add(time.Minute), 0)
	require.NoError(t, err)
	require.NoError(t, repository.GrantGroup(ctx, "opaque-user", "course-group", 0))

	url := os.Getenv("BOOK_TEST_DATABASE_URL")
	repository.Close()
	reopened, err := Open(ctx, url, 4, 10*time.Second)
	require.NoError(t, err)
	t.Cleanup(reopened.Close)
	state, err := reopened.Load(ctx)
	require.NoError(t, err)
	require.Len(t, state.Bookings, 1)
	require.True(t, state.Bookings[0].Booking.Started)
	require.Equal(t, start.UnixNano(), state.Bookings[0].Booking.When.Start.UnixNano())
	require.Equal(t, []string{"course-group"}, state.Groups["opaque-user"])
}

func TestTimestampsRoundTripAsUTCInstants(t *testing.T) {
	repository := integrationRepository(t)
	location := time.FixedZone("UTC+05:30", 5*60*60+30*60)
	start := time.Date(2026, 9, 1, 15, 30, 0, 456, location)
	end := start.Add(45 * time.Minute)
	_, _, err := repository.CreateBooking(context.Background(),
		request("utc-round-trip", "opaque-user", "slot-a", "resource-a", start, end))
	require.NoError(t, err)

	state, err := repository.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, state.Bookings, 1)
	when := state.Bookings[0].Booking.When
	require.Equal(t, start.UnixNano(), when.Start.UnixNano())
	require.Equal(t, end.UnixNano(), when.End.UnixNano())
	require.Equal(t, time.UTC, when.Start.Location())
	require.Equal(t, time.UTC, when.End.Location())
}

func TestStoreProjectionSurvivesProcessRestart(t *testing.T) {
	repository := integrationRepository(t)
	manifestBytes, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)
	var manifest store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &manifest))
	now := time.Date(2022, 11, 5, 10, 0, 0, 0, time.UTC)

	first := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, first.WithRepository(repository))
	require.NoError(t, first.ReplaceManifest(manifest))
	require.Positive(t, first.ManifestVersion())
	when := interval.Interval{Start: now.Add(time.Hour), End: now.Add(70 * time.Minute)}
	created, err := first.MakeBookingWithName("sl-a", "opaque-restart-user", when, "restart-booking", false)
	require.NoError(t, err)
	require.Equal(t, "restart-booking", created.Name)

	second := store.New().WithNow(func() time.Time { return now })
	// A new process has no local manifest. Attaching PostgreSQL must restore the
	// active manifest before replaying the durable booking.
	require.NoError(t, second.WithRepository(repository))
	require.Equal(t, first.ManifestVersion(), second.ManifestVersion())
	recovered, err := second.GetBooking("restart-booking")
	require.NoError(t, err)
	require.Equal(t, created, recovered)
}

func TestRestartDurablyExpiresOverdueGraceBooking(t *testing.T) {
	repository := integrationRepository(t)
	manifestBytes, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)
	var manifest store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &manifest))
	policy := manifest.Policies["p-a"]
	policy.EnforceGracePeriod = true
	policy.GracePeriod = 5 * time.Minute
	manifest.Policies["p-a"] = policy
	base := time.Date(2022, 11, 5, 10, 0, 0, 0, time.UTC)
	first := store.New().WithNow(func() time.Time { return base })
	require.NoError(t, first.WithRepository(repository))
	require.NoError(t, first.ReplaceManifest(manifest))
	when := interval.Interval{Start: base.Add(time.Hour), End: base.Add(70 * time.Minute)}
	_, err = first.MakeBookingWithName("sl-a", "grace-user", when, "overdue-grace", false)
	require.NoError(t, err)

	restartNow := base.Add(66 * time.Minute)
	second := store.New().WithNow(func() time.Time { return restartNow })
	require.NoError(t, second.WithRepository(repository))
	_, err = second.GetBooking("overdue-grace")
	require.Error(t, err)
	history := second.ExportOldBookings()
	require.True(t, history["overdue-grace"].Cancelled)
	require.Equal(t, "auto-grace-check", history["overdue-grace"].CancelledBy)
}

func TestManifestActivationIsIdempotentAndRefreshesAnotherInstance(t *testing.T) {
	repository := integrationRepository(t)
	manifestBytes, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)
	var initial store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &initial))
	now := time.Date(2022, 11, 5, 10, 0, 0, 0, time.UTC)

	first := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, first.WithRepository(repository))
	require.NoError(t, first.ReplaceManifest(initial))
	versionOne := first.ManifestVersion()
	require.Positive(t, versionOne)
	require.NoError(t, first.ReplaceManifest(initial))
	require.Equal(t, versionOne, first.ManifestVersion())

	databaseURL := os.Getenv("BOOK_TEST_DATABASE_URL")
	secondRepository, err := Open(context.Background(), databaseURL, 4, 10*time.Second)
	require.NoError(t, err)
	t.Cleanup(secondRepository.Close)
	second := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, second.WithRepository(secondRepository))
	require.Equal(t, versionOne, second.ManifestVersion())

	var changed store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &changed))
	description := changed.Descriptions["d-r-a"]
	description.Short = description.Short + " (updated)"
	changed.Descriptions["d-r-a"] = description
	require.NoError(t, first.ReplaceManifest(changed))
	versionTwo := first.ManifestVersion()
	require.Greater(t, versionTwo, versionOne)

	require.NoError(t, second.RefreshManifest(context.Background()))
	require.Equal(t, versionTwo, second.ManifestVersion())
	require.Equal(t, description.Short, second.ExportManifest().Descriptions["d-r-a"].Short)

	stale := request("stale-manifest-request", "opaque-user", "sl-a", "r-a", now.Add(time.Hour), now.Add(70*time.Minute))
	stale.ManifestVersion = versionOne
	_, _, err = secondRepository.CreateBooking(context.Background(), stale)
	require.ErrorIs(t, err, store.ErrStaleManifest)
}

func TestInvalidManifestReplacementLeavesActiveVersionUnchanged(t *testing.T) {
	repository := integrationRepository(t)
	manifestBytes, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)
	var initial store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &initial))
	now := time.Date(2022, 11, 5, 10, 0, 0, 0, time.UTC)
	s := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, s.WithRepository(repository))
	require.NoError(t, s.ReplaceManifest(initial))
	version := s.ManifestVersion()
	_, err = s.MakeBookingWithName("sl-a", "opaque-user", interval.Interval{
		Start: now.Add(time.Hour), End: now.Add(70 * time.Minute),
	}, "manifest-guard-booking", false)
	require.NoError(t, err)

	// This is structurally valid, but it changes the resource underlying an
	// existing booking. Replay validation must reject it transactionally.
	var incompatible store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &incompatible))
	slot := incompatible.Slots["sl-a"]
	slot.Resource = "r-b"
	incompatible.Slots["sl-a"] = slot
	err = s.ReplaceManifest(incompatible)
	require.Error(t, err)
	require.Equal(t, version, s.ManifestVersion())

	active, err := repository.ActiveManifestVersion(context.Background())
	require.NoError(t, err)
	require.Equal(t, version, active)
	state, err := repository.Load(context.Background())
	require.NoError(t, err)
	require.NotNil(t, state.Manifest)
	require.Equal(t, "r-a", state.Manifest.Manifest.Slots["sl-a"].Resource)
}

// TestExternalManifestPersistenceRoundTrip lets operators exercise a real,
// potentially large manifest without copying production configuration into
// this repository. The test is skipped unless BOOK_TEST_MANIFEST_PATH is set.
func TestExternalManifestPersistenceRoundTrip(t *testing.T) {
	path := os.Getenv("BOOK_TEST_MANIFEST_PATH")
	if path == "" {
		t.Skip("BOOK_TEST_MANIFEST_PATH is not set")
	}
	repository := integrationRepository(t)
	document, err := os.ReadFile(path)
	require.NoError(t, err)
	var manifest store.Manifest
	require.NoError(t, yaml.Unmarshal(document, &manifest))

	first := store.New()
	require.NoError(t, first.WithRepository(repository))
	require.NoError(t, first.ReplaceManifest(manifest))
	version := first.ManifestVersion()
	require.Positive(t, version)

	databaseURL := os.Getenv("BOOK_TEST_DATABASE_URL")
	reopened, err := Open(context.Background(), databaseURL, 4, 10*time.Second)
	require.NoError(t, err)
	t.Cleanup(reopened.Close)
	second := store.New()
	require.NoError(t, second.WithRepository(reopened))
	require.Equal(t, version, second.ManifestVersion())
	// Store construction and YAML round-tripping may represent an omitted list
	// as an empty list. Compare their public YAML representations, where those
	// forms are equivalent, while still detecting any changed manifest value.
	expected := store.New()
	require.NoError(t, expected.ReplaceManifest(manifest))
	expectedDocument, err := yaml.Marshal(expected.ExportManifest())
	require.NoError(t, err)
	actualDocument, err := yaml.Marshal(second.ExportManifest())
	require.NoError(t, err)
	require.YAMLEq(t, string(expectedDocument), string(actualDocument))
}

func TestConcurrentOverlapAndExactRetry(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
	requests := []store.CreateBookingRequest{
		request("concurrent-a", "user-a", "slot-a", "resource-a", start, start.Add(time.Hour)),
		request("concurrent-b", "user-b", "slot-b", "resource-a", start.Add(30*time.Minute), start.Add(90*time.Minute)),
	}
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range requests {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _, errs[i] = repository.CreateBooking(ctx, requests[i])
		}(i)
	}
	wg.Wait()
	successes, conflicts := 0, 0
	for _, err := range errs {
		if err == nil {
			successes++
		}
		if errors.Is(err, store.ErrBookingConflict) {
			conflicts++
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)

	state, err := repository.Load(ctx)
	require.NoError(t, err)
	require.Len(t, state.Bookings, 1)
	existing := state.Bookings[0]
	retry := request("different-generated-id", existing.Booking.User, existing.Booking.Slot,
		existing.Resource, existing.Booking.When.Start, existing.Booking.When.End)
	got, fresh, err := repository.CreateBooking(ctx, retry)
	require.NoError(t, err)
	require.False(t, fresh)
	require.Equal(t, existing.Booking.Name, got.Booking.Name)
}

func TestClosedNanosecondBoundariesAndIdentifierConflict(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 2, 14, 0, 0, 123, time.UTC)
	end := start.Add(time.Hour)
	_, _, err := repository.CreateBooking(ctx,
		request("boundary", "user-a", "slot-a", "resource-a", start, end))
	require.NoError(t, err)
	_, _, err = repository.CreateBooking(ctx,
		request("touching", "user-b", "slot-b", "resource-a", end, end.Add(time.Minute)))
	require.ErrorIs(t, err, store.ErrBookingConflict)
	_, _, err = repository.CreateBooking(ctx,
		request("one-nanosecond-gap", "user-c", "slot-c", "resource-a", end.Add(time.Nanosecond), end.Add(time.Minute)))
	require.NoError(t, err)
	differentPayload := request("boundary", "user-d", "slot-d", "resource-b", start.Add(2*time.Hour), start.Add(3*time.Hour))
	_, _, err = repository.CreateBooking(ctx, differentPayload)
	require.ErrorIs(t, err, store.ErrBookingIDConflict)
	state, err := repository.Load(ctx)
	require.NoError(t, err)
	require.Equal(t, start.UnixNano(), state.Bookings[0].Booking.When.Start.UnixNano())
}

func TestCancellationThenRebookingAndHistory(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	first := request("first", "user-a", "slot-a", "resource-a", start, start.Add(time.Hour))
	_, _, err := repository.CreateBooking(ctx, first)
	require.NoError(t, err)
	_, err = repository.CancelBooking(ctx, "first", start.Add(-time.Minute), "user-a", 0, 0)
	require.NoError(t, err)
	second := request("second", "user-b", "slot-b", "resource-a", start, start.Add(time.Hour))
	_, fresh, err := repository.CreateBooking(ctx, second)
	require.NoError(t, err)
	require.True(t, fresh)
	state, err := repository.Load(ctx)
	require.NoError(t, err)
	require.Len(t, state.Bookings, 2)
	current, historical := 0, 0
	for _, booking := range state.Bookings {
		if booking.Current {
			current++
		} else {
			historical++
		}
	}
	require.Equal(t, 1, current)
	require.Equal(t, 1, historical)
}

func TestPolicyLimitsIncludeHistoricalUsage(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	first := request("limit-a", "limited-user", "slot-a", "resource-a", start, start.Add(30*time.Minute))
	first.EnforceMaxBookings, first.MaxBookings = true, 1
	first.EnforceMaxUsage, first.MaxUsage = true, time.Hour
	_, _, err := repository.CreateBooking(ctx, first)
	require.NoError(t, err)
	concurrent := request("limit-b", "limited-user", "slot-b", "resource-b", start.Add(time.Hour), start.Add(90*time.Minute))
	concurrent.EnforceMaxBookings, concurrent.MaxBookings = true, 1
	_, _, err = repository.CreateBooking(ctx, concurrent)
	require.ErrorIs(t, err, store.ErrMaxBookings)

	_, err = repository.CancelBooking(ctx, "limit-a", start.Add(-time.Minute), "limited-user", 20*time.Minute, 0)
	require.NoError(t, err)
	usage := request("usage-b", "limited-user", "slot-b", "resource-b", start.Add(time.Hour), start.Add(110*time.Minute))
	usage.EnforceMaxUsage, usage.MaxUsage = true, time.Hour
	_, _, err = repository.CreateBooking(ctx, usage)
	require.ErrorIs(t, err, store.ErrMaxUsage)
}

func TestConcurrentPolicyLimit(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	requests := []store.CreateBookingRequest{
		request("policy-concurrent-a", "one-user", "slot-a", "resource-a", start, start.Add(10*time.Minute)),
		request("policy-concurrent-b", "one-user", "slot-b", "resource-b", start, start.Add(10*time.Minute)),
	}
	for i := range requests {
		requests[i].EnforceMaxBookings = true
		requests[i].MaxBookings = 1
	}
	var wg sync.WaitGroup
	errs := make([]error, len(requests))
	for i := range requests {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _, errs[i] = repository.CreateBooking(ctx, requests[i])
		}(i)
	}
	wg.Wait()
	successes, limited := 0, 0
	for _, err := range errs {
		if err == nil {
			successes++
		}
		if errors.Is(err, store.ErrMaxBookings) {
			limited++
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, limited)
}

func TestHistoricalImportPreservesLegacyFlags(t *testing.T) {
	repository := integrationRepository(t)
	start := time.Date(2026, 9, 4, 14, 0, 0, 77, time.UTC)
	booking := store.Booking{Name: "legacy", User: "opaque", Policy: "policy", Slot: "slot",
		Started: true, StartedAt: "legacy-start-value", Cancelled: true,
		CancelledBy: "legacy-import", When: interval.Interval{Start: start, End: start.Add(time.Hour)}}
	err := repository.ReplaceOldBookings(context.Background(), []store.PersistentBooking{{
		Booking: booking, Resource: "resource", ResourceConstrained: true,
		Current: false, UsageCharge: 17 * time.Minute,
	}}, 0)
	require.NoError(t, err)
	state, err := repository.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, state.Bookings, 1)
	require.Equal(t, booking, state.Bookings[0].Booking)
	require.Equal(t, 17*time.Minute, state.Bookings[0].UsageCharge)
}

func TestConcurrentExpiryTransitionsOnce(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 4, 16, 0, 0, 0, time.UTC)
	_, _, err := repository.CreateBooking(ctx,
		request("expire-once", "expire-user", "slot", "resource", start, start.Add(time.Minute)))
	require.NoError(t, err)
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = repository.ExpireBookings(ctx, start.Add(2*time.Minute))
		}(i)
	}
	wg.Wait()
	require.NoError(t, errs[0])
	require.NoError(t, errs[1])
	state, err := repository.Load(ctx)
	require.NoError(t, err)
	require.Len(t, state.Bookings, 1)
	require.False(t, state.Bookings[0].Current)
	var events int
	require.NoError(t, repository.pool.QueryRow(ctx,
		"SELECT count(*) FROM public.booking_events WHERE booking_name='expire-once' AND event_type='expired'").Scan(&events))
	require.Equal(t, 1, events)
}

func TestTransactionRollbackAndAtomicReplacement(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 5, 10, 0, 0, 0, time.UTC)
	_, err := repository.pool.Exec(ctx, `CREATE OR REPLACE FUNCTION public.fail_booking_event()
		RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected event failure'; END $$;
		CREATE TRIGGER fail_booking_event BEFORE INSERT ON public.booking_events
		FOR EACH ROW EXECUTE FUNCTION public.fail_booking_event()`)
	require.NoError(t, err)
	_, _, err = repository.CreateBooking(ctx, request("rollback", "user", "slot", "resource", start, start.Add(time.Hour)))
	require.Error(t, err)
	var count int
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.bookings").Scan(&count))
	require.Zero(t, count)
	_, err = repository.pool.Exec(ctx, "DROP TRIGGER fail_booking_event ON public.booking_events; DROP FUNCTION public.fail_booking_event()")
	require.NoError(t, err)

	original := request("original", "user", "slot", "resource", start, start.Add(time.Hour))
	_, _, err = repository.CreateBooking(ctx, original)
	require.NoError(t, err)
	overlapA := request("replacement-a", "user-a", "slot-a", "resource-z", start.Add(2*time.Hour), start.Add(3*time.Hour))
	overlapB := request("replacement-b", "user-b", "slot-b", "resource-z", start.Add(150*time.Minute), start.Add(4*time.Hour))
	err = repository.ReplaceBookings(ctx, []store.CreateBookingRequest{overlapA, overlapB}, 0)
	require.ErrorIs(t, err, store.ErrBookingConflict)
	state, err := repository.Load(ctx)
	require.NoError(t, err)
	require.Len(t, state.Bookings, 1)
	require.Equal(t, "original", state.Bookings[0].Booking.Name)
}
