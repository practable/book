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
	"github.com/practable/book/internal/operations"
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
		"TRUNCATE public.booking_activation_stages, public.booking_activation_runs, public.operational_schedule_occurrences, public.webhook_callback_receipts, public.webhook_deliveries, public.operational_jobs, public.relay_revocations, public.booking_events, public.booking_replacements, public.bookings, public.user_groups, public.resource_availability, public.slot_availability, public.service_state, public.active_manifest, public.manifest_versions RESTART IDENTITY")
	require.NoError(t, err)
	_, err = repository.pool.Exec(context.Background(), "INSERT INTO public.service_state(singleton,updated_at_ns) VALUES(true,0)")
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
	require.Equal(t, 16, versions)
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

func TestBookingEventsAreAvailableForAdminAudit(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	created, _, err := repository.CreateBooking(ctx, request("audit-api", "opaque-user", "slot-a", "resource-a", start, start.Add(time.Hour)))
	require.NoError(t, err)
	_, err = repository.StartBooking(ctx, created.Booking.Name, start.Add(time.Minute), 0)
	require.NoError(t, err)
	_, err = repository.CancelBooking(ctx, created.Booking.Name, start.Add(10*time.Minute), "technician: repair", 10*time.Minute, 0)
	require.NoError(t, err)
	events, err := repository.ListBookingEvents(ctx, created.Booking.Name)
	require.NoError(t, err)
	require.Equal(t, []string{"created", "started", "cancelled"}, []string{events[0].Type, events[1].Type, events[2].Type})
	require.Equal(t, "technician: repair", events[2].Actor)
}

func TestBookingReplacementRecordsItsActualActor(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	created, _, err := repository.CreateBooking(ctx, request("self-edit", "opaque-user", "slot-a", "resource-a", start, start.Add(time.Hour)))
	require.NoError(t, err)
	replacement := request("self-edit", "opaque-user", "slot-a", "resource-a", start.Add(time.Hour), start.Add(2*time.Hour))
	replacement.Actor = "user:opaque-user"
	_, _, err = repository.ReplaceBooking(ctx, created.Booking.Name, created.Revision, replacement)
	require.NoError(t, err)
	events, err := repository.ListBookingEvents(ctx, created.Booking.Name)
	require.NoError(t, err)
	require.Equal(t, "user:opaque-user", events[len(events)-1].Actor)
}

func operationalJob(now time.Time) (operations.Job, operations.Delivery) {
	job := operations.Job{ID: "job-1", Resource: "ripp02", Workflow: "fill", Kind: "setup", State: "scheduled", DueAt: now,
		StartsAt: now, EndsAt: now.Add(10 * time.Minute), ManifestVersion: 1, PlanRevision: 1, IdempotencyKey: "plan-1/setup", Payload: []byte(`{"reason":"out-of-hours"}`)}
	delivery := operations.Delivery{ID: "delivery-1", JobID: job.ID, Body: []byte(`{"version":1,"job_id":"job-1"}`)}
	return job, delivery
}

func activationRequest(booking, user, key string, now time.Time) operations.CreateActivationRequest {
	jobID := "activation-job-" + key
	body := []byte(`{"version":1,"job_id":"` + jobID + `","workflow":"video-health"}`)
	return operations.CreateActivationRequest{
		RunID: "activation-run-" + key, BookingName: booking, User: user, Resource: "resource-a", Stream: "video",
		Pipeline: "video-activation", IdempotencyKey: key, RequestedAt: now, ResolvedPlan: []byte(`{"pipeline":"video-activation"}`),
		Stages: []operations.ActivationStageSpec{
			{Name: "power-on", JobTemplate: "video-on", Workflow: "video-on", DueAt: now, TimeoutAt: now.Add(4 * time.Second), MaximumAttempts: 1, Parameters: map[string]string{"stream": "camera-a"}, ProgressMessage: "Starting video"},
			{Name: "health", JobTemplate: "video-check", Workflow: "video-health", DueAt: now.Add(time.Second), TimeoutAt: now.Add(5 * time.Second), MaximumAttempts: 3, Parameters: map[string]string{"stream": "camera-a"}, ProgressMessage: "Checking video"},
		},
		FirstJob: operations.Job{ID: jobID, Resource: "resource-a", Workflow: "video-on", Kind: "preflight", State: "reserved", DueAt: now,
			StartsAt: now, EndsAt: now.Add(4 * time.Second), TriggeringBookingName: booking, PlanRevision: 1, IdempotencyKey: "activation-job:" + key, Payload: body},
		FirstDelivery: operations.Delivery{ID: "activation-delivery-" + key, JobID: jobID, Body: body},
	}
}

func TestActivationCreationIsAtomicAndIdempotent(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	_, _, err := repository.CreateBooking(ctx, request("activation-booking", "opaque-user", "slot-a", "resource-a", now.Add(-time.Minute), now.Add(time.Hour)))
	require.NoError(t, err)
	req := activationRequest("activation-booking", "opaque-user", "request-1", now)

	type result struct {
		run   operations.ActivationRun
		fresh bool
		err   error
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			run, fresh, err := repository.CreateActivation(ctx, req)
			results <- result{run, fresh, err}
		}()
	}
	wg.Wait()
	close(results)
	freshCount := 0
	for item := range results {
		require.NoError(t, item.err)
		require.Equal(t, req.RunID, item.run.ID)
		require.Len(t, item.run.Stages, 2)
		if item.fresh {
			freshCount++
		}
	}
	require.Equal(t, 1, freshCount)
	loaded, err := repository.GetActivation(ctx, req.RunID)
	require.NoError(t, err)
	require.Equal(t, "preparing", loaded.State)
	require.Equal(t, "pending", loaded.Stages[0].State)
	require.Equal(t, "waiting", loaded.Stages[1].State)
	_, _, err = repository.CreateActivation(ctx, activationRequest("activation-booking", "opaque-user", "request-2", now))
	require.ErrorIs(t, err, operations.ErrActivationConflict)
}

