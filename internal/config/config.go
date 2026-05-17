package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	BnetClientID     string
	BnetClientSecret string
	SteamAPIKey      string
	PSNAuthToken     string
	XboxClientID     string
	XboxClientSecret string
	JWTSecret        string
	RateLimitRPS     int
}

func Load() (*Config, error) {
	cfg := &Config{}
	var missing []string

	get := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			missing = append(missing, key)
		}
		return v
	}

	cfg.BnetClientID = get("BNET_CLIENT_ID")
	cfg.BnetClientSecret = get("BNET_CLIENT_SECRET")
	cfg.SteamAPIKey = get("STEAM_API_KEY")
	cfg.PSNAuthToken = get("PSN_AUTH_TOKEN")
	cfg.XboxClientID = get("XBOX_CLIENT_ID")
	cfg.XboxClientSecret = get("XBOX_CLIENT_SECRET")
	cfg.JWTSecret = get("JWT_SECRET")

	rpsStr := os.Getenv("RATE_LIMIT_RPS")
	if rpsStr == "" {
		missing = append(missing, "RATE_LIMIT_RPS")
	} else {
		rps, err := strconv.Atoi(rpsStr)
		if err != nil || rps <= 0 {
			return nil, fmt.Errorf("RATE_LIMIT_RPS must be a positive integer, got %q", rpsStr)
		}
		cfg.RateLimitRPS = rps
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	return cfg, nil
}
