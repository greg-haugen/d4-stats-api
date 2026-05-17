package platform

import (
	"context"
	"errors"

	"github.com/ghaugen/d4-stats-api/internal/model"
)

// Client is the platform-agnostic interface all platform packages implement.
type Client interface {
	FetchStats(ctx context.Context, username string) (*model.PlayerStats, error)
}

var (
	ErrPlayerNotFound      = errors.New("player not found")
	ErrPSNTokenInvalid     = errors.New("PSN token invalid or expired")
	ErrPlatformRateLimited = errors.New("platform rate limited")
)
