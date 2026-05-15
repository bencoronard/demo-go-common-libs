package http

import (
	"errors"
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
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed),
			errors.Is(err, jwt.ErrTokenSignatureInvalid),
			errors.Is(err, jwt.ErrTokenInvalidClaims):
			return nil, fmt.Errorf("%w: %w", ErrTokenInvalid, err)
		default:
			return nil, fmt.Errorf("failed to verify token: %w", err)
		}
	}

	sub, _ := claims["sub"].(string)
	if sub == "" || !regexp.MustCompile(`^\d+$`).MatchString(sub) {
		return nil, fmt.Errorf("%w: `sub` expects integer, got %q", ErrTokenInvalid, sub)
	}

	return claims, nil
}
