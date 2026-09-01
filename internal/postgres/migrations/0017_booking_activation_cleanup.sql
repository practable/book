ALTER TABLE booking_activation_runs DROP CONSTRAINT booking_activation_runs_state_check;
ALTER TABLE booking_activation_runs ADD CONSTRAINT booking_activation_runs_state_check
    CHECK (state IN ('preparing','active','failed','cancelled','expired','cleaning','closed','cleanup_failed'));

ALTER TABLE booking_activation_stages
    ADD COLUMN phase text NOT NULL DEFAULT 'activation'
    CHECK (phase IN ('activation','cleanup'));

CREATE INDEX booking_activation_stages_phase
    ON booking_activation_stages (run_id, phase, stage_index);
