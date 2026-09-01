CREATE TABLE operational_jobs (
    job_id text PRIMARY KEY,
    resource_name text NOT NULL,
    workflow_name text NOT NULL,
    job_kind text NOT NULL CHECK (job_kind IN ('setup','teardown','settling','scheduled','health','preflight')),
    state text NOT NULL CHECK (state IN ('scheduled','reserved','dispatched','accepted','running','succeeded','failed','cancelled','expired')),
    due_at timestamptz NOT NULL,
    due_at_ns bigint NOT NULL,
    starts_at timestamptz NOT NULL,
    starts_at_ns bigint NOT NULL,
    ends_at timestamptz NOT NULL,
    ends_at_ns bigint NOT NULL,
    booking_row_id bigint REFERENCES bookings(row_id),
    triggering_booking_name text NOT NULL DEFAULT '',
    manifest_version bigint NOT NULL DEFAULT 0,
    plan_revision bigint NOT NULL DEFAULT 1,
    idempotency_key text NOT NULL UNIQUE,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    lease_owner text NOT NULL DEFAULT '',
    lease_until timestamptz,
    last_error text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    CHECK (ends_at_ns > starts_at_ns)
);
CREATE INDEX operational_jobs_due ON operational_jobs (state, due_at_ns);
CREATE INDEX operational_jobs_resource ON operational_jobs (resource_name, starts_at_ns, ends_at_ns);

CREATE TABLE webhook_deliveries (
    delivery_id text PRIMARY KEY,
    job_id text NOT NULL REFERENCES operational_jobs(job_id),
    direction text NOT NULL CHECK (direction = 'book-to-runner'),
    state text NOT NULL CHECK (state IN ('pending','leased','delivered','dead','cancelled')),
    body jsonb NOT NULL,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    next_attempt_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    next_attempt_at_ns bigint NOT NULL,
    lease_owner text NOT NULL DEFAULT '',
    lease_until timestamptz,
    response_status integer,
    last_error text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX webhook_deliveries_pending ON webhook_deliveries (state, next_attempt_at_ns);

CREATE TABLE webhook_callback_receipts (
    delivery_id text PRIMARY KEY,
    job_id text NOT NULL REFERENCES operational_jobs(job_id),
    body_sha256 text NOT NULL,
    received_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
