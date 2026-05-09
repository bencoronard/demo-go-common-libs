package jwt

import (
	"crypto/rsa"
	"fmt"
	"maps"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type unsignedIssuer struct {
	issuer string
}

func (i *unsignedIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl time.Duration, nbf time.Time) (string, error) {
	mc, err := buildClaims(i.issuer, sub, aud, claims, ttl, nbf)
	if err != nil {
		return "", err
	}
	return issueToken(jwt.SigningMethodNone, mc, jwt.UnsafeAllowNoneSignatureType)
}

type symmIssuer struct {
	issuer string
	key    []byte
}

func (i *symmIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl time.Duration, nbf time.Time) (string, error) {
	mc, err := buildClaims(i.issuer, sub, aud, claims, ttl, nbf)
	if err != nil {
		return "", err
	}
	return issueToken(jwt.SigningMethodHS256, mc, i.key)
}

type asymmIssuer struct {
	issuer string
	key    *rsa.PrivateKey
}

func (i *asymmIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl time.Duration, nbf time.Time) (string, error) {
	mc, err := buildClaims(i.issuer, sub, aud, claims, ttl, nbf)
	if err != nil {
		return "", err
	}
	return issueToken(jwt.SigningMethodRS256, mc, i.key)
}

func buildClaims(iss string, sub string, aud []string, claims map[string]any, ttl time.Duration, nbf time.Time) (jwt.MapClaims, error) {
	now := time.Now()

	rand, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token's ID: %w", err)
	}

	mc := jwt.MapClaims{}

	mc["iss"] = iss
	mc["jti"] = rand.String()
	mc["iat"] = jwt.NewNumericDate(now)

	if sub != "" {
		mc["sub"] = sub
	}

	if len(aud) > 0 {
		mc["aud"] = aud
	}

	if !nbf.IsZero() {
		mc["nbf"] = jwt.NewNumericDate(nbf)
	}

	if ttl > 0 {
		mc["exp"] = jwt.NewNumericDate(now.Add(ttl))
	}

	maps.Copy(mc, claims)

	return mc, nil
}

func issueToken(method jwt.SigningMethod, claims jwt.MapClaims, key any) (string, error) {
	token := jwt.NewWithClaims(method, claims)
	tokenStr, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenStr, nil
}
