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
	ID                 string         `json:"id"`
	TransactionID      string         `json:"transaction_id"`
	Decision           string         `json:"decision"`
	RiskScore          float64        `json:"risk_score"`
	Reasoning          string         `json:"reasoning"`
	Signals            map[string]any `json:"signals"`
	CitedSources       []any          `json:"cited_sources"`
	EscalationReason   *string        `json:"escalation_reason"`
	EscalationCategory *string        `json:"escalation_category"`
	LatencyMs          int            `json:"latency_ms"`
	Model              string         `json:"model"`
	ToolCalls          int            `json:"tool_calls"`
	LangfuseTraceID    string         `json:"langfuse_trace_id"`
	CreatedAt          time.Time      `json:"created_at"`
	InputTokens        *int           `json:"input_tokens,omitempty"`
	OutputTokens       *int           `json:"output_tokens,omitempty"`
	CostUSD            *float64       `json:"cost_usd,omitempty"`
}

// SaveVerdictInput holds the fields required to persist a new verdict.
type SaveVerdictInput struct {
	TransactionID      string         `json:"transaction_id"`
	Decision           string         `json:"decision"`
	RiskScore          float64        `json:"risk_score"`
	Reasoning          string         `json:"reasoning"`
	Signals            map[string]any `json:"signals"`
	CitedSources       []any          `json:"cited_sources"`
	EscalationReason   *string        `json:"escalation_reason"`
	EscalationCategory *string        `json:"escalation_category"`
	LatencyMs          int            `json:"latency_ms"`
	Model              string         `json:"model"`
	ToolCalls          int            `json:"tool_calls"`
	LangfuseTraceID    string         `json:"langfuse_trace_id"`
	InputTokens        *int           `json:"input_tokens,omitempty"`
	OutputTokens       *int           `json:"output_tokens,omitempty"`
	CostUSD            *float64       `json:"cost_usd,omitempty"`
	// Velocity fields — written back to the transactions table.
	UserID    string
	Amount    float64
	Currency  string
	IPAddress string
	CardBin   string
	DeviceID  string
}

// SaveVerdict inserts a new verdict and returns the generated id.
// It also upserts the velocity columns (ip_address, card_bin, device_id) into the
// transactions table — only filling them in when they are currently NULL, making
// this operation idempotent.
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
			signals, cited_sources, escalation_reason, escalation_category,
			latency_ms, model, tool_calls, langfuse_trace_id,
			input_tokens, output_tokens, cost_usd
		) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`,
		v.TransactionID,
		v.Decision,
		v.RiskScore,
		v.Reasoning,
		string(signalsJSON),
		string(citedSourcesJSON),
		v.EscalationReason,
		v.EscalationCategory,
		v.LatencyMs,
		v.Model,
		v.ToolCalls,
		v.LangfuseTraceID,
		v.InputTokens,
		v.OutputTokens,
		v.CostUSD,
	).Scan(&id)
	if err != nil {
		return "", err
	}

	// Backfill velocity columns into an existing transaction row.
	// Uses UPDATE (not INSERT) to avoid FK violations when the transaction
	// doesn't exist in the users-referenced transactions table (e.g. eval runs).
	// COALESCE keeps any already-populated value, making this idempotent.
	if v.TransactionID != "" && (v.IPAddress != "" || v.CardBin != "" || v.DeviceID != "") {
		var ipAddr, cardBin, deviceID *string
		if v.IPAddress != "" {
			ipAddr = &v.IPAddress
		}
		if v.CardBin != "" {
			cardBin = &v.CardBin
		}
		if v.DeviceID != "" {
			deviceID = &v.DeviceID
		}

		_, err = pool.Exec(ctx, `
			UPDATE transactions SET
			    ip_address = COALESCE(ip_address, $2),
			    card_bin   = COALESCE(card_bin,   $3),
			    device_id  = COALESCE(device_id,  $4)
			WHERE id = $1
		`, v.TransactionID, ipAddr, cardBin, deviceID)
		if err != nil {
			return "", err
		}
	}

	return id, nil
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
		       signals, cited_sources, escalation_reason, escalation_category,
		       latency_ms, model, tool_calls, langfuse_trace_id, created_at,
		       input_tokens, output_tokens, cost_usd
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
			&signalsJSON, &citedSourcesJSON, &v.EscalationReason, &v.EscalationCategory,
			&v.LatencyMs, &v.Model, &v.ToolCalls, &v.LangfuseTraceID, &v.CreatedAt,
			&v.InputTokens, &v.OutputTokens, &v.CostUSD,
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
		       signals, cited_sources, escalation_reason, escalation_category,
		       latency_ms, model, tool_calls, langfuse_trace_id, created_at,
		       input_tokens, output_tokens, cost_usd
		FROM verdicts
		WHERE transaction_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, transactionID).Scan(
		&v.ID, &v.TransactionID, &v.Decision, &v.RiskScore, &v.Reasoning,
		&signalsJSON, &citedSourcesJSON, &v.EscalationReason, &v.EscalationCategory,
		&v.LatencyMs, &v.Model, &v.ToolCalls, &v.LangfuseTraceID, &v.CreatedAt,
		&v.InputTokens, &v.OutputTokens, &v.CostUSD,
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
