package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// VelocityResult holds cross-user velocity counts for a given IP and card BIN.
type VelocityResult struct {
	DistinctUsersByIP  int `json:"distinct_users_by_ip"`
	DistinctUsersByBin int `json:"distinct_users_by_bin"`
	TotalTxnsInWindow  int `json:"total_txns_in_window"`
	WindowMinutes      int `json:"window_minutes"`
}

// VelocityStore queries cross-user velocity signals from the transactions table.
type VelocityStore struct {
	pool *pgxpool.Pool
}

// NewVelocityStore creates a VelocityStore backed by pool.
func NewVelocityStore(pool *pgxpool.Pool) *VelocityStore {
	return &VelocityStore{pool: pool}
}

// Check returns cross-user velocity counts for the given IP and card BIN over window.
// The query counts distinct users that share the IP, distinct users that share the BIN,
// and total transactions matching either signal within the window.
func (s *VelocityStore) Check(ctx context.Context, ip, bin string, window time.Duration) (VelocityResult, error) {
	windowMinutes := int(window.Minutes())

	var byIP, byBin, total int

	err := s.pool.QueryRow(ctx, `
		SELECT
		    COUNT(DISTINCT CASE WHEN ip_address = $1::inet THEN user_id END),
		    COUNT(DISTINCT CASE WHEN card_bin   = $2       THEN user_id END),
		    COUNT(*)
		FROM transactions
		WHERE (ip_address = $1::inet OR card_bin = $2)
		  AND created_at > NOW() - ($3 || ' minutes')::INTERVAL
	`, ip, bin, windowMinutes).Scan(&byIP, &byBin, &total)
	if err != nil {
		return VelocityResult{}, err
	}

	return VelocityResult{
		DistinctUsersByIP:  byIP,
		DistinctUsersByBin: byBin,
		TotalTxnsInWindow:  total,
		WindowMinutes:      windowMinutes,
	}, nil
}
