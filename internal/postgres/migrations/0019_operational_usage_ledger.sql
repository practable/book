CREATE TABLE operational_usage_ledger (
    ledger_id bigserial PRIMARY KEY,
    job_id text NOT NULL UNIQUE REFERENCES operational_jobs(job_id),
    activation_run_id text REFERENCES booking_activation_runs(run_id),
    triggering_booking_name text NOT NULL,
    user_name text NOT NULL,
    phase text NOT NULL CHECK (phase IN ('preparation','cleanup')),
    payer_kind text NOT NULL DEFAULT 'user' CHECK (payer_kind IN ('user','organisation','service')),
    payer_id text NOT NULL,
    chargeable boolean NOT NULL DEFAULT true,
    state text NOT NULL CHECK (state IN ('reserved','dispatched','accepted','running','succeeded','failed','cancelled','expired')),
    planned_duration_ns bigint NOT NULL CHECK (planned_duration_ns >= 0),
    actual_duration_ns bigint CHECK (actual_duration_ns >= 0),
    started_at timestamptz,
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX operational_usage_ledger_booking ON operational_usage_ledger (triggering_booking_name, phase);
CREATE INDEX operational_usage_ledger_payer ON operational_usage_ledger (payer_kind, payer_id, created_at);
