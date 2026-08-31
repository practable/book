CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version bigint PRIMARY KEY,
    name text NOT NULL,
    checksum text NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE bookings (
    row_id bigserial PRIMARY KEY,
    name text NOT NULL,
    collection text NOT NULL CHECK (collection IN ('live', 'history')),
    superseded boolean NOT NULL DEFAULT false,
    user_name text NOT NULL,
    policy_name text NOT NULL,
    slot_name text NOT NULL,
    resource_name text NOT NULL,
    resource_constrained boolean NOT NULL,
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    starts_ns bigint NOT NULL,
    ends_ns bigint NOT NULL,
    started_at timestamptz,
    started_at_ns bigint,
    cancelled_at timestamptz,
    cancelled_at_ns bigint,
    cancelled_by text NOT NULL DEFAULT '',
    unfulfilled boolean NOT NULL DEFAULT false,
    usage_charge_ns bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    CHECK (ends_ns >= starts_ns),
    CHECK ((started_at IS NULL) = (started_at_ns IS NULL)),
    CHECK ((cancelled_at IS NULL) = (cancelled_at_ns IS NULL)),
    CHECK ((cancelled_at IS NULL AND cancelled_by = '') OR
           (cancelled_at IS NOT NULL AND cancelled_by <> ''))
);

CREATE UNIQUE INDEX bookings_current_live_name
    ON bookings (name) WHERE collection = 'live' AND NOT superseded;
CREATE UNIQUE INDEX bookings_current_history_name
    ON bookings (name) WHERE collection = 'history' AND NOT superseded;
CREATE UNIQUE INDEX bookings_exact_active_request
    ON bookings (user_name, policy_name, slot_name, starts_ns, ends_ns)
    WHERE collection = 'live' AND NOT superseded;
ALTER TABLE bookings ADD CONSTRAINT bookings_no_resource_overlap
    EXCLUDE USING gist (
        resource_name WITH =,
        int8range(starts_ns, ends_ns, '[]') WITH &&
    ) WHERE (collection = 'live' AND NOT superseded AND resource_constrained);

CREATE INDEX bookings_user_policy_current
    ON bookings (user_name, policy_name, ends_ns)
    WHERE NOT superseded;

CREATE TABLE booking_events (
    event_id bigserial PRIMARY KEY,
    booking_row_id bigint NOT NULL REFERENCES bookings(row_id),
    booking_name text NOT NULL,
    event_type text NOT NULL CHECK (event_type IN
        ('created', 'started', 'cancelled', 'expired', 'imported', 'superseded')),
    occurred_at timestamptz NOT NULL,
    occurred_at_ns bigint NOT NULL,
    actor text NOT NULL DEFAULT '',
    details jsonb NOT NULL DEFAULT '{}'::jsonb
);
CREATE INDEX booking_events_booking_name ON booking_events (booking_name, event_id);

CREATE TABLE user_groups (
    user_name text NOT NULL,
    group_name text NOT NULL,
    granted_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (user_name, group_name)
);
