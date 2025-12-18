package http

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bencoronard/demo-go-common-libs/constant"
	"github.com/bencoronard/demo-go-common-libs/rfc9457"
)

func ApiKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			key := strings.TrimSpace(r.Header.Get(constant.MSG_HEADER_API_KEY))
			if key == "" {
				sendErrorResponse(w, http.StatusUnauthorized, "Missing API key")
				return
			}

			if subtle.ConstantTimeCompare([]byte(key), []byte(apiKey)) != 1 {
				sendErrorResponse(w, http.StatusUnauthorized, "Invalid API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func sendErrorResponse(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.Encode(rfc9457.ForStatusAndDetail(status, msg))
}
