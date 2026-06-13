package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feel-coding/tessera-data/internal/store"
)

// IPRisk is the service-layer view of an IP risk signal, with a derived high-risk flag.
type IPRisk struct {
	IP         string  `json:"ip"`
	RiskScore  float64 `json:"risk_score"`
	Country    string  `json:"country"`
	IsVPN      bool    `json:"is_vpn"`
	LastSeen   string  `json:"last_seen"`
	IsHighRisk bool    `json:"is_high_risk"`
}

// GetIPRisk returns risk signals for the given IP address.
// Returns store.ErrNotFound if the IP is not in the database.
func GetIPRisk(ctx context.Context, pool *pgxpool.Pool, ip string) (IPRisk, error) {
	s, err := store.GetIPRisk(ctx, pool, ip)
	if err != nil {
		return IPRisk{}, err
	}
	return IPRisk{
		IP:         s.IP,
		RiskScore:  s.RiskScore,
		Country:    s.Country,
		IsVPN:      s.IsVPN,
		LastSeen:   s.LastSeen.UTC().Format("2006-01-02T15:04:05Z"),
		IsHighRisk: s.RiskScore > 0.7,
	}, nil
}
