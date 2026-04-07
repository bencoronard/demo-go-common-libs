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
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return nil, fmt.Errorf("%w: missing Authorization header", ErrMissingRequestHeader)
	}

	claims, err := h.verifier.VerifyToken(header[len("Bearer "):])
	if err != nil {
		return nil, err
	}

	sub, _ := claims["sub"].(string)
	if sub == "" || !regexp.MustCompile(`^\d+$`).MatchString(sub) {
		return nil, fmt.Errorf("%w: expected integer, got %q", xjwt.ErrTokenMalformed, sub)
	}

	return claims, nil
}
