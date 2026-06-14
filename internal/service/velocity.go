package service

import (
	"context"
	"time"

	"github.com/feel-coding/tessera-data/internal/store"
)

// VelocityService wraps VelocityStore with any business-layer logic.
type VelocityService struct {
	store *store.VelocityStore
}

// NewVelocityService creates a VelocityService backed by s.
func NewVelocityService(s *store.VelocityStore) *VelocityService {
	return &VelocityService{store: s}
}

// Check returns cross-user velocity counts for ip and bin over windowMinutes.
func (s *VelocityService) Check(ctx context.Context, ip, bin string, windowMinutes int) (store.VelocityResult, error) {
	return s.store.Check(ctx, ip, bin, time.Duration(windowMinutes)*time.Minute)
}