func TestActivationRollsBackWhenDeliveryInsertFails(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	_, _, err := repository.CreateBooking(ctx, request("rollback-activation", "opaque-user", "slot-a", "resource-a", now.Add(-time.Minute), now.Add(time.Hour)))
	require.NoError(t, err)
	dummyJob, dummyDelivery := operationalJob(now)
	_, _, err = repository.CreateJob(ctx, dummyJob, dummyDelivery)
	require.NoError(t, err)
	req := activationRequest("rollback-activation", "opaque-user", "rollback", now)
	req.FirstDelivery.ID = dummyDelivery.ID
	_, _, err = repository.CreateActivation(ctx, req)
	require.Error(t, err)
	var count int
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.booking_activation_runs WHERE run_id=$1", req.RunID).Scan(&count))
	require.Zero(t, count)
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.operational_jobs WHERE job_id=$1", req.FirstJob.ID).Scan(&count))
	require.Zero(t, count)
}

func TestActivationCallbacksAdvanceStagesRetryAndStartBooking(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	_, _, err := repository.CreateBooking(ctx, request("advance-activation", "opaque-user", "slot-a", "resource-a", now.Add(-time.Minute), now.Add(time.Hour)))
	require.NoError(t, err)
	req := activationRequest("advance-activation", "opaque-user", "advance", now)
	req.Stages[1].InitialDelay = time.Second
	req.Stages[1].Backoff = 2
	req.Stages[1].MaximumDelay = 5 * time.Second
	req.Stages[1].TotalTimeout = time.Minute
	req.Stages[1].RetryableCodes = []string{"not_ready"}
	req.Stages[1].RetryMessage = "Video is not ready; trying again"
	run, _, err := repository.CreateActivation(ctx, req)
	require.NoError(t, err)

	_, _, err = repository.ApplyCallback(ctx, operations.Callback{DeliveryID: "advance-accepted-1", JobID: run.Stages[0].JobID, State: "accepted", At: now}, "hash-accepted-1")
	require.NoError(t, err)
	_, _, err = repository.ApplyCallback(ctx, operations.Callback{DeliveryID: "advance-succeeded-1", JobID: run.Stages[0].JobID, State: "succeeded", At: now.Add(time.Second)}, "hash-succeeded-1")
	require.NoError(t, err)
	run, err = repository.GetActivation(ctx, run.ID)
	require.NoError(t, err)
	require.Equal(t, 1, run.CurrentStage)
	require.Equal(t, "pending", run.Stages[1].State)
	secondJob := run.Stages[1].JobID

	_, _, err = repository.ApplyCallback(ctx, operations.Callback{DeliveryID: "advance-accepted-2", JobID: secondJob, State: "accepted", At: now.Add(2 * time.Second)}, "hash-accepted-2")
	require.NoError(t, err)
	_, _, err = repository.ApplyCallback(ctx, operations.Callback{DeliveryID: "advance-failed-2", JobID: secondJob, State: "failed", At: now.Add(3 * time.Second), Code: "not_ready", Error: "camera stream has not produced a frame"}, "hash-failed-2")
	require.NoError(t, err)
	run, err = repository.GetActivation(ctx, run.ID)
	require.NoError(t, err)
	require.Equal(t, "preparing", run.State)
	require.Equal(t, "pending", run.Stages[1].State)
	require.Equal(t, 2, run.Stages[1].Attempt)
	require.NotEqual(t, secondJob, run.Stages[1].JobID)
	require.Equal(t, req.Stages[1].RetryMessage, run.ProgressMessage)

	retryJob := run.Stages[1].JobID
	_, _, err = repository.ApplyCallback(ctx, operations.Callback{DeliveryID: "advance-accepted-3", JobID: retryJob, State: "accepted", At: now.Add(4 * time.Second)}, "hash-accepted-3")
	require.NoError(t, err)
	_, _, err = repository.ApplyCallback(ctx, operations.Callback{DeliveryID: "advance-succeeded-3", JobID: retryJob, State: "succeeded", At: now.Add(5 * time.Second)}, "hash-succeeded-3")
	require.NoError(t, err)
	run, err = repository.GetActivation(ctx, run.ID)
	require.NoError(t, err)
	require.Equal(t, "active", run.State)
	latest, err := repository.GetLatestActivationForBooking(ctx, req.BookingName)
	require.NoError(t, err)
	require.Equal(t, run.ID, latest.ID)
	booking, err := repository.GetBooking(ctx, req.BookingName)
	require.NoError(t, err)
	require.True(t, booking.Booking.Started)
}

func TestOperationalJobOutboxIsTransactionalIdempotentAndClaimedOnce(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	job, delivery := operationalJob(now)
	created, fresh, err := repository.CreateJob(ctx, job, delivery)
	require.NoError(t, err)
	require.True(t, fresh)
	require.Equal(t, job.ID, created.ID)
	retried, fresh, err := repository.CreateJob(ctx, job, operations.Delivery{ID: "different", JobID: job.ID, Body: delivery.Body})
	require.NoError(t, err)
	require.False(t, fresh)
	require.Equal(t, job.ID, retried.ID)

	var wg sync.WaitGroup
	claimed := make(chan []operations.Delivery, 2)
	for _, owner := range []string{"instance-a", "instance-b"} {
		wg.Add(1)
		go func(owner string) {
			defer wg.Done()
			values, claimErr := repository.ClaimDeliveries(ctx, owner, now, time.Minute, 10)
			require.NoError(t, claimErr)
			claimed <- values
		}(owner)
	}
	wg.Wait()
	close(claimed)
	count := 0
	for values := range claimed {
		count += len(values)
	}
	require.Equal(t, 1, count)
}

