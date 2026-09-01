package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/practable/book/internal/operations"
	"github.com/practable/book/internal/store"
)

func (r *Repository) CreateActivation(ctx context.Context, request operations.CreateActivationRequest) (operations.ActivationRun, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return operations.ActivationRun{}, false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	run, fresh, err := createActivationTx(ctx, tx, request)
	if err != nil {
		return operations.ActivationRun{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return operations.ActivationRun{}, false, err
	}
	return run, fresh, nil
}

func createActivationTx(ctx context.Context, tx pgx.Tx, request operations.CreateActivationRequest) (operations.ActivationRun, bool, error) {
	var err error
	if len(request.Stages) == 0 || request.RunID == "" || request.IdempotencyKey == "" || request.FirstJob.ID == "" || request.FirstDelivery.ID == "" || request.FirstDelivery.JobID != request.FirstJob.ID || request.RecoveryAttempts < 0 || request.RecoveryAttempts > 5 || (len(request.RecoveryStages) == 0) != (request.RecoveryAttempts == 0) {
		return operations.ActivationRun{}, false, errors.New("activation requires at least one stage")
	}
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock_shared($1)", maintenanceLock); err != nil {
		return operations.ActivationRun{}, false, err
	}
	if err := assertManifestVersion(ctx, tx, request.ManifestVersion); err != nil {
		return operations.ActivationRun{}, false, err
	}
	var rowID int64
	var resource string
	var started bool
	var startsAt, endsAt time.Time
	err = tx.QueryRow(ctx, `SELECT row_id,resource_name,started,starts_at,ends_at FROM public.bookings WHERE name=$1 AND user_name=$2 AND collection='live' AND NOT superseded FOR UPDATE`, request.BookingName, request.User).Scan(&rowID, &resource, &started, &startsAt, &endsAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return operations.ActivationRun{}, false, store.ErrPersistentNotFound
	}
	if err != nil {
		return operations.ActivationRun{}, false, err
	}
	requestedAt := request.RequestedAt.UTC()
	if requestedAt.IsZero() || resource != request.Resource || started || requestedAt.Before(startsAt) || !requestedAt.Before(endsAt) {
		return operations.ActivationRun{}, false, store.ErrBookingConflict
	}

	var existingID string
	err = tx.QueryRow(ctx, `SELECT run_id FROM public.booking_activation_runs WHERE booking_row_id=$1 AND idempotency_key=$2`, rowID, request.IdempotencyKey).Scan(&existingID)
	if err == nil {
		run, err := getActivationTx(ctx, tx, existingID)
		if err != nil {
			return operations.ActivationRun{}, false, err
		}
		return run, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return operations.ActivationRun{}, false, err
	}
	var open bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM public.booking_activation_runs WHERE booking_row_id=$1 AND state='preparing')`, rowID).Scan(&open); err != nil {
		return operations.ActivationRun{}, false, err
	}
	if open {
		return operations.ActivationRun{}, false, operations.ErrActivationConflict
	}

	progress := request.Stages[0].ProgressMessage
	cleanupState := "pending"
	if len(request.CleanupStages) == 0 {
		cleanupState = "not_required"
	}
	_, err = tx.Exec(ctx, `INSERT INTO public.booking_activation_runs
		(run_id,booking_row_id,booking_name,user_name,resource_name,stream_name,pipeline_name,manifest_version,idempotency_key,state,current_stage,resolved_plan,progress_message,cleanup_state,maximum_recovery_attempts,auto_close)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,'preparing',0,$10,$11,$12,$13,$14)`, request.RunID, rowID, request.BookingName, request.User,
		request.Resource, request.Stream, request.Pipeline, request.ManifestVersion, request.IdempotencyKey, string(request.ResolvedPlan), progress, cleanupState, request.RecoveryAttempts, request.AutoClose)
	if err != nil {
		return operations.ActivationRun{}, false, err
	}
	stageIndex := 0
	for _, phase := range []struct {
		name   string
		stages []operations.ActivationStageSpec
	}{{"activation", request.Stages}, {"recovery", request.RecoveryStages}, {"cleanup", request.CleanupStages}} {
		for phaseIndex, stage := range phase.stages {
			state, attempt := "waiting", 1
			if phase.name == "activation" && phaseIndex == 0 {
				state = "pending"
			}
			if phase.name != "activation" {
				attempt = 0
			}
			parameters, err := json.Marshal(stage.Parameters)
			if err != nil {
				return operations.ActivationRun{}, false, err
			}
			retryCodes, err := json.Marshal(stage.RetryableCodes)
			if err != nil {
				return operations.ActivationRun{}, false, err
			}
			guidance := interface{}(nil)
			if len(stage.FailureGuidance) > 0 {
				guidance = string(stage.FailureGuidance)
			}
			_, err = tx.Exec(ctx, `INSERT INTO public.booking_activation_stages
			(run_id,stage_index,stage_name,job_template_name,workflow_name,state,attempt,maximum_attempts,due_at,timeout_at,parameters,
			 progress_message,retry_message,wait_after_ns,retry_initial_delay_ns,retry_backoff,retry_maximum_delay_ns,retry_total_timeout_ns,retryable_codes,failure_guidance,phase,health_check)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`, request.RunID, stageIndex, stage.Name, stage.JobTemplate, stage.Workflow, state, attempt,
				stage.MaximumAttempts, stage.DueAt.UTC(), stage.TimeoutAt.UTC(), string(parameters), stage.ProgressMessage, stage.RetryMessage,
				stage.WaitAfter.Nanoseconds(), stage.InitialDelay.Nanoseconds(), normalizedBackoff(stage.Backoff), stage.MaximumDelay.Nanoseconds(),
				stage.TotalTimeout.Nanoseconds(), string(retryCodes), guidance, phase.name, stage.Kind == "health_check")
			if err != nil {
				return operations.ActivationRun{}, false, err
			}
			stageIndex++
		}
	}
	job := request.FirstJob
	job.BookingRowID = &rowID
	_, err = tx.Exec(ctx, `INSERT INTO public.operational_jobs
		(job_id,resource_name,workflow_name,job_kind,state,due_at,due_at_ns,starts_at,starts_at_ns,ends_at,ends_at_ns,booking_row_id,
		 triggering_booking_name,manifest_version,plan_revision,idempotency_key,payload,activation_run_id,activation_stage_index)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,0)`, job.ID, job.Resource, job.Workflow, job.Kind, job.State,
		job.DueAt.UTC(), job.DueAt.UnixNano(), job.StartsAt.UTC(), job.StartsAt.UnixNano(), job.EndsAt.UTC(), job.EndsAt.UnixNano(), rowID,
		job.TriggeringBookingName, job.ManifestVersion, job.PlanRevision, job.IdempotencyKey, string(job.Payload), request.RunID)
	if err != nil {
		return operations.ActivationRun{}, false, err
	}
	if err := insertOperationalUsageTx(ctx, tx, job.ID, request.RunID, "preparation", job.EndsAt.Sub(job.StartsAt)); err != nil {
		return operations.ActivationRun{}, false, err
	}
	if _, err = tx.Exec(ctx, `UPDATE public.booking_activation_stages SET job_id=$3 WHERE run_id=$1 AND stage_index=$2`, request.RunID, 0, job.ID); err != nil {
		return operations.ActivationRun{}, false, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO public.webhook_deliveries
		(delivery_id,job_id,direction,state,body,next_attempt_at,next_attempt_at_ns) VALUES($1,$2,'book-to-runner','pending',$3,$4,$5)`,
		request.FirstDelivery.ID, job.ID, string(request.FirstDelivery.Body), job.DueAt.UTC(), job.DueAt.UnixNano()); err != nil {
		return operations.ActivationRun{}, false, err
	}
	run, err := getActivationTx(ctx, tx, request.RunID)
	if err != nil {
		return operations.ActivationRun{}, false, err
	}
	return run, true, nil
}

