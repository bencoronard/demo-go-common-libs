package http

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	xjwt "github.com/bencoronard/demo-go-common-libs/jwt"
	"github.com/golang-jwt/jwt/v5"
)

type authHeaderResolver struct {
	verifier xjwt.Verifier
}

func (h *authHeaderResolver) ExtractClaims(r *http.Request) (jwt.MapClaims, error) {
	prefix := "Bearer "
	header := strings.TrimSpace(r.Header.Get("Authorization"))

	if !strings.HasPrefix(header, prefix) {
		return nil, ErrAuthHeaderInvalid
	}

	claims, err := h.verifier.VerifyToken(header[len(prefix):])
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	sub, _ := claims["sub"].(string)
	if sub == "" || !regexp.MustCompile(`^\d+$`).MatchString(sub) {
		return nil, fmt.Errorf("expected integer, got %q", sub)
	}

	return claims, nil
}
