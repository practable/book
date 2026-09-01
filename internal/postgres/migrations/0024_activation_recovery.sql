ALTER TABLE booking_activation_stages
    DROP CONSTRAINT booking_activation_stages_phase_check;
ALTER TABLE booking_activation_stages
    ADD CONSTRAINT booking_activation_stages_phase_check
    CHECK (phase IN ('activation','recovery','cleanup'));
ALTER TABLE booking_activation_stages
    ADD COLUMN health_check boolean NOT NULL DEFAULT false,
    ADD COLUMN generation integer NOT NULL DEFAULT 0 CHECK (generation >= 0);

ALTER TABLE operational_jobs DROP CONSTRAINT operational_jobs_job_kind_check;
ALTER TABLE operational_jobs ADD CONSTRAINT operational_jobs_job_kind_check
    CHECK (job_kind IN ('setup','teardown','settling','scheduled','health','preflight','recovery'));

ALTER TABLE booking_activation_runs
    ADD COLUMN recovery_attempt integer NOT NULL DEFAULT 0 CHECK (recovery_attempt >= 0),
    ADD COLUMN maximum_recovery_attempts integer NOT NULL DEFAULT 0 CHECK (maximum_recovery_attempts BETWEEN 0 AND 5),
    ADD COLUMN recovery_target_stage integer;

ALTER TABLE booking_activation_runs
    ADD CONSTRAINT booking_activation_runs_recovery_target
    FOREIGN KEY (run_id, recovery_target_stage)
    REFERENCES booking_activation_stages(run_id, stage_index);
