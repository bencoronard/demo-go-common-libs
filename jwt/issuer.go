package jwt

import (
	"crypto/rsa"
	"fmt"
	"maps"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Issuer interface {
	IssueToken(sub string, aud []string, claims map[string]any, ttl time.Duration, nbf time.Time) (string, error)
}

type unsignedIssuer struct {
	iss string
}

type symmIssuer struct {
	iss string
	key []byte
}

type asymmIssuer struct {
	iss string
	key *rsa.PrivateKey
}

func NewUnsignedIssuer(iss string) (Issuer, error) {
	return &unsignedIssuer{iss: iss}, nil
}

func NewSymmIssuer(iss string, key []byte) (Issuer, error) {
	if key == nil {
		return nil, fmt.Errorf("%w: key must not be nil", ErrConstructInstanceFail)
	}
	return &symmIssuer{iss: iss, key: key}, nil
}

func NewAsymmIssuer(iss string, key *rsa.PrivateKey) (Issuer, error) {
	if key == nil {
		return nil, fmt.Errorf("%w: private key must not be nil", ErrConstructInstanceFail)
	}
	return &asymmIssuer{iss: iss, key: key}, nil
}

func (i *unsignedIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl time.Duration, nbf time.Time) (string, error) {
	mc, err := buildClaims(i.iss, sub, aud, claims, ttl, nbf)
	if err != nil {
		return "", err
	}
	return issueToken(jwt.SigningMethodNone, mc, jwt.UnsafeAllowNoneSignatureType)
}

func (i *symmIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl time.Duration, nbf time.Time) (string, error) {
	mc, err := buildClaims(i.iss, sub, aud, claims, ttl, nbf)
	if err != nil {
		return "", err
	}
	return issueToken(jwt.SigningMethodHS256, mc, i.key)
}

func (i *asymmIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl time.Duration, nbf time.Time) (string, error) {
	mc, err := buildClaims(i.iss, sub, aud, claims, ttl, nbf)
	if err != nil {
		return "", err
	}
	return issueToken(jwt.SigningMethodRS256, mc, i.key)
}

func buildClaims(iss string, sub string, aud []string, claims map[string]any, ttl time.Duration, nbf time.Time) (jwt.MapClaims, error) {
	now := time.Now()
	mc := jwt.MapClaims{}

	if !nbf.IsZero() {
		mc["nbf"] = jwt.NewNumericDate(nbf)
	}

	if ttl > 0 {
		mc["exp"] = jwt.NewNumericDate(now.Add(ttl))
	}

	rand, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate token UUID: %v", ErrTokenIssuanceFail, err)
	}

	if sub != "" {
		mc["sub"] = sub
	}

	mc["iss"] = iss
	mc["jti"] = rand.String()
	mc["iat"] = jwt.NewNumericDate(now)
	mc["aud"] = aud
	maps.Copy(mc, claims)

	return mc, nil
}

func issueToken(method jwt.SigningMethod, claims jwt.MapClaims, key any) (string, error) {
	token := jwt.NewWithClaims(method, claims)
	tokenStr, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("%w: failed to sign token: %v", ErrTokenIssuanceFail, err)
	}
	return tokenStr, nil
}
