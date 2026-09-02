package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/practable/book/internal/interval"
	"github.com/practable/book/internal/operations"
	"github.com/practable/book/internal/store"
)

const jobSelect = `SELECT job_id,resource_name,workflow_name,job_kind,state,due_at,starts_at,ends_at,
	booking_row_id,triggering_booking_name,manifest_version,plan_revision,idempotency_key,payload::text,attempts,last_error
	FROM public.operational_jobs`

func scanJob(row pgx.Row) (operations.Job, error) {
	var job operations.Job
	var payload string
	err := row.Scan(&job.ID, &job.Resource, &job.Workflow, &job.Kind, &job.State, &job.DueAt, &job.StartsAt, &job.EndsAt,
		&job.BookingRowID, &job.TriggeringBookingName, &job.ManifestVersion, &job.PlanRevision, &job.IdempotencyKey, &payload, &job.Attempts, &job.LastError)
	job.Payload = []byte(payload)
	job.DueAt, job.StartsAt, job.EndsAt = job.DueAt.UTC(), job.StartsAt.UTC(), job.EndsAt.UTC()
	return job, err
}

func (r *Repository) CreateJob(ctx context.Context, job operations.Job, delivery operations.Delivery) (operations.Job, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return operations.Job{}, false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	result, err := tx.Exec(ctx, `INSERT INTO public.operational_jobs
		(job_id,resource_name,workflow_name,job_kind,state,due_at,due_at_ns,starts_at,starts_at_ns,ends_at,ends_at_ns,
		 booking_row_id,triggering_booking_name,manifest_version,plan_revision,idempotency_key,payload)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		ON CONFLICT (idempotency_key) DO NOTHING`, job.ID, job.Resource, job.Workflow, job.Kind, job.State, job.DueAt.UTC(), job.DueAt.UnixNano(),
		job.StartsAt.UTC(), job.StartsAt.UnixNano(), job.EndsAt.UTC(), job.EndsAt.UnixNano(), job.BookingRowID, job.TriggeringBookingName,
		job.ManifestVersion, job.PlanRevision, job.IdempotencyKey, string(job.Payload))
	if err != nil {
		return operations.Job{}, false, err
	}
	created := result.RowsAffected() == 1
	if created {
		_, err = tx.Exec(ctx, `INSERT INTO public.webhook_deliveries
			(delivery_id,job_id,direction,state,body,next_attempt_at,next_attempt_at_ns)
			VALUES($1,$2,'book-to-runner','pending',$3,$4,$5)`, delivery.ID, job.ID, string(delivery.Body), job.DueAt.UTC(), job.DueAt.UnixNano())
		if err != nil {
			return operations.Job{}, false, err
		}
	}
	persisted, err := scanJob(tx.QueryRow(ctx, jobSelect+" WHERE idempotency_key=$1", job.IdempotencyKey))
	if err != nil {
		return operations.Job{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return operations.Job{}, false, err
	}
	return persisted, created, nil
}

func (r *Repository) CreateBookingWithOperations(ctx context.Context, request store.CreateBookingRequest, planned []store.OperationalReservation, reclaim []string) (store.PersistentBooking, []store.PersistentBooking, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.PersistentBooking{}, nil, false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if err := supersedeReclaimableOperationsTx(ctx, tx, reclaim, request.Now); err != nil {
		return store.PersistentBooking{}, nil, false, err
	}
	primary, created, err := createBookingTx(ctx, tx, request)
	if err != nil {
		return store.PersistentBooking{}, nil, false, err
	}
	if !created {
		if err := tx.Commit(ctx); err != nil {
			return store.PersistentBooking{}, nil, false, err
		}
		return primary, nil, false, nil
	}
	reservations := make([]store.PersistentBooking, 0, len(planned))
	for _, item := range planned {
		reservation, fresh, err := createBookingTx(ctx, tx, item.Request)
		if err != nil {
			return store.PersistentBooking{}, nil, false, err
		}
		if !fresh {
			return store.PersistentBooking{}, nil, false, store.ErrBookingConflict
		}
		job := item.Job
		job.BookingRowID = &reservation.Revision
		if err := insertOperationalJobTx(ctx, tx, job, item.Delivery); err != nil {
			return store.PersistentBooking{}, nil, false, err
		}
		if err := insertGuardOperationalUsageTx(ctx, tx, job, primary); err != nil {
			return store.PersistentBooking{}, nil, false, err
		}
		reservations = append(reservations, reservation)
	}
	if err := tx.Commit(ctx); err != nil {
		return store.PersistentBooking{}, nil, false, mapWriteError(err)
	}
	return primary, reservations, true, nil
}

func (r *Repository) CancelBookingWithOperations(ctx context.Context, name string, at time.Time, actor string, charge time.Duration, manifestVersion int64) (store.PersistentBooking, []store.PersistentBooking, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.PersistentBooking{}, nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock_shared($1)", maintenanceLock); err != nil {
		return store.PersistentBooking{}, nil, err
	}
	if err := assertManifestVersion(ctx, tx, manifestVersion); err != nil {
		return store.PersistentBooking{}, nil, err
	}
	retired, err := retireTriggeredOperationalReservationsTx(ctx, tx, name, at, actor)
	if err != nil {
		return store.PersistentBooking{}, nil, err
	}
	persisted, err := cancelBookingTx(ctx, tx, name, at, actor, charge)
	if err != nil {
		return store.PersistentBooking{}, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.PersistentBooking{}, nil, mapWriteError(err)
	}
	return persisted, retired, nil
}

func (r *Repository) ReplaceBookingWithOperations(ctx context.Context, originalName string, expectedRevision int64, request store.CreateBookingRequest, planned []store.OperationalReservation, reclaim []string) (store.PersistentBooking, []store.PersistentBooking, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	if request.Booking.Started || request.Booking.Cancelled || request.Booking.Unfulfilled ||
		request.Booking.StartedAt != "" || !request.Booking.CancelledAt.IsZero() || request.Booking.CancelledBy != "" {
		return store.PersistentBooking{}, nil, false, errors.New("replacement must be an unstarted, uncancelled booking")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.PersistentBooking{}, nil, false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock_shared($1)", maintenanceLock); err != nil {
		return store.PersistentBooking{}, nil, false, err
	}
	if err := assertManifestVersion(ctx, tx, request.ManifestVersion); err != nil {
		return store.PersistentBooking{}, nil, false, err
	}
	if err := stopBookingActivationsTx(ctx, tx, originalName, request.Now, "cancelled", "booking_rescheduled", "The booking was rescheduled"); err != nil {
		return store.PersistentBooking{}, nil, false, err
	}
	retired, err := retireTriggeredOperationalReservationsTx(ctx, tx, originalName, request.Now, request.Actor)
	if err != nil {
		return store.PersistentBooking{}, nil, false, err
	}
	if err := supersedeReclaimableOperationsTx(ctx, tx, reclaim, request.Now); err != nil {
		return store.PersistentBooking{}, nil, false, err
	}
	persisted, fresh, err := replaceBookingTx(ctx, tx, originalName, expectedRevision, request)
	if err != nil {
		return store.PersistentBooking{}, nil, false, err
	}
	if !fresh {
		if err := tx.Commit(ctx); err != nil {
			return store.PersistentBooking{}, nil, false, err
		}
		return persisted, retired, false, nil
	}
	created := make([]store.PersistentBooking, 0, len(planned))
	for _, item := range planned {
		reservation, guardFresh, err := createBookingTx(ctx, tx, item.Request)
		if err != nil {
			return store.PersistentBooking{}, nil, false, err
		}
		if !guardFresh {
			return store.PersistentBooking{}, nil, false, store.ErrBookingConflict
		}
		job := item.Job
		job.BookingRowID = &reservation.Revision
		if err := insertOperationalJobTx(ctx, tx, job, item.Delivery); err != nil {
			return store.PersistentBooking{}, nil, false, err
		}
		if err := insertGuardOperationalUsageTx(ctx, tx, job, persisted); err != nil {
			return store.PersistentBooking{}, nil, false, err
		}
		created = append(created, reservation)
	}
	if err := tx.Commit(ctx); err != nil {
		return store.PersistentBooking{}, nil, false, mapWriteError(err)
	}
	return persisted, append(retired, created...), true, nil
}

