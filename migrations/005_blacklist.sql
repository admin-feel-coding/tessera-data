CREATE TABLE IF NOT EXISTS blacklist (
    id SERIAL PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('email','user_id','card_bin')),
    value TEXT NOT NULL,
    reason TEXT NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_blacklist_kind_value ON blacklist(kind, value);
