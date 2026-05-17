package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				writeAuthError(w, http.StatusUnauthorized, "missing token", "MISSING_TOKEN")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			_, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err == nil {
				next.ServeHTTP(w, r)
				return
			}

			if errors.Is(err, jwt.ErrTokenExpired) {
				writeAuthError(w, http.StatusUnauthorized, "token expired", "TOKEN_EXPIRED")
				return
			}

			writeAuthError(w, http.StatusUnauthorized, "invalid token", "INVALID_TOKEN")
		})
	}
}

func writeAuthError(w http.ResponseWriter, status int, msg, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}{Error: msg, Code: code})
}
