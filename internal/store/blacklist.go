package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BlacklistResult holds the outcome of a blacklist lookup.
// Match is false when none of the supplied identifiers appear in the blacklist — this is the normal case.
type BlacklistResult struct {
	Match  bool
	Kind   string
	Reason string
}

// CheckBlacklist checks whether userID, email, or cardBin matches any blacklist entry.
// Returns BlacklistResult{Match: false} (not ErrNotFound) when no match is found.
func CheckBlacklist(ctx context.Context, pool *pgxpool.Pool, userID, email, cardBin string) (BlacklistResult, error) {
	var result BlacklistResult
	err := pool.QueryRow(ctx, `
		SELECT kind, reason
		FROM blacklist
		WHERE (kind = 'user_id' AND value = $1)
		   OR (kind = 'email'   AND value = $2)
		   OR (kind = 'card_bin' AND value = $3)
		LIMIT 1
	`, userID, email, cardBin).Scan(&result.Kind, &result.Reason)
	if err == pgx.ErrNoRows {
		return BlacklistResult{Match: false}, nil
	}
	if err != nil {
		return BlacklistResult{}, err
	}
	result.Match = true
	return result, nil
}
