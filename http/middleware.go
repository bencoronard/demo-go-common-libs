package http

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bencoronard/demo-go-common-libs/constant"
)

func ApiKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			key := strings.TrimSpace(r.Header.Get(constant.MSG_HEADER_API_KEY))
			if key == "" {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]any{
					"type":   "about:blank",
					"title":  "Unauthorized",
					"status": 401,
					"detail": "Missing API key",
				})
				return
			}

			if subtle.ConstantTimeCompare([]byte(key), []byte(apiKey)) != 1 {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "invalid API key"}`))
				json.NewEncoder(w).Encode(map[string]any{
					"type":   "about:blank",
					"title":  "Unauthorized",
					"status": 401,
					"detail": "Invalid API key",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
