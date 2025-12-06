package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtIssuer interface {
	IssueToken(iss string, sub string, ttl time.Duration) (string, error)
}

type jwtIssuerImpl struct {
	key []byte
}

func NewJwtIssuer(secret string) JwtIssuer {
	return &jwtIssuerImpl{
		key: []byte(secret),
	}
}

func (j *jwtIssuerImpl) IssueToken(iss string, sub string, ttl time.Duration) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Issuer:    iss,
		Subject:   sub,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(j.key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenStr, nil
}