// CreateHealthCheck atomically reserves the physical resource and creates the
// immutable activation plan and first runner delivery. A process failure can
// therefore leave neither an orphan reservation nor an unreserved check.
func (r *Repository) CreateHealthCheck(ctx context.Context, bookingRequest store.CreateBookingRequest, activationRequest operations.CreateActivationRequest) (store.PersistentBooking, operations.ActivationRun, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return store.PersistentBooking{}, operations.ActivationRun{}, false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	if existing, err := getActivationTx(ctx, tx, activationRequest.RunID); err == nil {
		booking, err := scanBooking(tx.QueryRow(ctx, bookingSelect+" WHERE name=$1 AND NOT superseded", existing.BookingName))
		if err != nil {
			return store.PersistentBooking{}, operations.ActivationRun{}, false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return store.PersistentBooking{}, operations.ActivationRun{}, false, err
		}
		return booking, existing, false, nil
	} else if !errors.Is(err, operations.ErrNotFound) {
		return store.PersistentBooking{}, operations.ActivationRun{}, false, err
	}
	booking, created, err := createBookingTx(ctx, tx, bookingRequest)
	if err != nil {
		return store.PersistentBooking{}, operations.ActivationRun{}, false, err
	}
	if !created {
		return store.PersistentBooking{}, operations.ActivationRun{}, false, store.ErrBookingIDConflict
	}
	activationRequest.BookingName = booking.Booking.Name
	activationRequest.User = booking.Booking.User
	run, _, err := createActivationTx(ctx, tx, activationRequest)
	if err != nil {
		return store.PersistentBooking{}, operations.ActivationRun{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.PersistentBooking{}, operations.ActivationRun{}, false, err
	}
	return booking, run, true, nil
}

func normalizedBackoff(value float64) float64 {
	if value < 1 {
		return 1
	}
	return value
}

func (r *Repository) GetActivation(ctx context.Context, id string) (operations.ActivationRun, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	return getActivation(ctx, r.pool, id)
}

func (r *Repository) GetLatestActivationForBooking(ctx context.Context, bookingName string) (operations.ActivationRun, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	var id string
	err := r.pool.QueryRow(ctx, `SELECT r.run_id FROM public.booking_activation_runs r JOIN public.bookings b ON b.row_id=r.booking_row_id
		WHERE b.name=$1 AND b.collection='live' AND NOT b.superseded ORDER BY r.started_at DESC,r.run_id DESC LIMIT 1`, bookingName).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return operations.ActivationRun{}, operations.ErrNotFound
	}
	if err != nil {
		return operations.ActivationRun{}, err
	}
	return getActivation(ctx, r.pool, id)
}