func (r *Repository) CreateScheduledOperation(ctx context.Context, schedule string, occurrence time.Time, conflictMode string, item store.OperationalReservation) (store.OperationalScheduleResult, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.OperationalScheduleResult{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	identity := schedule + ":" + occurrence.UTC().Format(time.RFC3339Nano) + ":" + fmt.Sprint(item.Request.ManifestVersion)
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", "operational-schedule:"+identity); err != nil {
		return store.OperationalScheduleResult{}, err
	}
	var existingState string
	err = tx.QueryRow(ctx, `SELECT state FROM public.operational_schedule_occurrences
		WHERE schedule_name=$1 AND occurrence_at_ns=$2 AND manifest_version=$3`, schedule, occurrence.UnixNano(), item.Request.ManifestVersion).Scan(&existingState)
	if err == nil {
		return store.OperationalScheduleResult{State: existingState, Created: false}, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return store.OperationalScheduleResult{}, err
	}
	if err := lockCreate(ctx, tx, item.Request.Booking.User, item.Request.Booking.Policy, item.Request.Resource); err != nil {
		return store.OperationalScheduleResult{}, err
	}
	if err := assertManifestVersion(ctx, tx, item.Request.ManifestVersion); err != nil {
		return store.OperationalScheduleResult{}, err
	}
	if !item.Request.Booking.When.End.After(item.Request.Now) {
		if _, err := tx.Exec(ctx, `INSERT INTO public.operational_schedule_occurrences
			(schedule_name,occurrence_at,occurrence_at_ns,manifest_version,state,detail,slot_name,resource_name,workflow_name)
			VALUES($1,$2,$3,$4,'missed','occurrence ended before scheduler recovery',$5,$6,$7)`, schedule, occurrence.UTC(), occurrence.UnixNano(), item.Request.ManifestVersion,
			item.Request.Booking.Slot, item.Request.Resource, item.Job.Workflow); err != nil {
			return store.OperationalScheduleResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return store.OperationalScheduleResult{}, err
		}
		return store.OperationalScheduleResult{State: "missed", Created: true}, nil
	}
	var overlap bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM public.bookings WHERE resource_name=$1
		AND resource_constrained AND collection='live' AND NOT superseded
		AND int8range(starts_ns,ends_ns,'[)') && int8range($2,$3,'[)'))`, item.Request.Resource,
		item.Request.Booking.When.Start.UnixNano(), item.Request.Booking.When.End.UnixNano()).Scan(&overlap); err != nil {
		return store.OperationalScheduleResult{}, err
	}
	if overlap {
		state := "conflict"
		if conflictMode == store.OperationalConflictSkip {
			state = "skipped"
		}
		if _, err := tx.Exec(ctx, `INSERT INTO public.operational_schedule_occurrences
			(schedule_name,occurrence_at,occurrence_at_ns,manifest_version,state,detail,slot_name,resource_name,workflow_name)
			VALUES($1,$2,$3,$4,$5,'resource already reserved',$6,$7,$8)`, schedule, occurrence.UTC(), occurrence.UnixNano(), item.Request.ManifestVersion, state,
			item.Request.Booking.Slot, item.Request.Resource, item.Job.Workflow); err != nil {
			return store.OperationalScheduleResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return store.OperationalScheduleResult{}, err
		}
		return store.OperationalScheduleResult{State: state, Created: true}, nil
	}
	reservation, fresh, err := createBookingTx(ctx, tx, item.Request)
	if err != nil {
		return store.OperationalScheduleResult{}, err
	}
	if !fresh {
		return store.OperationalScheduleResult{}, store.ErrBookingConflict
	}
	job := item.Job
	job.BookingRowID = &reservation.Revision
	if err := insertOperationalJobTx(ctx, tx, job, item.Delivery); err != nil {
		return store.OperationalScheduleResult{}, err
	}
	if item.Usage == nil || item.Usage.Phase != "scheduled" || item.Usage.PayerKind != "experiment_owner" || item.Usage.PayerID == "" {
		return store.OperationalScheduleResult{}, errors.New("scheduled operation has no experiment-owner cost attribution")
	}
	if err := insertAttributedOperationalUsageTx(ctx, tx, job, reservation, *item.Usage); err != nil {
		return store.OperationalScheduleResult{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO public.operational_schedule_occurrences
		(schedule_name,occurrence_at,occurrence_at_ns,manifest_version,state,booking_row_id,job_id,slot_name,resource_name,workflow_name)
		VALUES($1,$2,$3,$4,'planned',$5,$6,$7,$8,$9)`, schedule, occurrence.UTC(), occurrence.UnixNano(), item.Request.ManifestVersion,
		reservation.Revision, job.ID, item.Request.Booking.Slot, item.Request.Resource, item.Job.Workflow); err != nil {
		return store.OperationalScheduleResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.OperationalScheduleResult{}, err
	}
	return store.OperationalScheduleResult{State: "planned", Created: true}, nil
}

func (r *Repository) ListScheduleOccurrences(ctx context.Context, from, until time.Time, state string, limit int) ([]operations.ScheduleOccurrence, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := r.pool.Query(ctx, `SELECT o.schedule_name,o.occurrence_at,o.manifest_version,o.state,
		o.slot_name,o.resource_name,o.workflow_name,coalesce(b.name,''),coalesce(o.job_id,''),o.detail
		FROM public.operational_schedule_occurrences o LEFT JOIN public.bookings b ON b.row_id=o.booking_row_id
		WHERE o.occurrence_at >= $1 AND o.occurrence_at < $2 AND ($3='' OR o.state=$3)
		ORDER BY o.occurrence_at,o.schedule_name LIMIT $4`, from.UTC(), until.UTC(), state, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]operations.ScheduleOccurrence, 0)
	for rows.Next() {
		var item operations.ScheduleOccurrence
		if err := rows.Scan(&item.Schedule, &item.OccurrenceAt, &item.ManifestVersion, &item.State, &item.Slot,
			&item.Resource, &item.Workflow, &item.BookingName, &item.JobID, &item.Detail); err != nil {
			return nil, err
		}
		item.OccurrenceAt = item.OccurrenceAt.UTC()
		result = append(result, item)
	}
	return result, rows.Err()
}

func retireTriggeredOperationalReservationsTx(ctx context.Context, tx pgx.Tx, triggeringName string, at time.Time, actor string) ([]store.PersistentBooking, error) {
	rows, err := tx.Query(ctx, `SELECT b.row_id,b.name,j.job_id FROM public.operational_jobs j
		JOIN public.bookings b ON b.row_id=j.booking_row_id
		JOIN public.webhook_deliveries d ON d.job_id=j.job_id AND d.direction='book-to-runner'
		WHERE j.triggering_booking_name=$1 AND j.state IN ('scheduled','reserved')
		AND d.state='pending' AND b.collection='live' AND NOT b.superseded FOR UPDATE OF b,j,d`, triggeringName)
	if err != nil {
		return nil, err
	}
	type candidate struct {
		rowID              int64
		bookingName, jobID string
	}
	candidates := make([]candidate, 0)
	for rows.Next() {
		var item candidate
		if err := rows.Scan(&item.rowID, &item.bookingName, &item.jobID); err != nil {
			rows.Close()
			return nil, err
		}
		candidates = append(candidates, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	retired := make([]store.PersistentBooking, 0, len(candidates))
	for _, item := range candidates {
		if _, err := tx.Exec(ctx, "UPDATE public.bookings SET superseded=true,updated_at=clock_timestamp() WHERE row_id=$1", item.rowID); err != nil {
			return nil, err
		}
		if err := insertEvent(ctx, tx, item.rowID, item.bookingName, "superseded", at, "operational-cancel:"+actor); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, "UPDATE public.operational_jobs SET state='cancelled',updated_at=clock_timestamp() WHERE job_id=$1", item.jobID); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, "UPDATE public.webhook_deliveries SET state='cancelled',updated_at=clock_timestamp() WHERE job_id=$1 AND state='pending'", item.jobID); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `UPDATE public.operational_usage_ledger SET state='cancelled',actual_duration_ns=0,
			completed_at=$2,updated_at=$2 WHERE job_id=$1 AND state IN ('reserved','dispatched')`, item.jobID, at.UTC()); err != nil {
			return nil, err
		}
		persisted, err := scanBooking(tx.QueryRow(ctx, bookingSelect+" WHERE row_id=$1", item.rowID))
		if err != nil {
			return nil, err
		}
		retired = append(retired, persisted)
	}
	return retired, nil
}

