package jwt

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type unsignedVerifier struct{}

func (v *unsignedVerifier) VerifyToken(token string) (jwt.MapClaims, error) {
	return verifyToken(token, func(t *jwt.Token) (any, error) {
		return jwt.UnsafeAllowNoneSignatureType, nil
	})
}

type symmVerifier struct {
	key []byte
}

func (v *symmVerifier) VerifyToken(token string) (jwt.MapClaims, error) {
	return verifyToken(token, func(t *jwt.Token) (any, error) {
		return v.key, nil
	})
}

type asymmVerifier struct {
	key *rsa.PublicKey
}

func (v *asymmVerifier) VerifyToken(token string) (jwt.MapClaims, error) {
	return verifyToken(token, func(t *jwt.Token) (any, error) {
		return v.key, nil
	})
}

func verifyToken(token string, kf jwt.Keyfunc) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	if _, err := jwt.ParseWithClaims(token, claims, kf); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return claims, nil
}

func verify(token string, kf jwt.Keyfunc) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{
			jwt.SigningMethodNone.Alg(),
			jwt.SigningMethodHS256.Alg(),
			jwt.SigningMethodRS256.Alg(),
		}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(""),
		jwt.WithAllAudiences(""),
	}

	parser := jwt.NewParser(opts...)
	_, err := parser.ParseWithClaims(token, claims, kf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return claims, nil
}
