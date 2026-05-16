package http

import (
	"errors"
	"fmt"
	"net/http"
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

	return claims, nil
}
