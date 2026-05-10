package jwt

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type unsignedVerifier struct {
	parser *jwt.Parser
}

func (v *unsignedVerifier) VerifyToken(token string) (jwt.MapClaims, error) {
	return verifyToken(v.parser, token, func(t *jwt.Token) (any, error) {
		return jwt.UnsafeAllowNoneSignatureType, nil
	})
}

type symmVerifier struct {
	parser *jwt.Parser
	key    []byte
}

func (v *symmVerifier) VerifyToken(token string) (jwt.MapClaims, error) {
	return verifyToken(v.parser, token, func(t *jwt.Token) (any, error) {
		return v.key, nil
	})
}

type asymmVerifier struct {
	parser *jwt.Parser
	key    *rsa.PublicKey
}

func (v *asymmVerifier) VerifyToken(token string) (jwt.MapClaims, error) {
	return verifyToken(v.parser, token, func(t *jwt.Token) (any, error) {
		return v.key, nil
	})
}

func verifyToken(p *jwt.Parser, token string, kf jwt.Keyfunc) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	if _, err := p.ParseWithClaims(token, claims, kf); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return claims, nil
}