func TestBookingAndOperationalGuardsCommitAtomically(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 8, 18, 0, 0, 0, time.UTC)
	primary := request("guarded-user-booking", "user", "slot-a", "resource-a", start, start.Add(time.Hour))
	planned := []store.OperationalReservation{
		operationalReservation("setup-job", "setup-booking", "setup-delivery", "guarded-user-booking", "resource-a", start.Add(-10*time.Minute), start),
		operationalReservation("teardown-job", "teardown-booking", "teardown-delivery", "guarded-user-booking", "resource-a", start.Add(time.Hour), start.Add(70*time.Minute)),
	}
	created, guards, fresh, err := repository.CreateBookingWithOperations(ctx, primary, planned, nil)
	require.NoError(t, err)
	require.True(t, fresh)
	require.Equal(t, "guarded-user-booking", created.Booking.Name)
	require.Len(t, guards, 2)
	var bookingCount, jobCount, deliveryCount int
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.bookings WHERE collection='live' AND NOT superseded").Scan(&bookingCount))
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.operational_jobs").Scan(&jobCount))
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.webhook_deliveries").Scan(&deliveryCount))
	require.Equal(t, 3, bookingCount)
	require.Equal(t, 2, jobCount)
	require.Equal(t, 2, deliveryCount)
}

func TestOperationalGuardFailureRollsBackUserBooking(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 8, 18, 0, 0, 0, time.UTC)
	primary := request("rollback-guarded-booking", "user", "slot-a", "resource-a", start, start.Add(time.Hour))
	// This guard overlaps its triggering booking and must roll back everything.
	planned := []store.OperationalReservation{operationalReservation("bad-job", "bad-booking", "bad-delivery", primary.Booking.Name, "resource-a", start.Add(-time.Minute), start.Add(time.Minute))}
	_, _, _, err := repository.CreateBookingWithOperations(ctx, primary, planned, nil)
	require.ErrorIs(t, err, store.ErrBookingConflict)
	var bookingCount, jobCount int
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.bookings").Scan(&bookingCount))
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.operational_jobs").Scan(&jobCount))
	require.Zero(t, bookingCount)
	require.Zero(t, jobCount)
}

func TestCancellingBookingAtomicallyRetiresUndispatchedOperationalWork(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 8, 18, 0, 0, 0, time.UTC)
	primary := request("cancel-guarded-booking", "user", "slot-a", "resource-a", start, start.Add(time.Hour))
	planned := []store.OperationalReservation{
		operationalReservation("cancel-setup-job", "cancel-setup-booking", "cancel-setup-delivery", primary.Booking.Name, "resource-a", start.Add(-10*time.Minute), start),
		operationalReservation("cancel-teardown-job", "cancel-teardown-booking", "cancel-teardown-delivery", primary.Booking.Name, "resource-a", start.Add(time.Hour), start.Add(70*time.Minute)),
	}
	created, _, fresh, err := repository.CreateBookingWithOperations(ctx, primary, planned, nil)
	require.NoError(t, err)
	require.True(t, fresh)

	cancelled, retired, err := repository.CancelBookingWithOperations(ctx, created.Booking.Name, start.Add(-time.Hour), "admin:test", 0, 0)
	require.NoError(t, err)
	require.True(t, cancelled.Booking.Cancelled)
	require.Len(t, retired, 2)
	var live, cancelledJobs, cancelledDeliveries int
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.bookings WHERE collection='live' AND NOT superseded").Scan(&live))
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.operational_jobs WHERE triggering_booking_name=$1 AND state='cancelled'", primary.Booking.Name).Scan(&cancelledJobs))
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.webhook_deliveries WHERE state='cancelled'").Scan(&cancelledDeliveries))
	require.Zero(t, live)
	require.Equal(t, 2, cancelledJobs)
	require.Equal(t, 2, cancelledDeliveries)
}

func TestCancellingBookingPreservesLeasedOperationalWork(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 8, 18, 0, 0, 0, time.UTC)
	primary := request("cancel-after-dispatch", "user", "slot-a", "resource-a", start, start.Add(time.Hour))
	planned := []store.OperationalReservation{
		operationalReservation("dispatched-setup-job", "dispatched-setup-booking", "dispatched-setup-delivery", primary.Booking.Name, "resource-a", start.Add(-10*time.Minute), start),
		operationalReservation("pending-teardown-job", "pending-teardown-booking", "pending-teardown-delivery", primary.Booking.Name, "resource-a", start.Add(time.Hour), start.Add(70*time.Minute)),
	}
	_, _, _, err := repository.CreateBookingWithOperations(ctx, primary, planned, nil)
	require.NoError(t, err)
	_, err = repository.pool.Exec(ctx, "UPDATE public.webhook_deliveries SET state='leased',lease_owner='worker',lease_until=$1 WHERE job_id='dispatched-setup-job'", start)
	require.NoError(t, err)

	_, retired, err := repository.CancelBookingWithOperations(ctx, primary.Booking.Name, start.Add(-time.Hour), "admin:test", 0, 0)
	require.NoError(t, err)
	require.Len(t, retired, 1)
	require.Equal(t, "pending-teardown-booking", retired[0].Booking.Name)
	var live int
	var setupState string
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.bookings WHERE collection='live' AND NOT superseded").Scan(&live))
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT state FROM public.operational_jobs WHERE job_id='dispatched-setup-job'").Scan(&setupState))
	require.Equal(t, 1, live)
	require.Equal(t, "reserved", setupState)
}

