ALTER TABLE operational_usage_ledger DROP CONSTRAINT operational_usage_ledger_payer_kind_check;
ALTER TABLE operational_usage_ledger ADD CONSTRAINT operational_usage_ledger_payer_kind_check
    CHECK (payer_kind IN ('user','organisation','service'));
