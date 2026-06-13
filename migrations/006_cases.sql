CREATE TABLE IF NOT EXISTS cases (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    transaction_id TEXT NOT NULL,
    decision TEXT NOT NULL CHECK (decision IN ('APPROVE', 'DECLINE', 'ESCALATE')),
    reasoning TEXT NOT NULL,
    signals JSONB NOT NULL DEFAULT '{}'::jsonb,
    embedding vector(1536) NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cases_transaction_id ON cases(transaction_id);

-- ivfflat cannot use IF NOT EXISTS on older pgvector builds, so we guard it with a DO block.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_cases_embedding') THEN
        CREATE INDEX idx_cases_embedding ON cases USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
    END IF;
END$$;
