ALTER TABLE verdicts
    ADD COLUMN IF NOT EXISTS escalation_category VARCHAR(32);
