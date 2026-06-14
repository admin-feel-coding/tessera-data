ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS ip_address  INET,
    ADD COLUMN IF NOT EXISTS card_bin    VARCHAR(8),
    ADD COLUMN IF NOT EXISTS device_id   VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_txn_ip_time  ON transactions (ip_address, created_at DESC)
    WHERE ip_address IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_txn_bin_time ON transactions (card_bin,   created_at DESC)
    WHERE card_bin IS NOT NULL;
