ALTER TABLE booking_activation_runs
    ADD COLUMN cleanup_state text NOT NULL DEFAULT 'pending'
    CHECK (cleanup_state IN ('not_required','pending','running','succeeded','failed'));

UPDATE booking_activation_runs SET cleanup_state='not_required'
WHERE NOT EXISTS (
    SELECT 1 FROM booking_activation_stages s
    WHERE s.run_id=booking_activation_runs.run_id AND s.phase='cleanup'
);
