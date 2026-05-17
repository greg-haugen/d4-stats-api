package xbox_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ghaugen/d4-stats-api/internal/model"
	"github.com/ghaugen/d4-stats-api/internal/platform"
	"github.com/ghaugen/d4-stats-api/internal/platform/xbox"
)

const oauthTokenFixture = `{"access_token":"ms-token","token_type":"bearer","expires_in":3600}`
const xstsTokenFixture = `{"Token":"xsts-token-abc","NotAfter":"2099-01-01T00:00:00.000Z"}`
const statsFixture = `{
	"gamertag": "XboxPlayer",
	"characters": [
		{"name": "Karn", "class": "druid", "level": 95, "paragonLevel": 180}
	],
	"season": {"seasonNumber": 4, "journeyProgress": 60}
}`

type xboxFixtureServer struct {
	oauthCalls int
	xstsCalls  int
	srv        *httptest.Server
}

func newXboxServer(t *testing.T, statsStatus int, statsBody string) *xboxFixtureServer {
	t.Helper()
	fs := &xboxFixtureServer{}
	fs.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			fs.oauthCalls++
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(json.RawMessage(oauthTokenFixture))
		case "/xsts/token":
			fs.xstsCalls++
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(json.RawMessage(xstsTokenFixture))
		default:
			w.WriteHeader(statsStatus)
			if statsStatus == http.StatusOK {
				w.Write([]byte(statsBody))
			}
		}
	}))
	return fs
}

func TestFetchStats_Success(t *testing.T) {
	fs := newXboxServer(t, http.StatusOK, statsFixture)
	defer fs.srv.Close()

	client := xbox.NewClient(fs.srv.URL, fs.srv.URL, fs.srv.URL, "xbox-id", "xbox-secret")
	stats, err := client.FetchStats(context.Background(), "XboxPlayer")
	if err != nil {
		t.Fatalf("FetchStats error: %v", err)
	}

	want := &model.PlayerStats{
		Username: "XboxPlayer",
		Platform: "xbox",
		Characters: []model.Character{
			{Name: "Karn", Class: "druid", Level: 95, ParagonLevel: 180},
		},
		Season: model.SeasonData{SeasonNumber: 4, JourneyProgress: 60},
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

func TestFetchStats_XSTSTokenCached(t *testing.T) {
	fs := newXboxServer(t, http.StatusOK, statsFixture)
	defer fs.srv.Close()

	client := xbox.NewClient(fs.srv.URL, fs.srv.URL, fs.srv.URL, "xbox-id", "xbox-secret")
	client.FetchStats(context.Background(), "XboxPlayer")
	client.FetchStats(context.Background(), "XboxPlayer")

	if fs.xstsCalls != 1 {
		t.Errorf("XSTS token fetched %d times, want 1 (should be cached)", fs.xstsCalls)
	}
}

func TestFetchStats_RateLimited(t *testing.T) {
	fs := newXboxServer(t, http.StatusTooManyRequests, "")
	defer fs.srv.Close()

	client := xbox.NewClient(fs.srv.URL, fs.srv.URL, fs.srv.URL, "xbox-id", "xbox-secret")
	_, err := client.FetchStats(context.Background(), "XboxPlayer")
	if !errors.Is(err, platform.ErrPlatformRateLimited) {
		t.Errorf("err = %v, want ErrPlatformRateLimited", err)
	}
}

func TestFetchStats_PlayerNotFound(t *testing.T) {
	fs := newXboxServer(t, http.StatusNotFound, "")
	defer fs.srv.Close()

	client := xbox.NewClient(fs.srv.URL, fs.srv.URL, fs.srv.URL, "xbox-id", "xbox-secret")
	_, err := client.FetchStats(context.Background(), "ghost")
	if !errors.Is(err, platform.ErrPlayerNotFound) {
		t.Errorf("err = %v, want ErrPlayerNotFound", err)
	}
}
