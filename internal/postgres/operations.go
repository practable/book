package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
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
		created = append(created, reservation)
	}
	if err := tx.Commit(ctx); err != nil {
		return store.PersistentBooking{}, nil, false, mapWriteError(err)
	}
	return persisted, append(retired, created...), true, nil
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
	err = tx.QueryRow(ctx, `SELECT j.booking_row_id,j.state FROM public.operational_jobs j
		JOIN public.bookings b ON b.row_id=j.booking_row_id
		WHERE j.job_id=$1 AND b.collection='live' AND NOT b.superseded FOR UPDATE OF j,b`, jobID).Scan(&rowID, &state)
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
	if !started {
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
	if callback.State == "succeeded" || callback.State == "failed" || callback.State == "cancelled" || callback.State == "expired" {
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
