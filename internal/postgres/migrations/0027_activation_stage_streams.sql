ALTER TABLE booking_activation_stages
    ADD COLUMN stream_name text NOT NULL DEFAULT '';

UPDATE booking_activation_stages AS stage
SET stream_name = run.stream_name
FROM booking_activation_runs AS run
WHERE run.run_id = stage.run_id;

ALTER TABLE booking_activation_stages
    ALTER COLUMN stream_name DROP DEFAULT;
