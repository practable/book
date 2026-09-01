CREATE TABLE booking_replacements (
    old_booking_row_id bigint PRIMARY KEY REFERENCES bookings(row_id),
    new_booking_row_id bigint NOT NULL UNIQUE REFERENCES bookings(row_id),
    replaced_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    replaced_at_ns bigint NOT NULL,
    actor text NOT NULL DEFAULT 'admin'
);
