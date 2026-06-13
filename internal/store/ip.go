package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IPSignal holds risk metadata for a single IP address.
type IPSignal struct {
	IP        string
	RiskScore float64
	Country   string
	IsVPN     bool
	LastSeen  time.Time
}

// GetIPRisk looks up the risk signal row for ip.
// Returns ErrNotFound if the IP is not in the table.
func GetIPRisk(ctx context.Context, pool *pgxpool.Pool, ip string) (IPSignal, error) {
	var s IPSignal
	err := pool.QueryRow(ctx, `
		SELECT ip, risk_score::float8, country, is_vpn, last_seen
		FROM ip_signals
		WHERE ip = $1
	`, ip).Scan(&s.IP, &s.RiskScore, &s.Country, &s.IsVPN, &s.LastSeen)
	if err == pgx.ErrNoRows {
		return IPSignal{}, ErrNotFound
	}
	if err != nil {
		return IPSignal{}, err
	}
	return s, nil
}
