CREATE TABLE IF NOT EXISTS verdicts (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    transaction_id TEXT NOT NULL,
    decision TEXT NOT NULL CHECK (decision IN ('APPROVE', 'DECLINE', 'ESCALATE')),
    risk_score NUMERIC(5,4) NOT NULL DEFAULT 0,
    reasoning TEXT NOT NULL DEFAULT '',
    signals JSONB NOT NULL DEFAULT '{}'::jsonb,
    cited_sources JSONB NOT NULL DEFAULT '[]'::jsonb,
    escalation_reason TEXT,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    model TEXT NOT NULL DEFAULT '',
    tool_calls INTEGER NOT NULL DEFAULT 0,
    langfuse_trace_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_verdicts_created_at ON verdicts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_verdicts_decision ON verdicts(decision);
CREATE INDEX IF NOT EXISTS idx_verdicts_transaction_id ON verdicts(transaction_id);
