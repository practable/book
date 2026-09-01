DROP INDEX booking_activation_stages_phase;
ALTER TABLE booking_activation_stages DROP COLUMN phase;

ALTER TABLE booking_activation_runs DROP CONSTRAINT booking_activation_runs_state_check;
ALTER TABLE booking_activation_runs ADD CONSTRAINT booking_activation_runs_state_check
    CHECK (state IN ('preparing','active','failed','cancelled','expired'));