func supersedeReclaimableOperationsTx(ctx context.Context, tx pgx.Tx, names []string, now time.Time) error {
	for _, name := range names {
		var rowID int64
		var jobID, state string
		err := tx.QueryRow(ctx, `SELECT b.row_id,j.job_id,j.state FROM public.bookings b
			JOIN public.operational_jobs j ON j.booking_row_id=b.row_id
			JOIN public.webhook_deliveries d ON d.job_id=j.job_id AND d.direction='book-to-runner'
			WHERE b.name=$1 AND b.policy_name='__operations_reclaimable__' AND b.collection='live' AND NOT b.superseded
			AND d.state='pending' FOR UPDATE OF b,j,d`, name).Scan(&rowID, &jobID, &state)
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrBookingConflict
		}
		if err != nil {
			return err
		}
		if state != "scheduled" && state != "reserved" {
			return store.ErrBookingConflict
		}
		if _, err := tx.Exec(ctx, "UPDATE public.bookings SET superseded=true,updated_at=clock_timestamp() WHERE row_id=$1", rowID); err != nil {
			return err
		}
		if err := insertEvent(ctx, tx, rowID, name, "superseded", now, "operational-replan"); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "UPDATE public.operational_jobs SET state='cancelled',updated_at=clock_timestamp() WHERE job_id=$1", jobID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "UPDATE public.webhook_deliveries SET state='cancelled',updated_at=clock_timestamp() WHERE job_id=$1 AND state='pending'", jobID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE public.operational_usage_ledger SET state='cancelled',actual_duration_ns=0,
			completed_at=$2,updated_at=$2 WHERE job_id=$1 AND state IN ('reserved','dispatched')`, jobID, now.UTC()); err != nil {
			return err
		}
	}
	return nil
}

func insertOperationalJobTx(ctx context.Context, tx pgx.Tx, job operations.Job, delivery operations.Delivery) error {
	_, err := tx.Exec(ctx, `INSERT INTO public.operational_jobs
		(job_id,resource_name,workflow_name,job_kind,state,due_at,due_at_ns,starts_at,starts_at_ns,ends_at,ends_at_ns,
		 booking_row_id,triggering_booking_name,manifest_version,plan_revision,idempotency_key,payload)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		job.ID, job.Resource, job.Workflow, job.Kind, job.State, job.DueAt.UTC(), job.DueAt.UnixNano(),
		job.StartsAt.UTC(), job.StartsAt.UnixNano(), job.EndsAt.UTC(), job.EndsAt.UnixNano(), job.BookingRowID,
		job.TriggeringBookingName, job.ManifestVersion, job.PlanRevision, job.IdempotencyKey, string(job.Payload))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO public.webhook_deliveries
		(delivery_id,job_id,direction,state,body,next_attempt_at,next_attempt_at_ns)
		VALUES($1,$2,'book-to-runner','pending',$3,$4,$5)`, delivery.ID, job.ID, string(delivery.Body), job.DueAt.UTC(), job.DueAt.UnixNano())
	return err
}

func insertGuardOperationalUsageTx(ctx context.Context, tx pgx.Tx, job operations.Job, triggering store.PersistentBooking) error {
	phase := "preparation"
	if job.Kind == "teardown" || job.Kind == "settling" {
		phase = "cleanup"
	}
	return insertAttributedOperationalUsageTx(ctx, tx, job, triggering, store.OperationalUsageAttribution{
		Phase: phase, PayerKind: "user", PayerID: triggering.Booking.User, Chargeable: true,
	})
}

func insertAttributedOperationalUsageTx(ctx context.Context, tx pgx.Tx, job operations.Job, triggering store.PersistentBooking, usage store.OperationalUsageAttribution) error {
	planned := job.EndsAt.Sub(job.StartsAt)
	if planned < 0 {
		return errors.New("operational job ends before it starts")
	}
	_, err := tx.Exec(ctx, `INSERT INTO public.operational_usage_ledger
		(job_id,triggering_booking_row_id,triggering_booking_name,user_name,phase,payer_kind,payer_id,chargeable,state,planned_duration_ns)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,'reserved',$9) ON CONFLICT (job_id) DO NOTHING`,
		job.ID, triggering.Revision, job.TriggeringBookingName, triggering.Booking.User, usage.Phase, usage.PayerKind, usage.PayerID, usage.Chargeable, planned.Nanoseconds())
	return err
}

func (r *Repository) GetJob(ctx context.Context, id string) (operations.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	job, err := scanJob(r.pool.QueryRow(ctx, jobSelect+" WHERE job_id=$1", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return operations.Job{}, operations.ErrNotFound
	}
	return job, err
}

func (r *Repository) ActivateOperationalJob(ctx context.Context, jobID, deliveryID, bodyHash string, at time.Time) (store.PersistentBooking, operations.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return store.PersistentBooking{}, operations.Job{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", "runner-delivery:"+deliveryID); err != nil {
		return store.PersistentBooking{}, operations.Job{}, err
	}
	var existingHash, existingJob string
	err = tx.QueryRow(ctx, "SELECT body_sha256,job_id FROM public.webhook_callback_receipts WHERE delivery_id=$1", deliveryID).Scan(&existingHash, &existingJob)
	if err == nil && (existingHash != bodyHash || existingJob != jobID) {
		return store.PersistentBooking{}, operations.Job{}, operations.ErrCallbackConflict
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return store.PersistentBooking{}, operations.Job{}, err
	}
	duplicate := err == nil
	var rowID int64
	var state string
	var activationRunID *string
	var maintenance bool
	err = tx.QueryRow(ctx, `SELECT j.booking_row_id,j.state,j.activation_run_id,b.maintenance FROM public.operational_jobs j
		JOIN public.bookings b ON b.row_id=j.booking_row_id
		WHERE j.job_id=$1 AND b.collection='live' AND NOT b.superseded FOR UPDATE OF j,b`, jobID).Scan(&rowID, &state, &activationRunID, &maintenance)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.PersistentBooking{}, operations.Job{}, operations.ErrNotFound
	}
	if err != nil {
		return store.PersistentBooking{}, operations.Job{}, err
	}
	if state != "accepted" && state != "running" {
		return store.PersistentBooking{}, operations.Job{}, operations.ErrInvalidTransition
	}
	if !duplicate {
		if _, err := tx.Exec(ctx, `INSERT INTO public.webhook_callback_receipts(delivery_id,job_id,body_sha256)
			VALUES($1,$2,$3)`, deliveryID, jobID, bodyHash); err != nil {
			return store.PersistentBooking{}, operations.Job{}, err
		}
	}
	var started bool
	var startsAt, endsAt time.Time
	var bookingName string
	if err := tx.QueryRow(ctx, `SELECT name,started,starts_at,ends_at FROM public.bookings WHERE row_id=$1 FOR UPDATE`, rowID).
		Scan(&bookingName, &started, &startsAt, &endsAt); err != nil {
		return store.PersistentBooking{}, operations.Job{}, err
	}
	if at.Before(startsAt) || !at.Before(endsAt) {
		return store.PersistentBooking{}, operations.Job{}, store.ErrBookingConflict
	}
	startBooking := activationRunID == nil || maintenance
	if !started && startBooking {
		if _, err := tx.Exec(ctx, `UPDATE public.bookings SET started=true,started_at=$2,started_at_ns=$3,
			started_at_text=$4,updated_at=clock_timestamp() WHERE row_id=$1`, rowID, at.UTC(), at.UnixNano(), at.UTC().Format(time.RFC3339Nano)); err != nil {
			return store.PersistentBooking{}, operations.Job{}, err
		}
		if err := insertEvent(ctx, tx, rowID, bookingName, "started", at, "operational-runner:"+jobID); err != nil {
			return store.PersistentBooking{}, operations.Job{}, err
		}
	}
	if state == "accepted" {
		if _, err := tx.Exec(ctx, "UPDATE public.operational_jobs SET state='running',updated_at=$2 WHERE job_id=$1", jobID, at.UTC()); err != nil {
			return store.PersistentBooking{}, operations.Job{}, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE public.operational_usage_ledger SET state='running',started_at=COALESCE(started_at,$2),updated_at=$2 WHERE job_id=$1`, jobID, at.UTC()); err != nil {
		return store.PersistentBooking{}, operations.Job{}, err
	}
	booking, err := scanBooking(tx.QueryRow(ctx, bookingSelect+" WHERE row_id=$1", rowID))
	if err != nil {
		return store.PersistentBooking{}, operations.Job{}, err
	}
	job, err := scanJob(tx.QueryRow(ctx, jobSelect+" WHERE job_id=$1", jobID))
	if err != nil {
		return store.PersistentBooking{}, operations.Job{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.PersistentBooking{}, operations.Job{}, err
	}
	return booking, job, nil
}

