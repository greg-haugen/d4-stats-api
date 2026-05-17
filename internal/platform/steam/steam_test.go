package steam_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ghaugen/d4-stats-api/internal/model"
	"github.com/ghaugen/d4-stats-api/internal/platform"
	"github.com/ghaugen/d4-stats-api/internal/platform/steam"
)

const statsFixture = `{
	"displayName": "SteamPlayer",
	"characters": [
		{"name": "Vex", "class": "sorcerer", "level": 80, "paragonLevel": 50}
	],
	"season": {"seasonNumber": 4, "journeyProgress": 30}
}`

func newSteamServer(t *testing.T, keysSeen *[]string, statusCode int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*keysSeen = append(*keysSeen, r.URL.Query().Get("key"))
		w.WriteHeader(statusCode)
		if statusCode == http.StatusOK {
			w.Write([]byte(body))
		}
	}))
}

func TestFetchStats_Success(t *testing.T) {
	var keys []string
	srv := newSteamServer(t, &keys, http.StatusOK, statsFixture)
	defer srv.Close()

	client := steam.NewClient(srv.URL, "my-steam-key")
	stats, err := client.FetchStats(context.Background(), "SteamPlayer")
	if err != nil {
		t.Fatalf("FetchStats error: %v", err)
	}

	want := &model.PlayerStats{
		Username: "SteamPlayer",
		Platform: "steam",
		Characters: []model.Character{
			{Name: "Vex", Class: "sorcerer", Level: 80, ParagonLevel: 50},
		},
		Season: model.SeasonData{SeasonNumber: 4, JourneyProgress: 30},
	}

	if stats.Platform != want.Platform {
		t.Errorf("Platform = %q, want %q", stats.Platform, want.Platform)
	}
	if stats.Username != want.Username {
		t.Errorf("Username = %q, want %q", stats.Username, want.Username)
	}
	if len(stats.Characters) != 1 || stats.Characters[0] != want.Characters[0] {
		t.Errorf("Characters = %+v, want %+v", stats.Characters, want.Characters)
	}
	if stats.Season != want.Season {
		t.Errorf("Season = %+v, want %+v", stats.Season, want.Season)
	}
}

func TestFetchStats_APIKeyPresentOnRequest(t *testing.T) {
	var keys []string
	srv := newSteamServer(t, &keys, http.StatusOK, statsFixture)
	defer srv.Close()

	client := steam.NewClient(srv.URL, "my-steam-key")
	client.FetchStats(context.Background(), "SteamPlayer")

	if len(keys) == 0 || keys[0] != "my-steam-key" {
		t.Errorf("expected key=my-steam-key in query, got %v", keys)
	}
}

func TestFetchStats_RateLimited(t *testing.T) {
	var keys []string
	srv := newSteamServer(t, &keys, http.StatusTooManyRequests, "")
	defer srv.Close()

	client := steam.NewClient(srv.URL, "my-steam-key")
	_, err := client.FetchStats(context.Background(), "SteamPlayer")
	if !errors.Is(err, platform.ErrPlatformRateLimited) {
		t.Errorf("err = %v, want ErrPlatformRateLimited", err)
	}
}

func TestFetchStats_PlayerNotFound(t *testing.T) {
	var keys []string
	srv := newSteamServer(t, &keys, http.StatusNotFound, "")
	defer srv.Close()

	client := steam.NewClient(srv.URL, "my-steam-key")
	_, err := client.FetchStats(context.Background(), "ghost")
	if !errors.Is(err, platform.ErrPlayerNotFound) {
		t.Errorf("err = %v, want ErrPlayerNotFound", err)
	}
}
