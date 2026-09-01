ALTER TABLE booking_activation_runs DROP CONSTRAINT booking_activation_runs_recovery_target;
ALTER TABLE operational_jobs DROP CONSTRAINT operational_jobs_job_kind_check;
ALTER TABLE operational_jobs ADD CONSTRAINT operational_jobs_job_kind_check
    CHECK (job_kind IN ('setup','teardown','settling','scheduled','health','preflight'));
ALTER TABLE booking_activation_runs
    DROP COLUMN recovery_target_stage,
    DROP COLUMN maximum_recovery_attempts,
    DROP COLUMN recovery_attempt;

ALTER TABLE booking_activation_stages
    DROP COLUMN generation,
    DROP COLUMN health_check,
    DROP CONSTRAINT booking_activation_stages_phase_check;
ALTER TABLE booking_activation_stages
    ADD CONSTRAINT booking_activation_stages_phase_check
    CHECK (phase IN ('activation','cleanup'));