func TestReplacingBookingAtomicallyReplansOperationalWork(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 8, 18, 0, 0, 0, time.UTC)
	original := request("original-with-guards", "user", "slot-a", "resource-a", start, start.Add(time.Hour))
	originalPlan := []store.OperationalReservation{
		operationalReservation("old-setup-job", "old-setup-booking", "old-setup-delivery", original.Booking.Name, "resource-a", start.Add(-10*time.Minute), start),
		operationalReservation("old-teardown-job", "old-teardown-booking", "old-teardown-delivery", original.Booking.Name, "resource-a", start.Add(time.Hour), start.Add(70*time.Minute)),
	}
	created, _, _, err := repository.CreateBookingWithOperations(ctx, original, originalPlan, nil)
	require.NoError(t, err)

	newStart := start.Add(2 * time.Hour)
	replacement := request("replacement-with-guards", "user", "slot-a", "resource-a", newStart, newStart.Add(time.Hour))
	replacementPlan := []store.OperationalReservation{
		operationalReservation("new-setup-job", "new-setup-booking", "new-setup-delivery", replacement.Booking.Name, "resource-a", newStart.Add(-10*time.Minute), newStart),
		operationalReservation("new-teardown-job", "new-teardown-booking", "new-teardown-delivery", replacement.Booking.Name, "resource-a", newStart.Add(time.Hour), newStart.Add(70*time.Minute)),
	}
	replaced, changed, fresh, err := repository.ReplaceBookingWithOperations(ctx, original.Booking.Name, created.Revision, replacement, replacementPlan, nil)
	require.NoError(t, err)
	require.True(t, fresh)
	require.Equal(t, replacement.Booking.Name, replaced.Booking.Name)
	require.Len(t, changed, 4)
	var live, oldCancelled, newJobs int
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.bookings WHERE collection='live' AND NOT superseded").Scan(&live))
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.operational_jobs WHERE triggering_booking_name=$1 AND state='cancelled'", original.Booking.Name).Scan(&oldCancelled))
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.operational_jobs WHERE triggering_booking_name=$1", replacement.Booking.Name).Scan(&newJobs))
	require.Equal(t, 3, live)
	require.Equal(t, 2, oldCancelled)
	require.Equal(t, 2, newJobs)
}

func TestAcceptedOperationalJobActivatesOnlyItsReservation(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 8, 18, 0, 0, 0, time.UTC)
	primary := request("activation-trigger", "user", "slot-a", "resource-a", start.Add(time.Hour), start.Add(2*time.Hour))
	guard := operationalReservation("activation-job", "activation-booking", "activation-delivery", primary.Booking.Name, "resource-a", start, start.Add(10*time.Minute))
	_, _, _, err := repository.CreateBookingWithOperations(ctx, primary, []store.OperationalReservation{guard}, nil)
	require.NoError(t, err)
	_, _, err = repository.ApplyCallback(ctx, operations.Callback{DeliveryID: "accept-activation", JobID: guard.Job.ID, State: "accepted", At: start}, "accept-hash")
	require.NoError(t, err)

	booking, job, err := repository.ActivateOperationalJob(ctx, guard.Job.ID, "activate-1", "body-hash", start.Add(time.Minute))
	require.NoError(t, err)
	require.True(t, booking.Booking.Started)
	require.Equal(t, "running", job.State)
	retried, retriedJob, err := repository.ActivateOperationalJob(ctx, guard.Job.ID, "activate-1", "body-hash", start.Add(2*time.Minute))
	require.NoError(t, err)
	require.Equal(t, booking.Booking.StartedAt, retried.Booking.StartedAt)
	require.Equal(t, "running", retriedJob.State)
	_, _, err = repository.ActivateOperationalJob(ctx, guard.Job.ID, "activate-1", "changed-hash", start.Add(2*time.Minute))
	require.ErrorIs(t, err, operations.ErrCallbackConflict)
	completion := operations.Callback{DeliveryID: "complete-activation", JobID: guard.Job.ID, State: "succeeded", At: start.Add(5 * time.Minute)}
	completed, fresh, err := repository.ApplyCallback(ctx, completion, "complete-hash")
	require.NoError(t, err)
	require.True(t, fresh)
	require.Equal(t, "succeeded", completed.State)
	_, fresh, err = repository.ApplyCallback(ctx, completion, "complete-hash")
	require.NoError(t, err)
	require.False(t, fresh)
	var collection string
	var usage int64
	var revocations int
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT collection,usage_charge_ns FROM public.bookings WHERE name=$1", guard.Request.Booking.Name).Scan(&collection, &usage))
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.relay_revocations WHERE booking_name=$1", guard.Request.Booking.Name).Scan(&revocations))
	require.Equal(t, "history", collection)
	require.Equal(t, (4 * time.Minute).Nanoseconds(), usage)
	require.Equal(t, 1, revocations)
	_, _, err = repository.ActivateOperationalJob(ctx, "missing-job", "activate-missing", "missing-hash", start)
	require.ErrorIs(t, err, operations.ErrNotFound)
}

func TestStoreReturnsActivityForAcceptedOperationalReservation(t *testing.T) {
	repository := integrationRepository(t)
	manifestBytes, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)
	var manifest store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &manifest))
	manifest.OperationalWorkflows = map[string]store.OperationalWorkflow{
		"prepare": {Description: "Prepare equipment", ExpectedDuration: 10 * time.Minute, MaximumDuration: 10 * time.Minute},
	}
	resource := manifest.Resources["r-a"]
	resource.Operations = store.OperationalProfile{BeforeBooking: []store.OperationalGuard{{Workflow: "prepare", Duration: 10 * time.Minute, Applies: store.OperationalAlways}}}
	manifest.Resources["r-a"] = resource
	now := time.Date(2022, 11, 5, 10, 0, 0, 0, time.UTC)
	s := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, s.WithRepository(repository))
	require.NoError(t, s.ReplaceManifest(manifest))
	_, err = s.MakeBookingWithName("sl-a", "runner-user", interval.Interval{Start: now.Add(10 * time.Minute), End: now.Add(20 * time.Minute)}, "runner-trigger", false)
	require.NoError(t, err)
	var jobID string
	require.NoError(t, repository.pool.QueryRow(context.Background(), "SELECT job_id FROM public.operational_jobs WHERE triggering_booking_name='runner-trigger'").Scan(&jobID))
	_, _, err = repository.ApplyCallback(context.Background(), operations.Callback{DeliveryID: "accept-store-runner", JobID: jobID, State: "accepted", At: now}, "accept-store-hash")
	require.NoError(t, err)
	activity, err := s.ActivateOperationalJob(context.Background(), jobID, "activate-store-runner", "activate-store-hash")
	require.NoError(t, err)
	require.NotEmpty(t, activity.BookingID)
	require.Equal(t, now, activity.NotBefore)
}

