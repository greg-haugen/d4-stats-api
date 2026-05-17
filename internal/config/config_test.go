package config_test

import (
	"testing"

	"github.com/ghaugen/d4-stats-api/internal/config"
)

func TestLoad_AllVarsPresent(t *testing.T) {
	t.Setenv("BNET_CLIENT_ID", "bnet-id")
	t.Setenv("BNET_CLIENT_SECRET", "bnet-secret")
	t.Setenv("STEAM_API_KEY", "steam-key")
	t.Setenv("PSN_AUTH_TOKEN", "psn-token")
	t.Setenv("XBOX_CLIENT_ID", "xbox-id")
	t.Setenv("XBOX_CLIENT_SECRET", "xbox-secret")
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("RATE_LIMIT_RPS", "10")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.BnetClientID != "bnet-id" {
		t.Errorf("BnetClientID = %q, want %q", cfg.BnetClientID, "bnet-id")
	}
	if cfg.BnetClientSecret != "bnet-secret" {
		t.Errorf("BnetClientSecret = %q, want %q", cfg.BnetClientSecret, "bnet-secret")
	}
	if cfg.SteamAPIKey != "steam-key" {
		t.Errorf("SteamAPIKey = %q, want %q", cfg.SteamAPIKey, "steam-key")
	}
	if cfg.PSNAuthToken != "psn-token" {
		t.Errorf("PSNAuthToken = %q, want %q", cfg.PSNAuthToken, "psn-token")
	}
	if cfg.XboxClientID != "xbox-id" {
		t.Errorf("XboxClientID = %q, want %q", cfg.XboxClientID, "xbox-id")
	}
	if cfg.XboxClientSecret != "xbox-secret" {
		t.Errorf("XboxClientSecret = %q, want %q", cfg.XboxClientSecret, "xbox-secret")
	}
	if cfg.JWTSecret != "jwt-secret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "jwt-secret")
	}
	if cfg.RateLimitRPS != 10 {
		t.Errorf("RateLimitRPS = %d, want %d", cfg.RateLimitRPS, 10)
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	vars := []string{
		"BNET_CLIENT_ID", "BNET_CLIENT_SECRET", "STEAM_API_KEY",
		"PSN_AUTH_TOKEN", "XBOX_CLIENT_ID", "XBOX_CLIENT_SECRET",
		"JWT_SECRET", "RATE_LIMIT_RPS",
	}
	for _, v := range vars {
		t.Setenv(v, "")
	}

	for _, missing := range vars {
		t.Run("missing_"+missing, func(t *testing.T) {
			// Set all vars, then clear the one under test.
			t.Setenv("BNET_CLIENT_ID", "x")
			t.Setenv("BNET_CLIENT_SECRET", "x")
			t.Setenv("STEAM_API_KEY", "x")
			t.Setenv("PSN_AUTH_TOKEN", "x")
			t.Setenv("XBOX_CLIENT_ID", "x")
			t.Setenv("XBOX_CLIENT_SECRET", "x")
			t.Setenv("JWT_SECRET", "x")
			t.Setenv("RATE_LIMIT_RPS", "1")
			t.Setenv(missing, "")

			_, err := config.Load()
			if err == nil {
				t.Errorf("expected error when %s is missing, got nil", missing)
			}
		})
	}
}

func TestLoad_NoGlobalState(t *testing.T) {
	t.Setenv("BNET_CLIENT_ID", "a")
	t.Setenv("BNET_CLIENT_SECRET", "a")
	t.Setenv("STEAM_API_KEY", "a")
	t.Setenv("PSN_AUTH_TOKEN", "a")
	t.Setenv("XBOX_CLIENT_ID", "a")
	t.Setenv("XBOX_CLIENT_SECRET", "a")
	t.Setenv("JWT_SECRET", "a")
	t.Setenv("RATE_LIMIT_RPS", "5")

	cfg1, _ := config.Load()

	t.Setenv("STEAM_API_KEY", "b")
	cfg2, _ := config.Load()

	if cfg1.SteamAPIKey == cfg2.SteamAPIKey {
		t.Error("Load() returns shared state: second call reflects env change on first result")
	}
}
