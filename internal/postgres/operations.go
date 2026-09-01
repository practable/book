package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/practable/book/internal/operations"
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

func (r *Repository) GetJob(ctx context.Context, id string) (operations.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	job, err := scanJob(r.pool.QueryRow(ctx, jobSelect+" WHERE job_id=$1", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return operations.Job{}, operations.ErrNotFound
	}
	return job, err
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
	updated, err := scanJob(tx.QueryRow(ctx, jobSelect+" WHERE job_id=$1", callback.JobID))
	if err != nil {
		return operations.Job{}, false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return operations.Job{}, false, err
	}
	return updated, true, nil
}

var _ operations.Repository = (*Repository)(nil)