func TestOperationalSchedulesAreDurableDeduplicatedAndRecordConflicts(t *testing.T) {
	repository := integrationRepository(t)
	manifestBytes, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)
	var manifest store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &manifest))
	manifest.OperationalWorkflows = map[string]store.OperationalWorkflow{
		"daily-check": {Description: "Daily check", ExpectedDuration: 10 * time.Minute, MaximumDuration: 10 * time.Minute},
	}
	recurrence := store.OperationalRecurrence{Timezone: "UTC", StartDate: "2022-11-05", EndDate: "2022-11-05", Weekdays: []string{"sat"}, Time: "10:00"}
	manifest.OperationalSchedules = map[string]store.OperationalSchedule{
		"a-required": {Slot: "sl-a", Workflow: "daily-check", Duration: 10 * time.Minute, Conflict: store.OperationalConflictRequire, Recurrence: recurrence},
		"b-skipped":  {Slot: "sl-a", Workflow: "daily-check", Duration: 10 * time.Minute, Conflict: store.OperationalConflictSkip, Recurrence: recurrence},
	}
	missed := recurrence
	missed.Time = "08:00"
	manifest.OperationalSchedules["c-missed"] = store.OperationalSchedule{Slot: "sl-a", Workflow: "daily-check", Duration: 10 * time.Minute, Conflict: store.OperationalConflictRequire, Recurrence: missed}
	now := time.Date(2022, 11, 5, 9, 0, 0, 0, time.UTC)
	first := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, first.WithRepository(repository))
	require.NoError(t, first.ReplaceManifest(manifest))
	summary, err := first.MaterializeOperationalSchedules(context.Background(), now.Add(-2*time.Hour), now.Add(2*time.Hour))
	require.NoError(t, err)
	require.Equal(t, 1, summary.Planned)
	require.Equal(t, 1, summary.Skipped)
	require.Equal(t, 1, summary.Missed)

	second := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, second.WithRepository(repository))
	summary, err = second.MaterializeOperationalSchedules(context.Background(), now.Add(-2*time.Hour), now.Add(2*time.Hour))
	require.NoError(t, err)
	require.Equal(t, 3, summary.Existing)
	var occurrences, jobs, live int
	require.NoError(t, repository.pool.QueryRow(context.Background(), "SELECT count(*) FROM public.operational_schedule_occurrences").Scan(&occurrences))
	require.NoError(t, repository.pool.QueryRow(context.Background(), "SELECT count(*) FROM public.operational_jobs WHERE workflow_name='daily-check'").Scan(&jobs))
	require.NoError(t, repository.pool.QueryRow(context.Background(), "SELECT count(*) FROM public.bookings WHERE collection='live' AND user_name='__operations__'").Scan(&live))
	require.Equal(t, 3, occurrences)
	require.Equal(t, 1, jobs)
	require.Equal(t, 1, live)
	listed, err := repository.ListScheduleOccurrences(context.Background(), now.Add(-2*time.Hour), now.Add(2*time.Hour), "", 20)
	require.NoError(t, err)
	require.Len(t, listed, 3)
	require.Equal(t, "sl-a", listed[0].Slot)
	require.Equal(t, "r-a", listed[0].Resource)
	require.Equal(t, "daily-check", listed[0].Workflow)
	conflicts, err := repository.ListScheduleOccurrences(context.Background(), now.Add(-2*time.Hour), now.Add(2*time.Hour), "skipped", 20)
	require.NoError(t, err)
	require.Len(t, conflicts, 1)
	require.Equal(t, "b-skipped", conflicts[0].Schedule)
}

func operationalReservation(jobID, bookingID, deliveryID, trigger, resource string, start, end time.Time) store.OperationalReservation {
	body := []byte(`{"version":1}`)
	booking := store.Booking{Name: bookingID, User: "__operations__", Policy: "__operations__", Slot: "slot-a", Maintenance: true, When: interval.Interval{Start: start, End: end}}
	return store.OperationalReservation{
		Request: store.CreateBookingRequest{Booking: booking, Resource: resource, ResourceConstrained: true, Now: start, Maintenance: true},
		Job: operations.Job{ID: jobID, Resource: resource, Workflow: "workflow", Kind: "setup", State: "reserved", DueAt: start, StartsAt: start, EndsAt: end,
			TriggeringBookingName: trigger, IdempotencyKey: "key-" + jobID, Payload: body},
		Delivery: operations.Delivery{ID: deliveryID, JobID: jobID, Body: body},
	}
}

