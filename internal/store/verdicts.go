package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Verdict is the stored verdict record.
type Verdict struct {
	ID               string         `json:"id"`
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
	CreatedAt        time.Time      `json:"created_at"`
}

// SaveVerdictInput holds the fields required to persist a new verdict.
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

// SaveVerdict inserts a new verdict and returns the generated id.
func SaveVerdict(ctx context.Context, pool *pgxpool.Pool, v *SaveVerdictInput) (string, error) {
	signalsJSON, err := json.Marshal(v.Signals)
	if err != nil {
		return "", err
	}
	citedSourcesJSON, err := json.Marshal(v.CitedSources)
	if err != nil {
		return "", err
	}

	var id string
	err = pool.QueryRow(ctx, `
		INSERT INTO verdicts (
			transaction_id, decision, risk_score, reasoning,
			signals, cited_sources, escalation_reason,
			latency_ms, model, tool_calls, langfuse_trace_id
		) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8, $9, $10, $11)
		RETURNING id
	`,
		v.TransactionID,
		v.Decision,
		v.RiskScore,
		v.Reasoning,
		string(signalsJSON),
		string(citedSourcesJSON),
		v.EscalationReason,
		v.LatencyMs,
		v.Model,
		v.ToolCalls,
		v.LangfuseTraceID,
	).Scan(&id)
	return id, err
}

// ListVerdicts returns verdicts ordered by created_at DESC with pagination.
// Returns the verdict slice and total count of all verdicts in the table.
func ListVerdicts(ctx context.Context, pool *pgxpool.Pool, limit, offset int) ([]Verdict, int, error) {
	var total int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM verdicts`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, transaction_id, decision, risk_score, reasoning,
		       signals, cited_sources, escalation_reason,
		       latency_ms, model, tool_calls, langfuse_trace_id, created_at
		FROM verdicts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	verdicts := make([]Verdict, 0)
	for rows.Next() {
		var v Verdict
		var signalsJSON, citedSourcesJSON []byte
		if err := rows.Scan(
			&v.ID, &v.TransactionID, &v.Decision, &v.RiskScore, &v.Reasoning,
			&signalsJSON, &citedSourcesJSON, &v.EscalationReason,
			&v.LatencyMs, &v.Model, &v.ToolCalls, &v.LangfuseTraceID, &v.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(signalsJSON, &v.Signals); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(citedSourcesJSON, &v.CitedSources); err != nil {
			return nil, 0, err
		}
		verdicts = append(verdicts, v)
	}
	return verdicts, total, rows.Err()
}

// GetVerdictByTransactionID returns the most recent verdict for a transaction_id.
// Returns ErrNotFound if no matching verdict exists.
func GetVerdictByTransactionID(ctx context.Context, pool *pgxpool.Pool, transactionID string) (*Verdict, error) {
	var v Verdict
	var signalsJSON, citedSourcesJSON []byte

	err := pool.QueryRow(ctx, `
		SELECT id, transaction_id, decision, risk_score, reasoning,
		       signals, cited_sources, escalation_reason,
		       latency_ms, model, tool_calls, langfuse_trace_id, created_at
		FROM verdicts
		WHERE transaction_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, transactionID).Scan(
		&v.ID, &v.TransactionID, &v.Decision, &v.RiskScore, &v.Reasoning,
		&signalsJSON, &citedSourcesJSON, &v.EscalationReason,
		&v.LatencyMs, &v.Model, &v.ToolCalls, &v.LangfuseTraceID, &v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(signalsJSON, &v.Signals); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(citedSourcesJSON, &v.CitedSources); err != nil {
		return nil, err
	}
	return &v, nil
}
