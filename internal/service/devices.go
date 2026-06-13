package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feel-coding/tessera-data/internal/store"
)

// DeviceFingerprint is the service-layer view of a device fingerprint record.
type DeviceFingerprint struct {
	DeviceID   string `json:"device_id"`
	FirstSeen  string `json:"first_seen"`
	LastSeen   string `json:"last_seen"`
	UserCount  int    `json:"user_count"`
	Suspicious bool   `json:"suspicious"`
}

// GetDeviceFingerprint returns fingerprint data for deviceID.
// Returns store.ErrNotFound if the device is not in the database.
func GetDeviceFingerprint(ctx context.Context, pool *pgxpool.Pool, deviceID string) (DeviceFingerprint, error) {
	f, err := store.GetDeviceFingerprint(ctx, pool, deviceID)
	if err != nil {
		return DeviceFingerprint{}, err
	}
	return DeviceFingerprint{
		DeviceID:   f.DeviceID,
		FirstSeen:  f.FirstSeen.UTC().Format("2006-01-02T15:04:05Z"),
		LastSeen:   f.LastSeen.UTC().Format("2006-01-02T15:04:05Z"),
		UserCount:  f.UserCount,
		Suspicious: f.Suspicious,
	}, nil
}
