package handler

import (
	"net/http"

	"github.com/ghaugen/d4-stats-api/internal/platform"
)

// NewRouter builds the mux. Middleware is applied by the caller in main.go;
// routes.go only owns path registration so tests can use it directly.
func NewRouter(clients map[string]platform.Client) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /stats/{platform}/{username}", statsHandler(clients))
	return mux
}
