package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ghaugen/d4-stats-api/internal/platform"
)

type errResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func writeError(w http.ResponseWriter, status int, msg, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errResponse{Error: msg, Code: code})
}

func statsHandler(clients map[string]platform.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		platformName := r.PathValue("platform")
		username := r.PathValue("username")

		client, ok := clients[platformName]
		if !ok {
			writeError(w, http.StatusBadRequest, "unsupported platform", "INVALID_PLATFORM")
			return
		}

		stats, err := client.FetchStats(r.Context(), username)
		if err != nil {
			switch {
			case errors.Is(err, platform.ErrPlayerNotFound):
				writeError(w, http.StatusNotFound, "player not found", "PLAYER_NOT_FOUND")
			case errors.Is(err, platform.ErrPSNTokenInvalid):
				writeError(w, http.StatusBadGateway, "platform authentication expired", "PSN_TOKEN_INVALID")
			case errors.Is(err, platform.ErrPlatformRateLimited):
				writeError(w, http.StatusBadGateway, "platform rate limited", "PLATFORM_RATE_LIMITED")
			default:
				writeError(w, http.StatusInternalServerError, "internal error", "INTERNAL_ERROR")
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
