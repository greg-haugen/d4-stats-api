package battlenet_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ghaugen/d4-stats-api/internal/model"
	"github.com/ghaugen/d4-stats-api/internal/platform"
	"github.com/ghaugen/d4-stats-api/internal/platform/battlenet"
)

// tokenFixture is the recorded OAuth token response from Blizzard.
const tokenFixture = `{"access_token":"test-token-123","token_type":"bearer","expires_in":86399}`

// statsFixture is a recorded Battle.net D4 profile response.
const statsFixture = `{
	"battletag": "Hero#1234",
	"characters": [
		{"name": "Draven", "class": "barbarian", "level": 100, "paragonLevel": 200}
	],
	"season": {"seasonNumber": 4, "journeyProgress": 75}
}`

func newTestServer(t *testing.T, tokenCalls *int, statusSequence []int, body string) *httptest.Server {
	t.Helper()
	callCount := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			*tokenCalls++
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(json.RawMessage(tokenFixture))
			return
		}
		status := 200
		if callCount < len(statusSequence) {
			status = statusSequence[callCount]
		}
		callCount++
		w.WriteHeader(status)
		if status == 200 {
			w.Write([]byte(body))
		}
	}))
}

func TestFetchStats_Success(t *testing.T) {
	var tokenCalls int
	srv := newTestServer(t, &tokenCalls, []int{200}, statsFixture)
	defer srv.Close()

	client := battlenet.NewClient(srv.URL, srv.URL, "client-id", "client-secret")
	stats, err := client.FetchStats(context.Background(), "Hero#1234")
	if err != nil {
		t.Fatalf("FetchStats error: %v", err)
	}

	want := &model.PlayerStats{
		Username: "Hero#1234",
		Platform: "battlenet",
		Characters: []model.Character{
			{Name: "Draven", Class: "barbarian", Level: 100, ParagonLevel: 200},
		},
		Season: model.SeasonData{SeasonNumber: 4, JourneyProgress: 75},
	}

	if stats.Username != want.Username {
		t.Errorf("Username = %q, want %q", stats.Username, want.Username)
	}
	if stats.Platform != want.Platform {
		t.Errorf("Platform = %q, want %q", stats.Platform, want.Platform)
	}
	if len(stats.Characters) != 1 || stats.Characters[0] != want.Characters[0] {
		t.Errorf("Characters = %+v, want %+v", stats.Characters, want.Characters)
	}
	if stats.Season != want.Season {
		t.Errorf("Season = %+v, want %+v", stats.Season, want.Season)
	}
}

func TestFetchStats_TokenReusedWithinTTL(t *testing.T) {
	var tokenCalls int
	srv := newTestServer(t, &tokenCalls, []int{200, 200}, statsFixture)
	defer srv.Close()

	client := battlenet.NewClient(srv.URL, srv.URL, "client-id", "client-secret")
	client.FetchStats(context.Background(), "Hero#1234")
	client.FetchStats(context.Background(), "Hero#1234")

	if tokenCalls != 1 {
		t.Errorf("token fetched %d times, want 1 (should be cached)", tokenCalls)
	}
}

func TestFetchStats_TokenRefreshedOn401(t *testing.T) {
	var tokenCalls int
	// First stats call returns 401, second (after token refresh) returns 200.
	srv := newTestServer(t, &tokenCalls, []int{401, 200}, statsFixture)
	defer srv.Close()

	client := battlenet.NewClient(srv.URL, srv.URL, "client-id", "client-secret")
	_, err := client.FetchStats(context.Background(), "Hero#1234")
	if err != nil {
		t.Fatalf("FetchStats error: %v", err)
	}
	if tokenCalls != 2 {
		t.Errorf("token fetched %d times, want 2 (initial + refresh on 401)", tokenCalls)
	}
}

func TestFetchStats_RateLimited(t *testing.T) {
	var tokenCalls int
	srv := newTestServer(t, &tokenCalls, []int{429}, "")
	defer srv.Close()

	client := battlenet.NewClient(srv.URL, srv.URL, "client-id", "client-secret")
	_, err := client.FetchStats(context.Background(), "Hero#1234")
	if !errors.Is(err, platform.ErrPlatformRateLimited) {
		t.Errorf("err = %v, want ErrPlatformRateLimited", err)
	}
}