// SweepActivationTimeouts claims expired stage jobs with SKIP LOCKED. All book
// instances may run this concurrently; each timeout transition is processed by
// at most one transaction and uses the same retry/failure path as a callback.
func (r *Repository) SweepActivationTimeouts(ctx context.Context, now time.Time, limit int) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.operationTimeout)
	defer cancel()
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	rows, err := tx.Query(ctx, `SELECT j.job_id,j.activation_run_id,j.activation_stage_index FROM public.operational_jobs j
		JOIN public.booking_activation_stages s ON s.run_id=j.activation_run_id AND s.stage_index=j.activation_stage_index
		JOIN public.booking_activation_runs r ON r.run_id=s.run_id
		WHERE ((s.phase IN ('activation','recovery') AND r.state='preparing') OR (s.phase='cleanup' AND r.cleanup_state='running'))
		AND s.state IN ('pending','dispatched','accepted','running') AND s.timeout_at <= $1
		AND j.state IN ('scheduled','reserved','dispatched','accepted','running')
		ORDER BY s.timeout_at,j.job_id FOR UPDATE OF j SKIP LOCKED LIMIT $2`, now.UTC(), limit)
	if err != nil {
		return 0, err
	}
	type expired struct {
		jobID, runID string
		stage        int
	}
	var items []expired
	for rows.Next() {
		var item expired
		if err := rows.Scan(&item.jobID, &item.runID, &item.stage); err != nil {
			rows.Close()
			return 0, err
		}
		items = append(items, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	for _, item := range items {
		job, err := scanJob(tx.QueryRow(ctx, jobSelect+" WHERE job_id=$1", item.jobID))
		if err != nil {
			return 0, err
		}
		if _, err := tx.Exec(ctx, `UPDATE public.operational_jobs SET state='failed',last_error='runner callback deadline elapsed',updated_at=$2 WHERE job_id=$1`, item.jobID, now.UTC()); err != nil {
			return 0, err
		}
		if _, err := tx.Exec(ctx, `UPDATE public.webhook_deliveries SET state='cancelled',last_error='activation stage timed out',updated_at=$2
			WHERE job_id=$1 AND state IN ('pending','leased')`, item.jobID, now.UTC()); err != nil {
			return 0, err
		}
		callback := operations.Callback{JobID: item.jobID, State: "failed", At: now.UTC(), Code: "activation_timeout", Error: "The preparation service did not report completion in time"}
		if err := applyActivationCallbackTx(ctx, tx, job, callback, item.runID, item.stage); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(items), nil
}

type activationQueryer interface {
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
}

func getActivation(ctx context.Context, queryer activationQueryer, id string) (operations.ActivationRun, error) {
	var run operations.ActivationRun
	var guidance []byte
	err := queryer.QueryRow(ctx, `SELECT run_id,booking_name,user_name,resource_name,stream_name,pipeline_name,manifest_version,idempotency_key,state,cleanup_state,
		current_stage,recovery_attempt,maximum_recovery_attempts,progress_message,failure_code,failure_message,COALESCE(failure_guidance,'null'::jsonb),started_at,updated_at,completed_at
		FROM public.booking_activation_runs WHERE run_id=$1`, id).Scan(&run.ID, &run.BookingName, &run.User, &run.Resource, &run.Stream, &run.Pipeline,
		&run.ManifestVersion, &run.IdempotencyKey, &run.State, &run.CleanupState, &run.CurrentStage, &run.RecoveryAttempt, &run.MaximumRecoveryAttempts, &run.ProgressMessage, &run.FailureCode, &run.FailureMessage,
		&guidance, &run.StartedAt, &run.UpdatedAt, &run.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return run, operations.ErrNotFound
	}
	if err != nil {
		return run, err
	}
	run.FailureGuidance = guidance
	rows, err := queryer.Query(ctx, `SELECT stage_index,phase,stage_name,job_template_name,workflow_name,state,attempt,maximum_attempts,due_at,timeout_at,
		parameters,progress_message,last_error_code,last_error,COALESCE(job_id,''),retry_message,wait_after_ns,retry_initial_delay_ns,retry_backoff,
		retry_maximum_delay_ns,retry_total_timeout_ns,retryable_codes,COALESCE(failure_guidance,'null'::jsonb)
		FROM public.booking_activation_stages WHERE run_id=$1 ORDER BY stage_index`, id)
	if err != nil {
		return run, err
	}
	defer rows.Close()
	for rows.Next() {
		var stage operations.ActivationStage
		var parameters, retryCodes, guidance []byte
		var waitAfter, initialDelay, maximumDelay, totalTimeout int64
		if err := rows.Scan(&stage.Index, &stage.Phase, &stage.Name, &stage.JobTemplate, &stage.Workflow, &stage.State, &stage.Attempt, &stage.MaximumAttempts,
			&stage.DueAt, &stage.TimeoutAt, &parameters, &stage.ProgressMessage, &stage.LastErrorCode, &stage.LastError, &stage.JobID,
			&stage.RetryMessage, &waitAfter, &initialDelay, &stage.Backoff, &maximumDelay, &totalTimeout, &retryCodes, &guidance); err != nil {
			return run, err
		}
		if err := json.Unmarshal(parameters, &stage.Parameters); err != nil {
			return run, err
		}
		stage.DueAt, stage.TimeoutAt = stage.DueAt.UTC(), stage.TimeoutAt.UTC()
		stage.WaitAfter, stage.InitialDelay, stage.MaximumDelay, stage.TotalTimeout = time.Duration(waitAfter), time.Duration(initialDelay), time.Duration(maximumDelay), time.Duration(totalTimeout)
		if err := json.Unmarshal(retryCodes, &stage.RetryableCodes); err != nil {
			return run, err
		}
		stage.FailureGuidance = guidance
		if stage.Phase == "cleanup" {
			run.CleanupStages = append(run.CleanupStages, stage)
		} else if stage.Phase == "recovery" {
			run.RecoveryStages = append(run.RecoveryStages, stage)
		} else {
			run.Stages = append(run.Stages, stage)
		}
	}
	return run, rows.Err()
}

func getActivationTx(ctx context.Context, tx pgx.Tx, id string) (operations.ActivationRun, error) {
	return getActivation(ctx, tx, id)
}
