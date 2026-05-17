package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

func RateLimit(rps int) func(http.Handler) http.Handler {
	type entry struct{ limiter *rate.Limiter }
	var mu sync.Mutex
	limiters := make(map[string]*entry)

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		e, ok := limiters[ip]
		if !ok {
			e = &entry{limiter: rate.NewLimiter(rate.Limit(rps), rps)}
			limiters[ip] = e
		}
		return e.limiter
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			if !getLimiter(ip).Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(struct {
					Error string `json:"error"`
					Code  string `json:"code"`
				}{Error: "rate limit exceeded", Code: "RATE_LIMITED"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
