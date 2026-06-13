package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feel-coding/tessera-data/internal/store"
)

// BlacklistResult is the service-layer view of a blacklist check outcome.
type BlacklistResult struct {
	Match  bool   `json:"match"`
	Kind   string `json:"kind,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// CheckBlacklist checks whether userID, email, or cardBin matches any blacklist entry.
func CheckBlacklist(ctx context.Context, pool *pgxpool.Pool, userID, email, cardBin string) (BlacklistResult, error) {
	r, err := store.CheckBlacklist(ctx, pool, userID, email, cardBin)
	if err != nil {
		return BlacklistResult{}, err
	}
	return BlacklistResult{
		Match:  r.Match,
		Kind:   r.Kind,
		Reason: r.Reason,
	}, nil
}
