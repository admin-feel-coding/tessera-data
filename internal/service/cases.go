// Package service contains business logic for tessera-data.
package service

import (
	"context"

	"github.com/feel-coding/tessera-data/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SimilarSearchInput is the input for pgvector similarity search.
type SimilarSearchInput struct {
	Embedding []float32 `json:"embedding"`
	Limit     int       `json:"limit"`
}

// FindSimilarCases retrieves top-k similar historical cases via pgvector cosine search.
func FindSimilarCases(ctx context.Context, pool *pgxpool.Pool, in SimilarSearchInput) ([]store.Case, error) {
	return store.FindSimilarCases(ctx, pool, in.Embedding, in.Limit)
}

// SaveCaseInput is the input for persisting a grounded fraud case.
type SaveCaseInput struct {
	TransactionID string         `json:"transaction_id"`
	Decision      string         `json:"decision"`
	Reasoning     string         `json:"reasoning"`
	Signals       map[string]any `json:"signals"`
	Embedding     []float32      `json:"embedding"`
	Metadata      map[string]any `json:"metadata"`
}

// SaveCase persists a new grounded case from analyst feedback.
func SaveCase(ctx context.Context, pool *pgxpool.Pool, in SaveCaseInput) (string, error) {
	c := store.Case{
		TransactionID: in.TransactionID,
		Decision:      in.Decision,
		Reasoning:     in.Reasoning,
		Signals:       in.Signals,
		Metadata:      in.Metadata,
	}
	return store.SaveCase(ctx, pool, c, in.Embedding)
}
