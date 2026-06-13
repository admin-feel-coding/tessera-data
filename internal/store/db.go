// Package store manages database connectivity for tessera-data.
package store

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect initializes a pgx connection pool. It does NOT fail if the DB is unreachable —
// it logs a warning and returns a pool that will retry connections on first use.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Warn("database unreachable at startup — will retry on first query", "error", err)
	}
	return pool, nil
}
