package psn_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ghaugen/d4-stats-api/internal/model"
	"github.com/ghaugen/d4-stats-api/internal/platform"
	"github.com/ghaugen/d4-stats-api/internal/platform/psn"
)

const statsFixture = `{
	"onlineId": "PSNPlayer",
	"characters": [
		{"name": "Lyra", "class": "necromancer", "level": 90, "paragonLevel": 120}
	],
	"season": {"seasonNumber": 4, "journeyProgress": 55}
}`

func newPSNServer(t *testing.T, authHeaders *[]string, statusCode int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*authHeaders = append(*authHeaders, r.Header.Get("Authorization"))
		w.WriteHeader(statusCode)
		if statusCode == http.StatusOK {
			w.Write([]byte(body))
		}
	}))
}

func TestFetchStats_Success(t *testing.T) {
	var headers []string
	srv := newPSNServer(t, &headers, http.StatusOK, statsFixture)
	defer srv.Close()

	client := psn.NewClient(srv.URL, "psn-npsso-token")
	stats, err := client.FetchStats(context.Background(), "PSNPlayer")
	if err != nil {
		t.Fatalf("FetchStats error: %v", err)
	}

	want := &model.PlayerStats{
		Username: "PSNPlayer",
		Platform: "psn",
		Characters: []model.Character{
			{Name: "Lyra", Class: "necromancer", Level: 90, ParagonLevel: 120},
		},
		Season: model.SeasonData{SeasonNumber: 4, JourneyProgress: 55},
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

func TestFetchStats_BearerTokenPresent(t *testing.T) {
	var headers []string
	srv := newPSNServer(t, &headers, http.StatusOK, statsFixture)
	defer srv.Close()

	client := psn.NewClient(srv.URL, "psn-npsso-token")
	client.FetchStats(context.Background(), "PSNPlayer")

	if len(headers) == 0 || headers[0] != "Bearer psn-npsso-token" {
		t.Errorf("Authorization header = %v, want [Bearer psn-npsso-token]", headers)
	}
}

func TestFetchStats_ExpiredToken(t *testing.T) {
	var headers []string
	srv := newPSNServer(t, &headers, http.StatusUnauthorized, "")
	defer srv.Close()

	client := psn.NewClient(srv.URL, "expired-token")
	_, err := client.FetchStats(context.Background(), "PSNPlayer")
	if !errors.Is(err, platform.ErrPSNTokenInvalid) {
		t.Errorf("err = %v, want ErrPSNTokenInvalid", err)
	}
}

func TestFetchStats_PlayerNotFound(t *testing.T) {
	var headers []string
	srv := newPSNServer(t, &headers, http.StatusNotFound, "")
	defer srv.Close()

	client := psn.NewClient(srv.URL, "valid-token")
	_, err := client.FetchStats(context.Background(), "ghost")
	if !errors.Is(err, platform.ErrPlayerNotFound) {
		t.Errorf("err = %v, want ErrPlayerNotFound", err)
	}
}