func (r *Repository) ClaimDeliveries(ctx context.Context, owner string, now time.Time, lease time.Duration, limit int) ([]operations.Delivery, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if lease <= 0 {
		lease = 30 * time.Second
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	rows, err := tx.Query(ctx, `WITH candidates AS (
		SELECT delivery_id FROM public.webhook_deliveries
		WHERE ((state='pending' AND next_attempt_at_ns <= $1) OR (state='leased' AND lease_until < $2))
		ORDER BY next_attempt_at_ns,delivery_id FOR UPDATE SKIP LOCKED LIMIT $3)
		UPDATE public.webhook_deliveries d SET state='leased',lease_owner=$4,lease_until=$5,attempts=d.attempts+1,updated_at=clock_timestamp()
		FROM candidates c WHERE d.delivery_id=c.delivery_id RETURNING d.delivery_id,d.job_id,d.body::text,d.attempts`,
		now.UnixNano(), now.UTC(), limit, owner, now.Add(lease).UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]operations.Delivery, 0)
	for rows.Next() {
		var item operations.Delivery
		var body string
		if err := rows.Scan(&item.ID, &item.JobID, &body, &item.Attempts); err != nil {
			return nil, err
		}
		item.Body = []byte(body)
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) CompleteDelivery(ctx context.Context, id, owner string, success bool, status int, failure string, now, nextAttempt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	state := "delivered"
	if !success {
		if nextAttempt.After(now) {
			state = "pending"
		} else {
			state = "dead"
		}
	}
	var jobID string
	err = tx.QueryRow(ctx, `UPDATE public.webhook_deliveries SET state=$3,response_status=$4,last_error=$5,next_attempt_at=$6,
		next_attempt_at_ns=$7,lease_owner='',lease_until=NULL,updated_at=clock_timestamp()
		WHERE delivery_id=$1 AND state='leased' AND lease_owner=$2 RETURNING job_id`, id, owner, state, status, failure, nextAttempt.UTC(), nextAttempt.UnixNano()).Scan(&jobID)
	if errors.Is(err, pgx.ErrNoRows) {
		return operations.ErrNotFound
	}
	if err != nil {
		return err
	}
	if success {
		_, err = tx.Exec(ctx, `UPDATE public.operational_jobs SET state='dispatched',attempts=attempts+1,updated_at=clock_timestamp() WHERE job_id=$1 AND state IN ('scheduled','reserved')`, jobID)
		if err == nil {
			_, err = tx.Exec(ctx, `UPDATE public.operational_usage_ledger SET state='dispatched',updated_at=$2 WHERE job_id=$1 AND state='reserved'`, jobID, now.UTC())
		}
	} else {
		_, err = tx.Exec(ctx, `UPDATE public.operational_jobs SET attempts=attempts+1,last_error=$2,updated_at=clock_timestamp() WHERE job_id=$1`, jobID, failure)
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func allowedTransition(from, to string) bool {
	if from == to {
		return true
	}
	switch to {
	case "accepted":
		return from == "scheduled" || from == "reserved" || from == "dispatched"
	case "running":
		return from == "accepted"
	case "succeeded", "failed":
		return from == "accepted" || from == "running"
	case "cancelled", "expired":
		return from != "succeeded" && from != "failed" && from != "cancelled" && from != "expired"
	}
	return false
}

func (r *Repository) ApplyCallback(ctx context.Context, callback operations.Callback, bodyHash string) (operations.Job, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return operations.Job{}, false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	var existingHash string
	err = tx.QueryRow(ctx, `SELECT body_sha256 FROM public.webhook_callback_receipts WHERE delivery_id=$1`, callback.DeliveryID).Scan(&existingHash)
	if err == nil {
		if existingHash != bodyHash {
			return operations.Job{}, false, operations.ErrCallbackConflict
		}
		job, err := scanJob(tx.QueryRow(ctx, jobSelect+" WHERE job_id=$1", callback.JobID))
		if err != nil {
			return operations.Job{}, false, err
		}
		return job, false, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return operations.Job{}, false, err
	}
	job, err := scanJob(tx.QueryRow(ctx, jobSelect+" WHERE job_id=$1 FOR UPDATE", callback.JobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return operations.Job{}, false, operations.ErrNotFound
	}
	if err != nil {
		return operations.Job{}, false, err
	}
	if !allowedTransition(job.State, callback.State) {
		return operations.Job{}, false, operations.ErrInvalidTransition
	}
	_, err = tx.Exec(ctx, `INSERT INTO public.webhook_callback_receipts(delivery_id,job_id,body_sha256) VALUES($1,$2,$3)`, callback.DeliveryID, callback.JobID, bodyHash)
	if err != nil {
		return operations.Job{}, false, err
	}
	_, err = tx.Exec(ctx, `UPDATE public.operational_jobs SET state=$2,last_error=$3,updated_at=$4 WHERE job_id=$1`, callback.JobID, callback.State, callback.Error, callback.At.UTC())
	if err != nil {
		return operations.Job{}, false, err
	}
	if err := updateOperationalUsageTx(ctx, tx, callback); err != nil {
		return operations.Job{}, false, err
	}
	var activationRunID *string
	var activationStageIndex *int
	if err := tx.QueryRow(ctx, `SELECT activation_run_id,activation_stage_index FROM public.operational_jobs WHERE job_id=$1`, callback.JobID).Scan(&activationRunID, &activationStageIndex); err != nil {
		return operations.Job{}, false, err
	}
	if activationRunID != nil && activationStageIndex != nil {
		if err := applyActivationCallbackTx(ctx, tx, job, callback, *activationRunID, *activationStageIndex); err != nil {
			return operations.Job{}, false, err
		}
	} else if callback.State == "succeeded" || callback.State == "failed" || callback.State == "cancelled" || callback.State == "expired" {
		if err := finishOperationalReservationTx(ctx, tx, job, callback); err != nil {
			return operations.Job{}, false, err
		}
	}
	updated, err := scanJob(tx.QueryRow(ctx, jobSelect+" WHERE job_id=$1", callback.JobID))
	if err != nil {
		return operations.Job{}, false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return operations.Job{}, false, err
	}
	return updated, true, nil
}

func applyActivationCallbackTx(ctx context.Context, tx pgx.Tx, job operations.Job, callback operations.Callback, runID string, stageIndex int) error {
	code := callback.Code
	if code == "" {
		code = callback.Error
	}
	var phase string
	if err := tx.QueryRow(ctx, `SELECT phase FROM public.booking_activation_stages WHERE run_id=$1 AND stage_index=$2`, runID, stageIndex).Scan(&phase); err != nil {
		return err
	}
	if callback.State == "accepted" || callback.State == "running" {
		_, err := tx.Exec(ctx, `UPDATE public.booking_activation_stages SET state=$3,updated_at=$4 WHERE run_id=$1 AND stage_index=$2`, runID, stageIndex, callback.State, callback.At.UTC())
		return err
	}
	if callback.State == "succeeded" {
		if _, err := tx.Exec(ctx, `UPDATE public.booking_activation_stages SET state='succeeded',completed_at=$3,updated_at=$3,last_error_code='',last_error='' WHERE run_id=$1 AND stage_index=$2`, runID, stageIndex, callback.At.UTC()); err != nil {
			return err
		}
		// A successful recheck completes the recovery cycle for this stream. Reset
		// the counter here (not when recovery schedules the recheck), so repeated
		// failures remain bounded while a later stream receives its own allowance.
		if _, err := tx.Exec(ctx, `UPDATE public.booking_activation_runs SET recovery_attempt=0,recovery_target_stage=NULL,updated_at=$3
			WHERE run_id=$1 AND recovery_target_stage=$2`, runID, stageIndex, callback.At.UTC()); err != nil {
			return err
		}
		var waitAfter int64
		var healthCheck bool
		if err := tx.QueryRow(ctx, `SELECT wait_after_ns,health_check FROM public.booking_activation_stages WHERE run_id=$1 AND stage_index=$2`, runID, stageIndex).Scan(&waitAfter, &healthCheck); err != nil {
			return err
		}
		if healthCheck {
			if err := recordActivationHealthSuccessTx(ctx, tx, runID, job.ID, callback.At.UTC()); err != nil {
				return err
			}
		}
		var nextStage *int
		nextQuery := `SELECT MIN(stage_index) FROM public.booking_activation_stages WHERE run_id=$1 AND stage_index>$2 AND phase=$3`
		if phase == "recovery" {
			nextQuery += ` AND stream_name=(SELECT stream_name FROM public.booking_activation_stages WHERE run_id=$1 AND stage_index=$2)`
		}
		if err := tx.QueryRow(ctx, nextQuery, runID, stageIndex, phase).Scan(&nextStage); err != nil {
			return err
		}
		if nextStage != nil {
			due := callback.At.UTC().Add(time.Duration(waitAfter))
			if err := createActivationStageJobTx(ctx, tx, runID, *nextStage, 1, due); err != nil {
				return err
			}
			var progress string
			if err := tx.QueryRow(ctx, `SELECT progress_message FROM public.booking_activation_stages WHERE run_id=$1 AND stage_index=$2`, runID, *nextStage).Scan(&progress); err != nil {
				return err
			}
			_, err := tx.Exec(ctx, `UPDATE public.booking_activation_runs SET current_stage=$2,progress_message=$3,updated_at=$4 WHERE run_id=$1 AND state='preparing'`, runID, *nextStage, progress, callback.At.UTC())
			return err
		}
		if phase == "cleanup" {
			if err := finishOperationalReservationTx(ctx, tx, job, callback); err != nil {
				return err
			}
			_, err := tx.Exec(ctx, `UPDATE public.booking_activation_runs SET cleanup_state='succeeded',
				state=CASE WHEN auto_close THEN 'closed' ELSE state END,updated_at=$2 WHERE run_id=$1`, runID, callback.At.UTC())
			return err
		}
		if phase == "recovery" {
			return resumeActivationAfterRecoveryTx(ctx, tx, runID, callback.At.UTC())
		}
		return activatePreparedBookingTx(ctx, tx, runID, callback.At.UTC(), job.ID)
	}
	if callback.State != "failed" && callback.State != "cancelled" && callback.State != "expired" {
		return nil
	}
	var attempt, maximum int
	var initialDelay, maximumDelay, totalTimeout int64
	var backoff float64
	var retryCodesJSON []byte
	var retryMessage string
	var createdAt time.Time
	err := tx.QueryRow(ctx, `SELECT attempt,maximum_attempts,retry_initial_delay_ns,retry_backoff,retry_maximum_delay_ns,retry_total_timeout_ns,
		retryable_codes,retry_message,created_at FROM public.booking_activation_stages WHERE run_id=$1 AND stage_index=$2 FOR UPDATE`, runID, stageIndex).
		Scan(&attempt, &maximum, &initialDelay, &backoff, &maximumDelay, &totalTimeout, &retryCodesJSON, &retryMessage, &createdAt)
	if err != nil {
		return err
	}
	var retryCodes []string
	if err := json.Unmarshal(retryCodesJSON, &retryCodes); err != nil {
		return err
	}
	retryable := callback.State == "failed" && len(retryCodes) == 0
	for _, retryCode := range retryCodes {
		if retryCode == code {
			retryable = true
		}
	}
	delay := time.Duration(float64(initialDelay) * math.Pow(backoff, float64(attempt-1)))
	if maximumDelay > 0 && delay > time.Duration(maximumDelay) {
		delay = time.Duration(maximumDelay)
	}
	due := callback.At.UTC().Add(delay)
	withinTotal := totalTimeout == 0 || !due.After(createdAt.Add(time.Duration(totalTimeout)))
	if retryable && attempt < maximum && withinTotal {
		if err := createActivationStageJobTx(ctx, tx, runID, stageIndex, attempt+1, due); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE public.booking_activation_runs SET progress_message=$2,updated_at=$3 WHERE run_id=$1`, runID, retryMessage, callback.At.UTC())
		return err
	}
	terminal := callback.State
	if terminal == "failed" {
		terminal = "failed"
	}
	_, err = tx.Exec(ctx, `UPDATE public.booking_activation_stages SET state=$3,last_error_code=$4,last_error=$5,completed_at=$6,updated_at=$6 WHERE run_id=$1 AND stage_index=$2`, runID, stageIndex, callback.State, code, callback.Error, callback.At.UTC())
	if err != nil {
		return err
	}
	if phase == "cleanup" {
		if err := finishOperationalReservationTx(ctx, tx, job, callback); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE public.booking_activation_runs r SET cleanup_state='failed',
			state=CASE WHEN r.auto_close THEN 'cleanup_failed' ELSE r.state END,updated_at=$2 FROM public.booking_activation_stages s
			WHERE r.run_id=$1 AND s.run_id=r.run_id AND s.stage_index=$3`, runID, callback.At.UTC(), stageIndex)
		return err
	}
	if phase == "recovery" {
		if _, err := tx.Exec(ctx, `UPDATE public.booking_activation_runs SET state='failed',failure_code=$2,failure_message=$3,
			completed_at=$4,updated_at=$4 WHERE run_id=$1`, runID, code, callback.Error, callback.At.UTC()); err != nil {
			return err
		}
		if err := recordActivationHealthFailureTx(ctx, tx, runID, job.ID, code, callback.Error, callback.At.UTC()); err != nil {
			return err
		}
		if err := finishAutoCloseActivationReservationTx(ctx, tx, runID, job, callback); err != nil {
			return err
		}
		return startActivationCleanupRunTx(ctx, tx, runID, callback.At.UTC())
	}
	var healthCheck bool
	if err := tx.QueryRow(ctx, `SELECT health_check FROM public.booking_activation_stages WHERE run_id=$1 AND stage_index=$2`, runID, stageIndex).Scan(&healthCheck); err != nil {
		return err
	}
	if callback.State == "failed" && healthCheck {
		started, err := startActivationRecoveryTx(ctx, tx, runID, stageIndex, callback.At.UTC())
		if err != nil {
			return err
		}
		if started {
			return nil
		}
	}
	_, err = tx.Exec(ctx, `UPDATE public.booking_activation_runs r SET state=$2,failure_code=$3,failure_message=$4,
		failure_guidance=s.failure_guidance,completed_at=$5,updated_at=$5 FROM public.booking_activation_stages s
		WHERE r.run_id=$1 AND s.run_id=r.run_id AND s.stage_index=$6`, runID, terminal, code, callback.Error, callback.At.UTC(), stageIndex)
	if err != nil {
		return err
	}
	if callback.State == "failed" {
		if err := recordActivationHealthFailureTx(ctx, tx, runID, job.ID, code, callback.Error, callback.At.UTC()); err != nil {
			return err
		}
	}
	if err := finishAutoCloseActivationReservationTx(ctx, tx, runID, job, callback); err != nil {
		return err
	}
	return startActivationCleanupRunTx(ctx, tx, runID, callback.At.UTC())
}

func finishAutoCloseActivationReservationTx(ctx context.Context, tx pgx.Tx, runID string, job operations.Job, callback operations.Callback) error {
	var autoClose bool
	if err := tx.QueryRow(ctx, `SELECT auto_close FROM public.booking_activation_runs WHERE run_id=$1`, runID).Scan(&autoClose); err != nil {
		return err
	}
	if !autoClose {
		return nil
	}
	return finishOperationalReservationTx(ctx, tx, job, callback)
}

func startActivationRecoveryTx(ctx context.Context, tx pgx.Tx, runID string, targetStage int, at time.Time) (bool, error) {
	var attempt, maximum int
	if err := tx.QueryRow(ctx, `SELECT recovery_attempt,maximum_recovery_attempts FROM public.booking_activation_runs WHERE run_id=$1 FOR UPDATE`, runID).Scan(&attempt, &maximum); err != nil {
		return false, err
	}
	var first *int
	err := tx.QueryRow(ctx, `SELECT MIN(stage_index) FROM public.booking_activation_stages WHERE run_id=$1 AND phase='recovery'
		AND stream_name=(SELECT stream_name FROM public.booking_activation_stages WHERE run_id=$1 AND stage_index=$2)`, runID, targetStage).Scan(&first)
	if errors.Is(err, pgx.ErrNoRows) || first == nil || maximum == 0 || attempt >= maximum {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err := tx.Exec(ctx, `UPDATE public.booking_activation_stages SET state='waiting',attempt=0,job_id=NULL,
		generation=generation+1,last_error_code='',last_error='',completed_at=NULL,updated_at=$3 WHERE run_id=$1 AND phase='recovery'
		AND stream_name=(SELECT stream_name FROM public.booking_activation_stages WHERE run_id=$1 AND stage_index=$2)`, runID, targetStage, at); err != nil {
		return false, err
	}
	if _, err := tx.Exec(ctx, `UPDATE public.booking_activation_runs SET recovery_attempt=recovery_attempt+1,recovery_target_stage=$2,
		current_stage=$3,progress_message='Attempting automatic recovery',updated_at=$4 WHERE run_id=$1`, runID, targetStage, *first, at); err != nil {
		return false, err
	}
	if err := createActivationStageJobTx(ctx, tx, runID, *first, 1, at); err != nil {
		return false, err
	}
	return true, nil
}

func resumeActivationAfterRecoveryTx(ctx context.Context, tx pgx.Tx, runID string, at time.Time) error {
	var target *int
	if err := tx.QueryRow(ctx, `SELECT recovery_target_stage FROM public.booking_activation_runs WHERE run_id=$1 FOR UPDATE`, runID).Scan(&target); err != nil {
		return err
	}
	if target == nil {
		return errors.New("activation recovery has no target health-check stage")
	}
	var progress string
	if err := tx.QueryRow(ctx, `UPDATE public.booking_activation_stages SET state='waiting',attempt=0,job_id=NULL,generation=generation+1,
		last_error_code='',last_error='',completed_at=NULL,updated_at=$3 WHERE run_id=$1 AND stage_index=$2 RETURNING progress_message`, runID, *target, at).Scan(&progress); err != nil {
		return err
	}
	if err := createActivationStageJobTx(ctx, tx, runID, *target, 1, at); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `UPDATE public.booking_activation_runs SET current_stage=$2,progress_message=$3,updated_at=$4 WHERE run_id=$1`, runID, *target, progress, at)
	return err
}

func recordActivationHealthFailureTx(ctx context.Context, tx pgx.Tx, runID, jobID, code, message string, at time.Time) error {
	if code == "" {
		code = "activation_failed"
	}
	var resource, stream string
	var manifestVersion int64
	if err := tx.QueryRow(ctx, `SELECT r.resource_name,COALESCE(NULLIF(s.stream_name,''),r.stream_name),r.manifest_version
		FROM public.booking_activation_runs r JOIN public.booking_activation_stages s ON s.run_id=r.run_id
		WHERE r.run_id=$1 AND s.job_id=$2`, runID, jobID).Scan(&resource, &stream, &manifestVersion); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO public.operational_stream_health
		(resource_name,stream_name,status,result_code,message,job_id,manifest_version,checked_at)
		VALUES($1,$2,'unhealthy',$3,$4,$5,$6,$7)
		ON CONFLICT (resource_name,stream_name) DO UPDATE SET status='unhealthy',result_code=EXCLUDED.result_code,
		message=EXCLUDED.message,job_id=EXCLUDED.job_id,manifest_version=EXCLUDED.manifest_version,
		checked_at=EXCLUDED.checked_at,updated_at=clock_timestamp()`, resource, stream, code, message, jobID, manifestVersion, at); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `INSERT INTO public.operational_alerts
		(resource_name,stream_name,result_code,message,job_id,manifest_version,status,first_seen_at,last_seen_at)
		VALUES($1,$2,$3,$4,$5,$6,'open',$7,$7)
		ON CONFLICT (resource_name,stream_name,result_code) WHERE status IN ('open','acknowledged')
		DO UPDATE SET occurrences=operational_alerts.occurrences+1,last_seen_at=EXCLUDED.last_seen_at,message=EXCLUDED.message,
		job_id=EXCLUDED.job_id,manifest_version=EXCLUDED.manifest_version,updated_at=clock_timestamp()`, resource, stream, code, message, jobID, manifestVersion, at)
	return err
}

func recordActivationHealthSuccessTx(ctx context.Context, tx pgx.Tx, runID, jobID string, at time.Time) error {
	var resource, stream string
	var manifestVersion int64
	if err := tx.QueryRow(ctx, `SELECT r.resource_name,COALESCE(NULLIF(s.stream_name,''),r.stream_name),r.manifest_version
		FROM public.booking_activation_runs r JOIN public.booking_activation_stages s ON s.run_id=r.run_id
		WHERE r.run_id=$1 AND s.job_id=$2`, runID, jobID).Scan(&resource, &stream, &manifestVersion); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO public.operational_stream_health
		(resource_name,stream_name,status,result_code,message,job_id,manifest_version,checked_at)
		VALUES($1,$2,'healthy','','',$3,$4,$5)
		ON CONFLICT (resource_name,stream_name) DO UPDATE SET status='healthy',result_code='',message='',job_id=EXCLUDED.job_id,
		manifest_version=EXCLUDED.manifest_version,checked_at=EXCLUDED.checked_at,updated_at=clock_timestamp()`, resource, stream, jobID, manifestVersion, at); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `UPDATE public.operational_alerts SET status='resolved',resolved_at=$3,resolved_by='successful_activation_check',
		updated_at=clock_timestamp() WHERE resource_name=$1 AND stream_name=$2 AND status IN ('open','acknowledged')`, resource, stream, at)
	if err != nil {
		return err
	}
	return releaseResourceIfAllChecksPassedTx(ctx, tx, resource, at)
}

func releaseResourceIfAllChecksPassedTx(ctx context.Context, tx pgx.Tx, resource string, at time.Time) error {
	var requiredJSON []byte
	var requestedAt time.Time
	var requestedBy string
	var manifestVersion int64
	err := tx.QueryRow(ctx, `SELECT required_streams,requested_at,requested_by,manifest_version FROM public.resource_release_state
		WHERE resource_name=$1 AND state='pending_checks' FOR UPDATE`, resource).Scan(&requiredJSON, &requestedAt, &requestedBy, &manifestVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	var required []string
	if err := json.Unmarshal(requiredJSON, &required); err != nil {
		return err
	}
	if len(required) == 0 {
		return errors.New("verified resource release requires at least one health-check stream")
	}
	var healthy int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM public.operational_stream_health
		WHERE resource_name=$1 AND stream_name=ANY($2) AND status='healthy' AND checked_at >= $3`, resource, required, requestedAt.UTC()).Scan(&healthy); err != nil {
		return err
	}
	if healthy != len(required) {
		return nil
	}
	if _, err := tx.Exec(ctx, `UPDATE public.resource_availability SET available=true,reason='Verified healthy after technician release request',
		updated_at=clock_timestamp(),updated_at_ns=$2,held_since=NULL,held_by='' WHERE resource_name=$1 AND NOT available`, resource, at.UnixNano()); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE public.resource_release_state SET state='verified',failing_streams='[]'::jsonb,released_at=$2,
		updated_at=clock_timestamp() WHERE resource_name=$1`, resource, at.UTC()); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO public.resource_release_events(resource_name,event_type,occurred_at,actor,reason,required_streams,failing_streams,manifest_version)
		VALUES($1,'verified_release',$2,$3,'All required health checks passed',$4,'[]'::jsonb,$5)`, resource, at.UTC(), requestedBy, string(requiredJSON), manifestVersion)
	return err
}

func createActivationStageJobTx(ctx context.Context, tx pgx.Tx, runID string, stageIndex, attempt int, due time.Time) error {
	var workflow, resource, bookingName, phase string
	var manifestVersion int64
	var generation int
	var parametersJSON []byte
	var timeoutAt, oldDue time.Time
	err := tx.QueryRow(ctx, `SELECT s.workflow_name,s.parameters,s.due_at,s.timeout_at,s.phase,r.resource_name,r.booking_name,r.manifest_version,s.generation
		FROM public.booking_activation_stages s JOIN public.booking_activation_runs r ON r.run_id=s.run_id
		WHERE s.run_id=$1 AND s.stage_index=$2 FOR UPDATE OF s,r`, runID, stageIndex).
		Scan(&workflow, &parametersJSON, &oldDue, &timeoutAt, &phase, &resource, &bookingName, &manifestVersion, &generation)
	if err != nil {
		return err
	}
	var parameters map[string]string
	if err := json.Unmarshal(parametersJSON, &parameters); err != nil {
		return err
	}
	timeout := timeoutAt.Sub(oldDue)
	jobID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("booking-activation-job\x00%s\x00%d\x00%d\x00%d", runID, stageIndex, generation, attempt))).String()
	deliveryID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("booking-activation-delivery\x00"+jobID)).String()
	key := fmt.Sprintf("booking-activation:%s:%d:%d:%d", runID, stageIndex, generation, attempt)
	jobKind := "preflight"
	if phase == "cleanup" {
		jobKind = "teardown"
	} else if phase == "recovery" {
		jobKind = "recovery"
	}
	command := operations.Command{Version: 1, JobID: jobID, Workflow: workflow, Resource: resource, Kind: jobKind, StartsAt: due, EndsAt: due.Add(timeout), BookingName: bookingName, PlanRevision: int64(attempt), IdempotencyKey: key, Parameters: parameters}
	body, err := json.Marshal(command)
	if err != nil {
		return err
	}
	var bookingRowID int64
	if err := tx.QueryRow(ctx, `SELECT booking_row_id FROM public.booking_activation_runs WHERE run_id=$1`, runID).Scan(&bookingRowID); err != nil {
		return err
	}
	if phase == "cleanup" {
		if attempt > 1 {
			_ = tx.QueryRow(ctx, `SELECT booking_row_id FROM public.operational_jobs WHERE activation_run_id=$1 AND activation_stage_index=$2 AND booking_row_id IS NOT NULL ORDER BY created_at LIMIT 1`, runID, stageIndex).Scan(&bookingRowID)
		} else {
			var slot, collection, runState string
			if err := tx.QueryRow(ctx, `SELECT b.slot_name,b.collection,r.state FROM public.booking_activation_runs r JOIN public.bookings b ON b.row_id=r.booking_row_id WHERE r.run_id=$1`, runID).Scan(&slot, &collection, &runState); err != nil {
				return err
			}
			useTriggeringBooking := collection == "live" && runState == "failed"
			if !useTriggeringBooking {
				if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", "resource:"+resource); err != nil {
					return err
				}
				var conflict bool
				if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM public.bookings WHERE collection='live' AND NOT superseded AND resource_constrained
				AND resource_name=$1 AND starts_ns < $3 AND ends_ns > $2)`, resource, due.UnixNano(), due.Add(timeout).UnixNano()).Scan(&conflict); err != nil {
					return err
				}
				if conflict {
					return operations.ErrActivationConflict
				}
				reservationName := uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("booking-cleanup-reservation\x00%s\x00%d", runID, stageIndex))).String()
				reservation, fresh, err := createBookingTx(ctx, tx, store.CreateBookingRequest{Booking: store.Booking{Name: reservationName, User: "__operations__", Policy: "__operations__:cleanup", Slot: slot,
					Maintenance: true, When: interval.Interval{Start: due.UTC(), End: due.Add(timeout).UTC()}}, Resource: resource, ResourceConstrained: true,
					Now: due.UTC(), ManifestVersion: manifestVersion, Maintenance: true, Actor: "activation-cleanup"})
				if err != nil {
					return err
				}
				if !fresh && !reservation.Booking.Maintenance {
					return operations.ErrActivationConflict
				}
				bookingRowID = reservation.Revision
			}
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO public.operational_jobs
		(job_id,resource_name,workflow_name,job_kind,state,due_at,due_at_ns,starts_at,starts_at_ns,ends_at,ends_at_ns,booking_row_id,
		triggering_booking_name,manifest_version,plan_revision,idempotency_key,payload,activation_run_id,activation_stage_index)
		VALUES($1,$2,$3,$4,'reserved',$5,$6,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`, jobID, resource, workflow, jobKind, due.UTC(), due.UnixNano(), due.Add(timeout).UTC(), due.Add(timeout).UnixNano(), bookingRowID, bookingName, manifestVersion, attempt, key, string(body), runID, stageIndex)
	if err != nil {
		return err
	}
	ledgerPhase := "preparation"
	if phase == "cleanup" {
		ledgerPhase = "cleanup"
	}
	if err := insertOperationalUsageTx(ctx, tx, jobID, runID, ledgerPhase, timeout); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO public.webhook_deliveries(delivery_id,job_id,direction,state,body,next_attempt_at,next_attempt_at_ns)
		VALUES($1,$2,'book-to-runner','pending',$3,$4,$5)`, deliveryID, jobID, string(body), due.UTC(), due.UnixNano()); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE public.booking_activation_stages SET state='pending',attempt=$3,due_at=$4,timeout_at=$5,job_id=$6,
		last_error_code='',last_error='',completed_at=NULL,updated_at=clock_timestamp() WHERE run_id=$1 AND stage_index=$2`, runID, stageIndex, attempt, due.UTC(), due.Add(timeout).UTC(), jobID)
	return err
}

func insertOperationalUsageTx(ctx context.Context, tx pgx.Tx, jobID, runID, phase string, planned time.Duration) error {
	_, err := tx.Exec(ctx, `INSERT INTO public.operational_usage_ledger
		(job_id,activation_run_id,triggering_booking_row_id,triggering_booking_name,user_name,phase,payer_kind,payer_id,chargeable,state,planned_duration_ns)
		SELECT $1,r.run_id,r.booking_row_id,r.booking_name,r.user_name,$3,'user',r.user_name,true,'reserved',$4 FROM public.booking_activation_runs r WHERE r.run_id=$2
		ON CONFLICT (job_id) DO NOTHING`, jobID, runID, phase, planned.Nanoseconds())
	return err
}

func updateOperationalUsageTx(ctx context.Context, tx pgx.Tx, callback operations.Callback) error {
	terminal := callback.State == "succeeded" || callback.State == "failed" || callback.State == "cancelled" || callback.State == "expired"
	if terminal {
		_, err := tx.Exec(ctx, `UPDATE public.operational_usage_ledger SET state=$2,completed_at=$3,
			actual_duration_ns=CASE WHEN started_at IS NULL THEN 0 ELSE LEAST(planned_duration_ns,GREATEST(0,EXTRACT(EPOCH FROM ($3-started_at))*1000000000)::bigint) END,
			updated_at=$3 WHERE job_id=$1`, callback.JobID, callback.State, callback.At.UTC())
		return err
	}
	_, err := tx.Exec(ctx, `UPDATE public.operational_usage_ledger SET state=$2,updated_at=$3 WHERE job_id=$1`, callback.JobID, callback.State, callback.At.UTC())
	return err
}

func activatePreparedBookingTx(ctx context.Context, tx pgx.Tx, runID string, at time.Time, jobID string) error {
	var rowID int64
	var bookingName, resource, slot string
	var started, maintenance bool
	var startsAt, endsAt time.Time
	if err := tx.QueryRow(ctx, `SELECT b.row_id,b.name,b.started,b.starts_at,b.ends_at,b.resource_name,b.slot_name,b.maintenance FROM public.booking_activation_runs r
		JOIN public.bookings b ON b.row_id=r.booking_row_id WHERE r.run_id=$1 FOR UPDATE OF b,r`, runID).Scan(&rowID, &bookingName, &started, &startsAt, &endsAt, &resource, &slot, &maintenance); err != nil {
		return err
	}
	if at.Before(startsAt) || !at.Before(endsAt) {
		_, err := tx.Exec(ctx, `UPDATE public.booking_activation_runs SET state='expired',failure_code='booking_ended',failure_message='The booking is no longer active',completed_at=$2,updated_at=$2 WHERE run_id=$1`, runID, at)
		return err
	}
	var autoClose bool
	if err := tx.QueryRow(ctx, `SELECT auto_close FROM public.booking_activation_runs WHERE run_id=$1`, runID).Scan(&autoClose); err != nil {
		return err
	}
	if err := assertAvailable(ctx, tx, resource, slot); err != nil && !maintenance {
		if _, updateErr := tx.Exec(ctx, `UPDATE public.booking_activation_runs SET state='failed',failure_code='technician_suspended',
			failure_message='This experiment is temporarily offline for maintenance. Please choose another experiment or contact the laboratory.',
			completed_at=$2,updated_at=$2 WHERE run_id=$1`, runID, at); updateErr != nil {
			return updateErr
		}
		return startActivationCleanupRunTx(ctx, tx, runID, at)
	}
	if !started {
		if _, err := tx.Exec(ctx, `UPDATE public.bookings SET started=true,started_at=$2,started_at_ns=$3,started_at_text=$4,updated_at=clock_timestamp() WHERE row_id=$1`, rowID, at, at.UnixNano(), at.Format(time.RFC3339Nano)); err != nil {
			return err
		}
		if err := insertEvent(ctx, tx, rowID, bookingName, "started", at, "activation-runner:"+jobID); err != nil {
			return err
		}
	}
	if autoClose {
		job, err := scanJob(tx.QueryRow(ctx, jobSelect+" WHERE job_id=$1", jobID))
		if err != nil {
			return err
		}
		callback := operations.Callback{JobID: jobID, State: "succeeded", At: at}
		if err := finishOperationalReservationTx(ctx, tx, job, callback); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE public.booking_activation_runs SET state='cleaning',completed_at=$2,updated_at=$2 WHERE run_id=$1`, runID, at); err != nil {
			return err
		}
		if err := startActivationCleanupRunTx(ctx, tx, runID, at); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE public.booking_activation_runs SET state='closed' WHERE run_id=$1 AND cleanup_state='not_required'`, runID)
		return err
	}
	_, err := tx.Exec(ctx, `UPDATE public.booking_activation_runs SET state='active',progress_message='Ready',completed_at=$2,updated_at=$2 WHERE run_id=$1`, runID, at)
	return err
}

func finishOperationalReservationTx(ctx context.Context, tx pgx.Tx, job operations.Job, callback operations.Callback) error {
	if job.BookingRowID == nil {
		return nil
	}
	var bookingName string
	var started bool
	var startedAt *time.Time
	var endsAt time.Time
	err := tx.QueryRow(ctx, `SELECT name,started,started_at,ends_at FROM public.bookings
		WHERE row_id=$1 AND collection='live' AND NOT superseded FOR UPDATE`, *job.BookingRowID).
		Scan(&bookingName, &started, &startedAt, &endsAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	usage := time.Duration(0)
	if started && startedAt != nil {
		effectiveEnd := callback.At.UTC()
		if effectiveEnd.After(endsAt) {
			effectiveEnd = endsAt
		}
		if effectiveEnd.After(*startedAt) {
			usage = effectiveEnd.Sub(*startedAt)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO public.relay_revocations
			(booking_row_id,booking_name,expires_at,expires_at_ns,revoked_by)
			VALUES($1,$2,$3,$4,$5) ON CONFLICT (booking_row_id) DO NOTHING`,
			*job.BookingRowID, bookingName, endsAt.UTC(), endsAt.UnixNano(), "operational-runner:"+callback.State); err != nil {
			return err
		}
	}
	if err := supersedeHistoryName(ctx, tx, bookingName, callback.At, "operational-runner:"+callback.State); err != nil {
		return err
	}
	eventType := "completed"
	cancelled := false
	unfulfilled := false
	if callback.State != "succeeded" {
		cancelled = true
		eventType = "cancelled"
		unfulfilled = !started
		if callback.State == "expired" {
			eventType = "expired"
		}
	}
	if cancelled {
		_, err = tx.Exec(ctx, `UPDATE public.bookings SET collection='history',cancelled=true,cancelled_at=$2,
			cancelled_at_ns=$3,cancelled_by=$4,cancelled_by_text=$4,unfulfilled=$5,usage_charge_ns=$6,
			updated_at=clock_timestamp() WHERE row_id=$1`, *job.BookingRowID, callback.At.UTC(), callback.At.UnixNano(),
			"operational-runner:"+callback.State, unfulfilled, usage.Nanoseconds())
	} else {
		_, err = tx.Exec(ctx, `UPDATE public.bookings SET collection='history',usage_charge_ns=$2,
			updated_at=clock_timestamp() WHERE row_id=$1`, *job.BookingRowID, usage.Nanoseconds())
	}
	if err != nil {
		return err
	}
	return insertEvent(ctx, tx, *job.BookingRowID, bookingName, eventType, callback.At, "operational-runner:"+callback.State)
}

var _ operations.Repository = (*Repository)(nil)
