package middleware

import (
	"net/http"
	"strings"

	"github.com/bisheshops/dynamic-crm-engine/internal/response"
)

func RequireAPIKey(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			providedKey := r.Header.Get("X-API-Key")

			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					providedKey = parts[1]
				}
			}

			if providedKey == "" || providedKey != expectedKey {
				response.Error(w, http.StatusUnauthorized, "Unauthorized: Invalid or missing API key")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
