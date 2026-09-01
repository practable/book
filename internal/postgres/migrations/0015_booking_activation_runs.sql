CREATE TABLE booking_activation_runs (
    run_id text PRIMARY KEY,
    booking_row_id bigint NOT NULL REFERENCES bookings(row_id),
    booking_name text NOT NULL,
    user_name text NOT NULL,
    resource_name text NOT NULL,
    stream_name text NOT NULL,
    pipeline_name text NOT NULL,
    manifest_version bigint NOT NULL,
    idempotency_key text NOT NULL,
    state text NOT NULL CHECK (state IN ('preparing','active','failed','cancelled','expired')),
    current_stage integer NOT NULL DEFAULT 0 CHECK (current_stage >= 0),
    resolved_plan jsonb NOT NULL,
    progress_message text NOT NULL DEFAULT '',
    failure_code text NOT NULL DEFAULT '',
    failure_message text NOT NULL DEFAULT '',
    failure_guidance jsonb,
    started_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    completed_at timestamptz,
    UNIQUE (booking_row_id, idempotency_key)
);

CREATE UNIQUE INDEX booking_activation_runs_one_open
    ON booking_activation_runs (booking_row_id)
    WHERE state = 'preparing';
CREATE INDEX booking_activation_runs_booking
    ON booking_activation_runs (booking_name, started_at DESC);

CREATE TABLE booking_activation_stages (
    run_id text NOT NULL REFERENCES booking_activation_runs(run_id) ON DELETE CASCADE,
    stage_index integer NOT NULL CHECK (stage_index >= 0),
    stage_name text NOT NULL,
    job_template_name text NOT NULL,
    workflow_name text NOT NULL,
    state text NOT NULL CHECK (state IN ('waiting','pending','dispatched','accepted','running','succeeded','failed','cancelled','expired')),
    attempt integer NOT NULL DEFAULT 0 CHECK (attempt >= 0),
    maximum_attempts integer NOT NULL CHECK (maximum_attempts >= 1),
    due_at timestamptz NOT NULL,
    timeout_at timestamptz NOT NULL,
    parameters jsonb NOT NULL DEFAULT '{}'::jsonb,
    progress_message text NOT NULL DEFAULT '',
    last_error_code text NOT NULL DEFAULT '',
    last_error text NOT NULL DEFAULT '',
    job_id text REFERENCES operational_jobs(job_id),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    completed_at timestamptz,
    PRIMARY KEY (run_id, stage_index),
    UNIQUE (job_id),
    CHECK (timeout_at >= due_at)
);

CREATE INDEX booking_activation_stages_due
    ON booking_activation_stages (state, due_at);

ALTER TABLE operational_jobs
    ADD COLUMN activation_run_id text REFERENCES booking_activation_runs(run_id),
    ADD COLUMN activation_stage_index integer,
    ADD CONSTRAINT operational_jobs_activation_stage
        FOREIGN KEY (activation_run_id, activation_stage_index)
        REFERENCES booking_activation_stages(run_id, stage_index),
    ADD CONSTRAINT operational_jobs_activation_pair
        CHECK ((activation_run_id IS NULL) = (activation_stage_index IS NULL));
