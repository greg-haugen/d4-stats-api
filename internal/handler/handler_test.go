package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ghaugen/d4-stats-api/internal/handler"
	"github.com/ghaugen/d4-stats-api/internal/model"
	"github.com/ghaugen/d4-stats-api/internal/platform"
)

// mockClient implements platform.Client for testing.
type mockClient struct {
	stats *model.PlayerStats
	err   error
}

func (m *mockClient) FetchStats(_ context.Context, _ string) (*model.PlayerStats, error) {
	return m.stats, m.err
}

var successStats = &model.PlayerStats{
	Username: "Hero#1234",
	Platform: "battlenet",
	Characters: []model.Character{
		{Name: "Draven", Class: "barbarian", Level: 100, ParagonLevel: 200},
	},
	Season: model.SeasonData{SeasonNumber: 4, JourneyProgress: 75},
}

func buildRouter(clients map[string]platform.Client) http.Handler {
	return handler.NewRouter(clients)
}

func TestStatsHandler_Success(t *testing.T) {
	clients := map[string]platform.Client{
		"battlenet": &mockClient{stats: successStats},
	}
	req := httptest.NewRequest(http.MethodGet, "/stats/battlenet/Hero%231234", nil)
	w := httptest.NewRecorder()

	buildRouter(clients).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var got model.PlayerStats
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got.Username != successStats.Username {
		t.Errorf("Username = %q, want %q", got.Username, successStats.Username)
	}
	if got.Platform != successStats.Platform {
		t.Errorf("Platform = %q, want %q", got.Platform, successStats.Platform)
	}
}

func TestStatsHandler_UnknownPlatform(t *testing.T) {
	clients := map[string]platform.Client{}
	req := httptest.NewRequest(http.MethodGet, "/stats/nintendo/player1", nil)
	w := httptest.NewRecorder()

	buildRouter(clients).ServeHTTP(w, req)

	assertErrorResponse(t, w, http.StatusBadRequest, "INVALID_PLATFORM")
}

func TestStatsHandler_PlayerNotFound(t *testing.T) {
	clients := map[string]platform.Client{
		"steam": &mockClient{err: platform.ErrPlayerNotFound},
	}
	req := httptest.NewRequest(http.MethodGet, "/stats/steam/ghost", nil)
	w := httptest.NewRecorder()

	buildRouter(clients).ServeHTTP(w, req)

	assertErrorResponse(t, w, http.StatusNotFound, "PLAYER_NOT_FOUND")
}

func TestStatsHandler_PSNTokenInvalid(t *testing.T) {
	clients := map[string]platform.Client{
		"psn": &mockClient{err: platform.ErrPSNTokenInvalid},
	}
	req := httptest.NewRequest(http.MethodGet, "/stats/psn/player", nil)
	w := httptest.NewRecorder()

	buildRouter(clients).ServeHTTP(w, req)

	assertErrorResponse(t, w, http.StatusBadGateway, "PSN_TOKEN_INVALID")
}

func TestStatsHandler_PlatformRateLimited(t *testing.T) {
	clients := map[string]platform.Client{
		"xbox": &mockClient{err: platform.ErrPlatformRateLimited},
	}
	req := httptest.NewRequest(http.MethodGet, "/stats/xbox/player", nil)
	w := httptest.NewRecorder()

	buildRouter(clients).ServeHTTP(w, req)

	assertErrorResponse(t, w, http.StatusBadGateway, "PLATFORM_RATE_LIMITED")
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	buildRouter(map[string]platform.Client{}).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("status = %q, want %q", body["status"], "ok")
	}
}

func TestStatsHandler_ResponseHasAllFields(t *testing.T) {
	statsWithEmptyChars := &model.PlayerStats{
		Username:   "Empty",
		Platform:   "steam",
		Characters: []model.Character{},
		Season:     model.SeasonData{SeasonNumber: 1},
	}
	clients := map[string]platform.Client{
		"steam": &mockClient{stats: statsWithEmptyChars},
	}
	req := httptest.NewRequest(http.MethodGet, "/stats/steam/Empty", nil)
	w := httptest.NewRecorder()

	buildRouter(clients).ServeHTTP(w, req)

	var m map[string]json.RawMessage
	json.NewDecoder(w.Body).Decode(&m)

	for _, key := range []string{"username", "platform", "characters", "season"} {
		if _, ok := m[key]; !ok {
			t.Errorf("response missing key %q", key)
		}
	}
	if string(m["characters"]) == "null" {
		t.Error("characters is null, want []")
	}
}

func assertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()
	if w.Code != wantStatus {
		t.Errorf("status = %d, want %d", w.Code, wantStatus)
	}
	var body struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}
	if body.Error == "" {
		t.Error("error field is empty")
	}
	if body.Code != wantCode {
		t.Errorf("code = %q, want %q", body.Code, wantCode)
	}
	_ = errors.New("") // keep errors import alive
}
