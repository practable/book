ALTER TABLE operational_usage_ledger
    ADD COLUMN triggering_booking_row_id bigint REFERENCES bookings(row_id);

UPDATE operational_usage_ledger l
SET triggering_booking_row_id = r.booking_row_id
FROM booking_activation_runs r
WHERE r.run_id = l.activation_run_id;

ALTER TABLE operational_usage_ledger
    ALTER COLUMN triggering_booking_row_id SET NOT NULL;

ALTER TABLE operational_usage_ledger DROP CONSTRAINT operational_usage_ledger_phase_check;
ALTER TABLE operational_usage_ledger ADD CONSTRAINT operational_usage_ledger_phase_check
    CHECK (phase IN ('preparation','cleanup','quality_control','scheduled'));

CREATE INDEX operational_usage_ledger_triggering_booking
    ON operational_usage_ledger (triggering_booking_row_id, phase);
