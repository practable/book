CREATE TABLE resource_release_state (
    resource_name text PRIMARY KEY,
    state text NOT NULL CHECK (state IN ('pending_checks','verified','degraded_override')),
    required_streams jsonb NOT NULL DEFAULT '[]'::jsonb,
    requested_at timestamptz NOT NULL,
    requested_by text NOT NULL,
    manifest_version bigint NOT NULL,
    override_reason text NOT NULL DEFAULT '',
    failing_streams jsonb NOT NULL DEFAULT '[]'::jsonb,
    released_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    CHECK (state <> 'degraded_override' OR override_reason <> '')
);

CREATE TABLE resource_release_events (
    event_id bigserial PRIMARY KEY,
    resource_name text NOT NULL,
    event_type text NOT NULL CHECK (event_type IN ('verification_requested','verified_release','degraded_override')),
    occurred_at timestamptz NOT NULL,
    actor text NOT NULL,
    reason text NOT NULL DEFAULT '',
    required_streams jsonb NOT NULL DEFAULT '[]'::jsonb,
    failing_streams jsonb NOT NULL DEFAULT '[]'::jsonb,
    manifest_version bigint NOT NULL
);

CREATE INDEX resource_release_events_resource
    ON resource_release_events(resource_name, occurred_at DESC);
