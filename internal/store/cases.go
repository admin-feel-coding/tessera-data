package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Case represents a stored fraud case with embedding for RAG retrieval.
type Case struct {
	ID            string         `json:"id"`
	TransactionID string         `json:"transaction_id"`
	Decision      string         `json:"decision"`
	Reasoning     string         `json:"reasoning"`
	Signals       map[string]any `json:"signals"`
	Metadata      map[string]any `json:"metadata"`
	CreatedAt     time.Time      `json:"created_at"`
	Similarity    float64        `json:"similarity,omitempty"` // only set in FindSimilarCases results
}

// FindSimilarCases returns up to limit cases ordered by cosine distance to embedding.
// Uses the pgvector <=> cosine-distance operator; similarity = 1 - distance.
// Returns an empty slice if the cases table is empty.
func FindSimilarCases(ctx context.Context, pool *pgxpool.Pool, embedding []float32, limit int) ([]Case, error) {
	if limit <= 0 || limit > 50 {
		limit = 5
	}
	vec := encodeVector(embedding)

	rows, err := pool.Query(ctx, `
		SELECT id, transaction_id, decision, reasoning, signals, metadata, created_at,
		       1 - (embedding <=> $1::vector) AS similarity
		FROM cases
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`, vec, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cases := make([]Case, 0)
	for rows.Next() {
		var c Case
		var signalsJSON, metadataJSON []byte
		if err := rows.Scan(&c.ID, &c.TransactionID, &c.Decision, &c.Reasoning, &signalsJSON, &metadataJSON, &c.CreatedAt, &c.Similarity); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(signalsJSON, &c.Signals); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(metadataJSON, &c.Metadata); err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}
	return cases, rows.Err()
}

// SaveCase inserts a new case and returns the generated id.
func SaveCase(ctx context.Context, pool *pgxpool.Pool, c Case, embedding []float32) (string, error) {
	signalsJSON, err := json.Marshal(c.Signals)
	if err != nil {
		return "", err
	}
	metadataJSON, err := json.Marshal(c.Metadata)
	if err != nil {
		return "", err
	}

	var id string
	err = pool.QueryRow(ctx, `
		INSERT INTO cases (transaction_id, decision, reasoning, signals, embedding, metadata)
		VALUES ($1, $2, $3, $4::jsonb, $5::vector, $6::jsonb)
		RETURNING id
	`, c.TransactionID, c.Decision, c.Reasoning, string(signalsJSON), encodeVector(embedding), string(metadataJSON)).Scan(&id)
	return id, err
}

// encodeVector formats a float32 slice as the pgvector text literal: [v1,v2,...].
// We use the text-cast approach ($1::vector) to avoid adding pgvector-go as a dependency.
func encodeVector(v []float32) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
