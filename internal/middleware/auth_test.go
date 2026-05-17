package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ghaugen/d4-stats-api/internal/middleware"
)

const testSecret = "test-jwt-secret"

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func makeToken(secret string, exp time.Time) string {
	claims := jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(exp)}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return tok
}

func TestAuth_ValidToken(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	token := makeToken(testSecret, time.Now().Add(time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/stats/steam/player", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestAuth_MissingToken(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/stats/steam/player", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assertAuthError(t, w, http.StatusUnauthorized, "MISSING_TOKEN")
}

func TestAuth_ExpiredToken(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	token := makeToken(testSecret, time.Now().Add(-time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/stats/steam/player", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assertAuthError(t, w, http.StatusUnauthorized, "TOKEN_EXPIRED")
}

func TestAuth_MalformedToken(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/stats/steam/player", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assertAuthError(t, w, http.StatusUnauthorized, "INVALID_TOKEN")
}

func TestAuth_HealthBypassesAuth(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("health check status = %d, want 200 (should bypass auth)", w.Code)
	}
}

func assertAuthError(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()
	if w.Code != wantStatus {
		t.Errorf("status = %d, want %d", w.Code, wantStatus)
	}
	var body struct {
		Code string `json:"code"`
	}
	json.NewDecoder(w.Body).Decode(&body)
	if body.Code != wantCode {
		t.Errorf("code = %q, want %q", body.Code, wantCode)
	}
}
