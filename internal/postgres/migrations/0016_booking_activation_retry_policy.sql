ALTER TABLE booking_activation_stages
    ADD COLUMN wait_after_ns bigint NOT NULL DEFAULT 0 CHECK (wait_after_ns >= 0),
    ADD COLUMN retry_initial_delay_ns bigint NOT NULL DEFAULT 0 CHECK (retry_initial_delay_ns >= 0),
    ADD COLUMN retry_backoff double precision NOT NULL DEFAULT 1 CHECK (retry_backoff >= 1),
    ADD COLUMN retry_maximum_delay_ns bigint NOT NULL DEFAULT 0 CHECK (retry_maximum_delay_ns >= 0),
    ADD COLUMN retry_total_timeout_ns bigint NOT NULL DEFAULT 0 CHECK (retry_total_timeout_ns >= 0),
    ADD COLUMN retryable_codes jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN retry_message text NOT NULL DEFAULT '',
    ADD COLUMN failure_guidance jsonb;
