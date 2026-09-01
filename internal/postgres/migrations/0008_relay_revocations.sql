-- A durable, append-only safety control. Relay instances receive only SELECT
-- permission on this table; booking services are its sole writers.
CREATE TABLE relay_revocations (
    booking_row_id bigint PRIMARY KEY REFERENCES bookings(row_id),
    booking_name text NOT NULL,
    expires_at timestamptz NOT NULL,
    expires_at_ns bigint NOT NULL,
    revoked_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    revoked_by text NOT NULL DEFAULT ''
);
CREATE INDEX relay_revocations_live_lookup
    ON relay_revocations (booking_name, expires_at_ns);