func TestOperationalDeliveryRetryAndCallbackIdempotency(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	job, delivery := operationalJob(now)
	_, _, err := repository.CreateJob(ctx, job, delivery)
	require.NoError(t, err)
	claimed, err := repository.ClaimDeliveries(ctx, "sender", now, time.Minute, 1)
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	require.NoError(t, repository.CompleteDelivery(ctx, delivery.ID, "sender", false, 503, "runner unavailable", now, now.Add(time.Minute)))
	claimed, err = repository.ClaimDeliveries(ctx, "sender", now.Add(30*time.Second), time.Minute, 1)
	require.NoError(t, err)
	require.Empty(t, claimed)
	claimed, err = repository.ClaimDeliveries(ctx, "sender", now.Add(time.Minute), time.Minute, 1)
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	require.NoError(t, repository.CompleteDelivery(ctx, delivery.ID, "sender", true, 202, "", now.Add(time.Minute), now.Add(time.Minute)))

	callback := operations.Callback{DeliveryID: "callback-1", JobID: job.ID, State: "accepted", At: now.Add(2 * time.Minute)}
	updated, fresh, err := repository.ApplyCallback(ctx, callback, "body-hash")
	require.NoError(t, err)
	require.True(t, fresh)
	require.Equal(t, "accepted", updated.State)
	_, fresh, err = repository.ApplyCallback(ctx, callback, "body-hash")
	require.NoError(t, err)
	require.False(t, fresh)
	_, _, err = repository.ApplyCallback(ctx, callback, "different-hash")
	require.ErrorIs(t, err, operations.ErrCallbackConflict)
	_, _, err = repository.ApplyCallback(ctx, operations.Callback{DeliveryID: "callback-2", JobID: job.ID, State: "scheduled", At: now}, "hash-2")
	require.ErrorIs(t, err, operations.ErrInvalidTransition)
}

func TestMaintenanceStateSurvivesRestart(t *testing.T) {
	repository := integrationRepository(t)
	message := "Technicians replacing camera"
	state, err := repository.SetMaintenance(context.Background(), true, &message)
	require.NoError(t, err)
	require.True(t, state.Locked)
	require.Equal(t, message, state.Message)
	repository.Close()
	reopened, err := Open(context.Background(), os.Getenv("BOOK_TEST_DATABASE_URL"), 4, 10*time.Second)
	require.NoError(t, err)
	t.Cleanup(reopened.Close)
	loaded, err := reopened.Load(context.Background())
	require.NoError(t, err)
	require.True(t, loaded.Maintenance.Locked)
	require.Equal(t, message, loaded.Maintenance.Message)
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

func TestSuspensionPreventsTakeUpInsideDatabaseTransaction(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	created, _, err := repository.CreateBooking(ctx, request("suspended-start", "opaque-user", "slot-a", "resource-a", start, start.Add(time.Hour)))
	require.NoError(t, err)
	require.NoError(t, repository.SetResourceAvailability(ctx, "resource-a", false, "failed health check", 0))
	_, err = repository.StartBooking(ctx, created.Booking.Name, start.Add(time.Minute), 0)
	require.EqualError(t, err, "unavailable because failed health check")
	persisted, err := repository.GetBooking(ctx, created.Booking.Name)
	require.NoError(t, err)
	require.False(t, persisted.Booking.Started)
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

func TestStoreBookingCreatesDurableOperationalPlanAndRecovers(t *testing.T) {
	repository := integrationRepository(t)
	manifestBytes, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)
	var manifest store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &manifest))
	manifest.OperationalWorkflows = map[string]store.OperationalWorkflow{
		"prepare": {Description: "Prepare equipment", ExpectedDuration: 10 * time.Minute, MaximumDuration: 10 * time.Minute},
		"settle":  {Description: "Settle equipment", ExpectedDuration: 10 * time.Minute, MaximumDuration: 10 * time.Minute},
	}
	resource := manifest.Resources["r-a"]
	resource.Operations = store.OperationalProfile{
		BeforeBooking: []store.OperationalGuard{{Workflow: "prepare", Duration: 10 * time.Minute, Applies: store.OperationalAlways, Reclaimable: true}},
		AfterBooking:  []store.OperationalGuard{{Workflow: "settle", Duration: 10 * time.Minute, Applies: store.OperationalAlways, Reclaimable: true}},
	}
	manifest.Resources["r-a"] = resource
	now := time.Date(2022, 11, 5, 10, 0, 0, 0, time.UTC)
	first := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, first.WithRepository(repository))
	require.NoError(t, first.ReplaceManifest(manifest))
	when := interval.Interval{Start: now.Add(time.Hour), End: now.Add(70 * time.Minute)}
	_, err = first.MakeBookingWithName("sl-a", "guard-user", when, "guard-trigger", false)
	require.NoError(t, err)
	var bookingCount, jobCount int
	require.NoError(t, repository.pool.QueryRow(context.Background(), "SELECT count(*) FROM public.bookings WHERE collection='live' AND NOT superseded").Scan(&bookingCount))
	require.NoError(t, repository.pool.QueryRow(context.Background(), "SELECT count(*) FROM public.operational_jobs WHERE triggering_booking_name='guard-trigger'").Scan(&jobCount))
	require.Equal(t, 3, bookingCount)
	require.Equal(t, 2, jobCount)
	secondWhen := interval.Interval{Start: when.End.Add(5 * time.Minute), End: when.End.Add(15 * time.Minute)}
	_, err = first.MakeBookingWithName("sl-a", "next-user", secondWhen, "next-trigger", false)
	require.NoError(t, err)
	require.NoError(t, repository.pool.QueryRow(context.Background(), "SELECT count(*) FROM public.bookings WHERE collection='live' AND NOT superseded").Scan(&bookingCount))
	require.Equal(t, 4, bookingCount)
	var cancelledJobs int
	require.NoError(t, repository.pool.QueryRow(context.Background(), "SELECT count(*) FROM public.operational_jobs WHERE triggering_booking_name='guard-trigger' AND state='cancelled'").Scan(&cancelledJobs))
	require.Equal(t, 1, cancelledJobs)

	second := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, second.WithRepository(repository))
	require.Len(t, second.ExportBookings(), 4)
}

