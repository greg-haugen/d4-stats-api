package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ghaugen/d4-stats-api/internal/middleware"
)

func TestRateLimit_AllowsWithinLimit(t *testing.T) {
	handler := middleware.RateLimit(10)(http.HandlerFunc(okHandler))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/stats/steam/player", nil)
		req.RemoteAddr = "1.2.3.4:9999"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i+1, w.Code)
		}
	}
}

func TestRateLimit_RejectsExcess(t *testing.T) {
	// Limit of 1 RPS: burst=1, then immediately fire many requests from same IP.
	handler := middleware.RateLimit(1)(http.HandlerFunc(okHandler))

	got429 := false
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodGet, "/stats/steam/player", nil)
		req.RemoteAddr = "5.6.7.8:9999"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			got429 = true
			var body struct {
				Code string `json:"code"`
			}
			decodeJSON(t, w, &body)
			if body.Code != "RATE_LIMITED" {
				t.Errorf("code = %q, want RATE_LIMITED", body.Code)
			}
			break
		}
	}
	if !got429 {
		t.Error("expected at least one 429, got none")
	}
}

func TestRateLimit_PerSourceIP(t *testing.T) {
	// Low limit: IP A consumes it; IP B should still get through.
	handler := middleware.RateLimit(1)(http.HandlerFunc(okHandler))

	// Exhaust IP A's bucket.
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodGet, "/stats/steam/player", nil)
		req.RemoteAddr = "10.0.0.1:9999"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// IP B should still be allowed.
	req := httptest.NewRequest(http.MethodGet, "/stats/steam/player", nil)
	req.RemoteAddr = "10.0.0.2:9999"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code == http.StatusTooManyRequests {
		t.Error("IP B rate-limited by IP A's usage; limits must be per-IP")
	}
}
