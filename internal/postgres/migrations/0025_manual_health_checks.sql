ALTER TABLE booking_activation_runs
    ADD COLUMN auto_close boolean NOT NULL DEFAULT false;