func TestAvailabilitySuspensionsAreDurableSharedAndSlotScoped(t *testing.T) {
	repository := integrationRepository(t)
	manifestBytes, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)
	var manifest store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &manifest))
	manifest.Slots["sl-a-alias"] = manifest.Slots["sl-a"]
	policy := manifest.Policies["p-a"]
	policy.Slots = append(policy.Slots, "sl-a-alias")
	manifest.Policies["p-a"] = policy
	now := time.Date(2022, 11, 5, 10, 0, 0, 0, time.UTC)
	first := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, first.WithRepository(repository))
	require.NoError(t, first.ReplaceManifest(manifest))

	secondRepository, err := Open(context.Background(), os.Getenv("BOOK_TEST_DATABASE_URL"), 4, 10*time.Second)
	require.NoError(t, err)
	t.Cleanup(secondRepository.Close)
	second := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, second.WithRepository(secondRepository))

	require.NoError(t, first.SetSlotIsAvailable("sl-a", false, "UI maintenance"))
	require.NoError(t, second.RefreshManifest(context.Background()))
	ok, reason, err := second.GetSlotIsAvailable("sl-a")
	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, "unavailable because UI maintenance", reason)
	ok, _, err = second.GetSlotIsAvailable("sl-a-alias")
	require.NoError(t, err)
	require.True(t, ok)
	_, err = second.MakeBookingWithName("sl-a", "availability-user", interval.Interval{Start: now.Add(time.Hour), End: now.Add(70 * time.Minute)}, "blocked-slot", false)
	require.EqualError(t, err, "unavailable because UI maintenance")
	_, err = second.MakeBookingWithName("sl-a-alias", "availability-user", interval.Interval{Start: now.Add(time.Hour), End: now.Add(70 * time.Minute)}, "allowed-alias", false)
	require.NoError(t, err)

	require.NoError(t, first.SetResourceIsAvailable("r-a", false, "failed health check"))
	require.NoError(t, second.RefreshManifest(context.Background()))
	ok, reason, err = second.GetSlotIsAvailable("sl-a-alias")
	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, "unavailable because failed health check", reason)
	state, err := secondRepository.Load(context.Background())
	require.NoError(t, err)
	require.False(t, state.ResourceAvailability["r-a"].Available)
	require.False(t, state.SlotAvailability["sl-a"].Available)
}

func TestMaintenanceBookingSurvivesSuspensionAndRestart(t *testing.T) {
	repository := integrationRepository(t)
	manifestBytes, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)
	var manifest store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &manifest))
	now := time.Date(2022, 11, 5, 10, 0, 0, 0, time.UTC)
	first := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, first.WithRepository(repository))
	require.NoError(t, first.ReplaceManifest(manifest))
	require.NoError(t, first.SetResourceIsAvailable("r-a", false, "repair"))
	when := interval.Interval{Start: now.Add(time.Hour), End: now.Add(2 * time.Hour)}
	booking, err := first.MakeMaintenanceBooking("sl-a", "tech-a", when)
	require.NoError(t, err)
	require.True(t, booking.Maintenance)
	_, err = first.MakeMaintenanceBooking("sl-a", "tech-b", when)
	require.Error(t, err)

	second := store.New().WithNow(func() time.Time { return when.Start })
	require.NoError(t, second.WithRepository(repository))
	recovered, err := second.GetBooking(booking.Name)
	require.NoError(t, err)
	require.True(t, recovered.Maintenance)
	_, err = second.GetActivity(recovered)
	require.NoError(t, err)
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

func TestIndividualBookingReplacementIsAuditableAndIdempotent(t *testing.T) {
	repository := integrationRepository(t)
	manifestBytes, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)
	var manifest store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &manifest))
	now := time.Date(2022, 11, 5, 10, 0, 0, 0, time.UTC)
	s := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, s.WithRepository(repository))
	require.NoError(t, s.ReplaceManifest(manifest))
	_, err = s.MakeBookingWithName("sl-a", "edit-user", interval.Interval{
		Start: now.Add(time.Hour), End: now.Add(70 * time.Minute),
	}, "edit-me", false)
	require.NoError(t, err)

	edit, err := s.GetBookingForEdit("edit-me")
	require.NoError(t, err)
	oldRevision := edit.Revision
	edit.Booking.When = interval.Interval{Start: now.Add(80 * time.Minute), End: now.Add(90 * time.Minute)}
	replaced, err := s.ReplaceBooking(edit)
	require.NoError(t, err)
	require.NotEqual(t, oldRevision, replaced.Revision)
	require.Equal(t, edit.Booking.When, replaced.Booking.When)

	// Re-sending the exported edit is a retry, not another correction.
	retried, err := s.ReplaceBooking(edit)
	require.NoError(t, err)
	require.Equal(t, replaced, retried)

	var superseded bool
	require.NoError(t, repository.pool.QueryRow(context.Background(),
		"SELECT superseded FROM public.bookings WHERE row_id=$1", oldRevision).Scan(&superseded))
	require.True(t, superseded)
	var replacementRow int64
	require.NoError(t, repository.pool.QueryRow(context.Background(),
		"SELECT new_booking_row_id FROM public.booking_replacements WHERE old_booking_row_id=$1", oldRevision).Scan(&replacementRow))
	require.Equal(t, replaced.Revision, replacementRow)
	var events int
	require.NoError(t, repository.pool.QueryRow(context.Background(),
		"SELECT count(*) FROM public.booking_events WHERE booking_row_id=$1 AND event_type='superseded'", oldRevision).Scan(&events))
	require.Equal(t, 1, events)
}

