package http

import (
	xjwt "github.com/bencoronard/demo-go-common-libs/jwt"
	"github.com/golang-jwt/jwt/v5"
)

type AuthTokenResolver interface {
	ResolveToken(tokenStr string) (jwt.MapClaims, error)
}

type authTokenResolverImpl struct {
	verifier xjwt.JwtVerifier
}

func NewAuthTokenResolver(verifier xjwt.JwtVerifier) (AuthTokenResolver, error) {
	return &authTokenResolverImpl{
		verifier: verifier,
	}, nil
}

func (r *authTokenResolverImpl) ResolveToken(tokenStr string) (jwt.MapClaims, error) {
	return r.verifier.VerifyToken(tokenStr)
}
