package http

import (
	"net/http"
	"regexp"
	"strings"

	xjwt "github.com/bencoronard/demo-go-common-libs/jwt"
	"github.com/golang-jwt/jwt/v5"
)

type HttpAuthHeaderResolver interface {
	ExtractClaims(r *http.Request) (jwt.MapClaims, error)
}

type httpAuthHeaderResolverImpl struct {
	verifier xjwt.JwtVerifier
}

func NewHttpAuthHeaderResolver(verifier xjwt.JwtVerifier) HttpAuthHeaderResolver {
	return &httpAuthHeaderResolverImpl{verifier: verifier}
}

func (h *httpAuthHeaderResolverImpl) ExtractClaims(r *http.Request) (jwt.MapClaims, error) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return nil, ErrMissingRequestHeader
	}

	claims, err := h.verifier.VerifyToken(header[len("Bearer "):])
	if err != nil {
		return nil, err
	}

	sub, _ := claims["sub"].(string)
	if sub == "" || !regexp.MustCompile(`^\d+$`).MatchString(sub) {
		return nil, xjwt.ErrTokenMalformed
	}

	return claims, nil
}
