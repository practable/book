DROP INDEX operational_usage_ledger_triggering_booking;

ALTER TABLE operational_usage_ledger DROP CONSTRAINT operational_usage_ledger_phase_check;
ALTER TABLE operational_usage_ledger ADD CONSTRAINT operational_usage_ledger_phase_check
    CHECK (phase IN ('preparation','cleanup'));

ALTER TABLE operational_usage_ledger DROP COLUMN triggering_booking_row_id;
