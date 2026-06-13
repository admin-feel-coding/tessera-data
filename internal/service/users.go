package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feel-coding/tessera-data/internal/store"
)

// UserHistory is the service-layer view of a user's transaction history.
type UserHistory struct {
	UserID           string   `json:"user_id"`
	TransactionCount int      `json:"transaction_count"`
	AvgAmount        float64  `json:"avg_amount"`
	Countries        []string `json:"countries"`
	LastTxnAt        *string  `json:"last_txn_at"`
	HighVelocity     bool     `json:"high_velocity"`
	NewUser          bool     `json:"new_user"`
}

// GetUserHistory returns transaction history for userID.
// A user with no transactions is valid (new_user=true, not an error).
func GetUserHistory(ctx context.Context, pool *pgxpool.Pool, userID string) (UserHistory, error) {
	h, err := store.GetUserHistory(ctx, pool, userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return UserHistory{
				UserID:    userID,
				NewUser:   true,
				Countries: []string{},
			}, nil
		}
		return UserHistory{}, err
	}

	result := UserHistory{
		UserID:           h.UserID,
		TransactionCount: h.TransactionCount,
		AvgAmount:        h.AvgAmount,
		Countries:        h.Countries,
		HighVelocity:     h.HighVelocity,
	}
	if result.Countries == nil {
		result.Countries = []string{}
	}
	if h.LastTxnAt != nil {
		s := h.LastTxnAt.UTC().Format("2006-01-02T15:04:05Z")
		result.LastTxnAt = &s
	}
	return result, nil
}
