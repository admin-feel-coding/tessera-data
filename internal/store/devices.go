package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DeviceFingerprint holds usage metadata for a single device identifier.
type DeviceFingerprint struct {
	DeviceID   string
	FirstSeen  time.Time
	LastSeen   time.Time
	UserCount  int
	Suspicious bool
}

// GetDeviceFingerprint looks up the fingerprint row for deviceID.
// Returns ErrNotFound if the device is not in the table.
func GetDeviceFingerprint(ctx context.Context, pool *pgxpool.Pool, deviceID string) (DeviceFingerprint, error) {
	var f DeviceFingerprint
	err := pool.QueryRow(ctx, `
		SELECT device_id, first_seen, last_seen, user_count, suspicious
		FROM device_fingerprints
		WHERE device_id = $1
	`, deviceID).Scan(&f.DeviceID, &f.FirstSeen, &f.LastSeen, &f.UserCount, &f.Suspicious)
	if err == pgx.ErrNoRows {
		return DeviceFingerprint{}, ErrNotFound
	}
	if err != nil {
		return DeviceFingerprint{}, err
	}
	return f, nil
}