func TestIndividualBookingReplacementRejectsStartedStaleAndConflictingEdits(t *testing.T) {
	repository := integrationRepository(t)
	manifestBytes, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)
	var manifest store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &manifest))
	now := time.Date(2022, 11, 5, 10, 0, 0, 0, time.UTC)
	s := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, s.WithRepository(repository))
	require.NoError(t, s.ReplaceManifest(manifest))
	first, err := s.MakeBookingWithName("sl-a", "edit-user-a", interval.Interval{
		Start: now.Add(time.Hour), End: now.Add(70 * time.Minute),
	}, "edit-first", false)
	require.NoError(t, err)
	_, err = s.MakeBookingWithName("sl-a", "edit-user-b", interval.Interval{
		Start: now.Add(80 * time.Minute), End: now.Add(90 * time.Minute),
	}, "edit-second", false)
	require.NoError(t, err)
	edit, err := s.GetBookingForEdit(first.Name)
	require.NoError(t, err)

	conflicting := edit
	conflicting.Booking.When = interval.Interval{Start: now.Add(80 * time.Minute), End: now.Add(90 * time.Minute)}
	_, err = s.ReplaceBooking(conflicting)
	require.Error(t, err)
	stillCurrent, err := s.GetBookingForEdit(first.Name)
	require.NoError(t, err)
	require.Equal(t, edit.Revision, stillCurrent.Revision)

	now = first.When.Start
	_, err = s.GetActivity(first)
	require.NoError(t, err)
	_, err = s.ReplaceBooking(edit)
	require.ErrorIs(t, err, store.ErrBookingStarted)

	stale := edit
	stale.Revision++
	_, err = s.ReplaceBooking(stale)
	require.ErrorIs(t, err, store.ErrBookingRevision)
}

func TestConcurrentIndividualBookingEditsAllowOnlyOneWinner(t *testing.T) {
	repository := integrationRepository(t)
	manifestBytes, err := os.ReadFile("../../demo/manifest.yaml")
	require.NoError(t, err)
	var manifest store.Manifest
	require.NoError(t, yaml.Unmarshal(manifestBytes, &manifest))
	now := time.Date(2022, 11, 5, 10, 0, 0, 0, time.UTC)
	first := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, first.WithRepository(repository))
	require.NoError(t, first.ReplaceManifest(manifest))
	_, err = first.MakeBookingWithName("sl-a", "concurrent-edit-user", interval.Interval{
		Start: now.Add(time.Hour), End: now.Add(70 * time.Minute),
	}, "concurrent-edit", false)
	require.NoError(t, err)

	databaseURL := os.Getenv("BOOK_TEST_DATABASE_URL")
	secondRepository, err := Open(context.Background(), databaseURL, 4, 10*time.Second)
	require.NoError(t, err)
	t.Cleanup(secondRepository.Close)
	second := store.New().WithNow(func() time.Time { return now })
	require.NoError(t, second.WithRepository(secondRepository))
	edit, err := first.GetBookingForEdit("concurrent-edit")
	require.NoError(t, err)
	left, right := edit, edit
	left.Booking.When = interval.Interval{Start: now.Add(80 * time.Minute), End: now.Add(90 * time.Minute)}
	right.Booking.When = interval.Interval{Start: now.Add(100 * time.Minute), End: now.Add(110 * time.Minute)}

	results := make([]error, 2)
	var wg sync.WaitGroup
	for i, candidate := range []store.EditableBooking{left, right} {
		wg.Add(1)
		go func(i int, candidate store.EditableBooking) {
			defer wg.Done()
			_, results[i] = mapStoreReplace(first, second, i, candidate)
		}(i, candidate)
	}
	wg.Wait()
	successes, stale := 0, 0
	for _, result := range results {
		if result == nil {
			successes++
		}
		if errors.Is(result, store.ErrBookingRevision) {
			stale++
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, stale)
}

func mapStoreReplace(first, second *store.Store, index int, edit store.EditableBooking) (store.EditableBooking, error) {
	if index == 0 {
		return first.ReplaceBooking(edit)
	}
	return second.ReplaceBooking(edit)
}

func TestIndividualReplacementDatabaseFailureRollsBackOriginal(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	first := request("replace-rollback-a", "user-a", "slot-a", "resource-a", time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC), time.Date(2026, 9, 3, 11, 0, 0, 0, time.UTC))
	second := request("replace-rollback-b", "user-b", "slot-b", "resource-a", first.Booking.When.End.Add(time.Minute), first.Booking.When.End.Add(61*time.Minute))
	_, _, err := repository.CreateBooking(ctx, first)
	require.NoError(t, err)
	_, _, err = repository.CreateBooking(ctx, second)
	require.NoError(t, err)
	persisted, err := repository.GetBooking(ctx, first.Booking.Name)
	require.NoError(t, err)
	conflicting := first
	conflicting.Booking.When = second.Booking.When
	conflicting.ManifestVersion = 0
	_, _, err = repository.ReplaceBooking(ctx, first.Booking.Name, persisted.Revision, conflicting)
	require.ErrorIs(t, err, store.ErrBookingConflict)
	stillCurrent, err := repository.GetBooking(ctx, first.Booking.Name)
	require.NoError(t, err)
	require.Equal(t, persisted.Revision, stillCurrent.Revision)
	var replacements int
	require.NoError(t, repository.pool.QueryRow(ctx, "SELECT count(*) FROM public.booking_replacements").Scan(&replacements))
	require.Zero(t, replacements)
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

func TestHalfOpenBoundariesAndIdentifierConflict(t *testing.T) {
	repository := integrationRepository(t)
	ctx := context.Background()
	start := time.Date(2026, 9, 2, 14, 0, 0, 123, time.UTC)
	end := start.Add(time.Hour)
	_, _, err := repository.CreateBooking(ctx,
		request("boundary", "user-a", "slot-a", "resource-a", start, end))
	require.NoError(t, err)
	_, _, err = repository.CreateBooking(ctx,
		request("touching", "user-b", "slot-b", "resource-a", end, end.Add(time.Minute)))
	require.NoError(t, err)
	_, _, err = repository.CreateBooking(ctx,
		request("overlap-after-boundary", "user-c", "slot-c", "resource-a", end.Add(time.Nanosecond), end.Add(time.Minute)))
	require.ErrorIs(t, err, store.ErrBookingConflict)
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
