CREATE TABLE resource_availability (
    resource_name text PRIMARY KEY,
    available boolean NOT NULL,
    reason text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at_ns bigint NOT NULL
);

CREATE TABLE slot_availability (
    slot_name text PRIMARY KEY,
    available boolean NOT NULL,
    reason text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at_ns bigint NOT NULL
);
