package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserHistory aggregates a user's transaction history and velocity signals.
type UserHistory struct {
	UserID           string
	TransactionCount int
	AvgAmount        float64
	Countries        []string
	LastTxnAt        *time.Time
	HighVelocity     bool
}

// GetUserHistory queries transaction aggregates for userID.
// Returns ErrNotFound if the user row does not exist.
func GetUserHistory(ctx context.Context, pool *pgxpool.Pool, userID string) (UserHistory, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`,
		userID,
	).Scan(&exists)
	if err != nil {
		return UserHistory{}, err
	}
	if !exists {
		return UserHistory{}, ErrNotFound
	}

	h := UserHistory{UserID: userID}

	// Aggregate count, avg, distinct countries, and most recent timestamp in one pass.
	row := pool.QueryRow(ctx, `
		SELECT
			COUNT(*)::int,
			COALESCE(AVG(amount), 0)::float8,
			ARRAY_REMOVE(ARRAY_AGG(DISTINCT currency), NULL),
			MAX(created_at)
		FROM transactions
		WHERE user_id = $1
	`, userID)

	var countries []string
	var lastTxnAt *time.Time
	if err := row.Scan(&h.TransactionCount, &h.AvgAmount, &countries, &lastTxnAt); err != nil {
		if err == pgx.ErrNoRows {
			return h, nil
		}
		return UserHistory{}, err
	}
	h.Countries = countries
	h.LastTxnAt = lastTxnAt

	// High velocity: more than 20 transactions in the last 24 hours.
	var recentCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM transactions
		WHERE user_id = $1 AND created_at >= NOW() - INTERVAL '24 hours'
	`, userID).Scan(&recentCount)
	if err != nil {
		return UserHistory{}, err
	}
	h.HighVelocity = recentCount > 20

	return h, nil
}
