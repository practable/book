CREATE TABLE service_state (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    booking_creation_paused boolean NOT NULL DEFAULT false,
    welcome_message text NOT NULL DEFAULT 'Welcome to the interval booking store',
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at_ns bigint NOT NULL DEFAULT 0
);

INSERT INTO service_state(singleton,updated_at_ns) VALUES(true,0)
ON CONFLICT (singleton) DO NOTHING;
