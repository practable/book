CREATE TABLE operational_stream_health (
    resource_name text NOT NULL,
    stream_name text NOT NULL,
    status text NOT NULL CHECK (status IN ('healthy','unhealthy','unknown')),
    result_code text NOT NULL DEFAULT '',
    message text NOT NULL DEFAULT '',
    job_id text REFERENCES operational_jobs(job_id),
    manifest_version bigint NOT NULL,
    checked_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (resource_name, stream_name)
);

CREATE TABLE operational_alerts (
    alert_id bigserial PRIMARY KEY,
    resource_name text NOT NULL,
    stream_name text NOT NULL,
    result_code text NOT NULL,
    message text NOT NULL DEFAULT '',
    job_id text REFERENCES operational_jobs(job_id),
    manifest_version bigint NOT NULL,
    status text NOT NULL CHECK (status IN ('open','acknowledged','resolved')),
    occurrences bigint NOT NULL DEFAULT 1 CHECK (occurrences > 0),
    first_seen_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    acknowledged_at timestamptz,
    acknowledged_by text NOT NULL DEFAULT '',
    resolved_at timestamptz,
    resolved_by text NOT NULL DEFAULT '',
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE UNIQUE INDEX operational_alerts_one_active
    ON operational_alerts (resource_name, stream_name, result_code)
    WHERE status IN ('open','acknowledged');
CREATE INDEX operational_alerts_dashboard
    ON operational_alerts (status, last_seen_at DESC);
