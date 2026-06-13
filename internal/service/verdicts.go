package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feel-coding/tessera-data/internal/store"
)

var validDecisions = map[string]bool{
	"APPROVE":  true,
	"DECLINE":  true,
	"ESCALATE": true,
}

// SaveVerdictInput is the input for persisting a grounded verdict.
type SaveVerdictInput struct {
	TransactionID    string         `json:"transaction_id"`
	Decision         string         `json:"decision"`
	RiskScore        float64        `json:"risk_score"`
	Reasoning        string         `json:"reasoning"`
	Signals          map[string]any `json:"signals"`
	CitedSources     []any          `json:"cited_sources"`
	EscalationReason *string        `json:"escalation_reason"`
	LatencyMs        int            `json:"latency_ms"`
	Model            string         `json:"model"`
	ToolCalls        int            `json:"tool_calls"`
	LangfuseTraceID  string         `json:"langfuse_trace_id"`
}

// SaveVerdict validates the input and persists a new verdict.
func SaveVerdict(ctx context.Context, pool *pgxpool.Pool, in SaveVerdictInput) (string, error) {
	if !validDecisions[in.Decision] {
		return "", errors.New("decision must be one of APPROVE, DECLINE, ESCALATE")
	}
	s := &store.SaveVerdictInput{
		TransactionID:    in.TransactionID,
		Decision:         in.Decision,
		RiskScore:        in.RiskScore,
		Reasoning:        in.Reasoning,
		Signals:          in.Signals,
		CitedSources:     in.CitedSources,
		EscalationReason: in.EscalationReason,
		LatencyMs:        in.LatencyMs,
		Model:            in.Model,
		ToolCalls:        in.ToolCalls,
		LangfuseTraceID:  in.LangfuseTraceID,
	}
	return store.SaveVerdict(ctx, pool, s)
}

// ListVerdicts returns a paginated list of verdicts and the total count.
func ListVerdicts(ctx context.Context, pool *pgxpool.Pool, limit, offset int) ([]store.Verdict, int, error) {
	return store.ListVerdicts(ctx, pool, limit, offset)
}

// GetVerdictByTransactionID returns the most recent verdict for a transaction_id.
func GetVerdictByTransactionID(ctx context.Context, pool *pgxpool.Pool, transactionID string) (*store.Verdict, error) {
	return store.GetVerdictByTransactionID(ctx, pool, transactionID)
}
