package jwt

import (
	"crypto/rsa"
	"errors"
	"maps"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtIssuer interface {
	IssueToken(sub string, aud []string, claims map[string]any, ttl *time.Duration, nbf *time.Time) (string, error)
}

type unsignedJwtIssuer struct {
	iss string
}

type symmJwtIssuer struct {
	iss string
	key []byte
}

type asymmJwtIssuer struct {
	iss string
	key *rsa.PrivateKey
}

func NewUnsignedJwtIssuer(iss string) (JwtIssuer, error) {
	return &unsignedJwtIssuer{iss: iss}, nil
}

func NewSymmJwtIssuer(iss string, key []byte) (JwtIssuer, error) {
	if key == nil {
		return nil, errors.New("symmetric key must not be nil")
	}
	return &symmJwtIssuer{iss: iss, key: key}, nil
}

func NewAsymmJwtIssuer(iss string, key *rsa.PrivateKey) (JwtIssuer, error) {
	if key == nil {
		return nil, errors.New("private key must not be nil")
	}
	return &asymmJwtIssuer{iss: iss, key: key}, nil
}

func (i *unsignedJwtIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl *time.Duration, nbf *time.Time) (string, error) {
	mc, err := buildClaims(i.iss, sub, aud, claims, ttl, nbf)
	if err != nil {
		return "", err
	}
	return issueToken(jwt.SigningMethodNone, mc, jwt.UnsafeAllowNoneSignatureType)
}

func (i *symmJwtIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl *time.Duration, nbf *time.Time) (string, error) {
	mc, err := buildClaims(i.iss, sub, aud, claims, ttl, nbf)
	if err != nil {
		return "", err
	}
	return issueToken(jwt.SigningMethodHS256, mc, i.key)
}

func (i *asymmJwtIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl *time.Duration, nbf *time.Time) (string, error) {
	mc, err := buildClaims(i.iss, sub, aud, claims, ttl, nbf)
	if err != nil {
		return "", err
	}
	return issueToken(jwt.SigningMethodRS256, mc, i.key)
}

func buildClaims(iss, sub string, aud []string, claims map[string]any, ttl *time.Duration, nbf *time.Time) (jwt.MapClaims, error) {
	now := time.Now()
	mc := jwt.MapClaims{}

	if nbf != nil {
		mc["nbf"] = jwt.NewNumericDate(*nbf)
	}

	if ttl != nil {
		exp := now.Add(*ttl)
		if exp.Before(now) {
			return nil, ErrExpiredTokenIssueAttempted
		}
		mc["exp"] = jwt.NewNumericDate(exp)
	}

	rand, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	mc["iss"] = iss
	mc["jti"] = rand.String()
	mc["sub"] = sub
	mc["iat"] = jwt.NewNumericDate(now)
	mc["aud"] = aud
	maps.Copy(mc, claims)

	return mc, nil
}

func issueToken(method jwt.SigningMethod, claims jwt.MapClaims, key any) (string, error) {
	token := jwt.NewWithClaims(method, claims)
	tokenStr, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}
