package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
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

func NewUnsignedJwtIssuer(iss string) JwtIssuer {
	return &unsignedJwtIssuer{iss: iss}
}

func NewSymmJwtIssuer(iss string, key []byte) JwtIssuer {
	return &symmJwtIssuer{iss: iss, key: key}
}

func NewAsymmJwtIssuer(iss string, key *rsa.PrivateKey) JwtIssuer {
	return &asymmJwtIssuer{iss: iss, key: key}
}

func (i *unsignedJwtIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl *time.Duration, nbf *time.Time) (string, error) {
	now := time.Now()
	eff := now

	tkn := jwt.RegisteredClaims{}

	if nbf != nil {
		eff = *nbf
		tkn.NotBefore = jwt.NewNumericDate(eff)
	}

	if ttl != nil {
		exp := now.Add(*ttl)
		if exp.Before(now) {
			return "", errors.New("token expiration cannot be in the past")
		}
		tkn.ExpiresAt = jwt.NewNumericDate(exp)
	}

	rand, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	tkn.ID = rand.String()
	tkn.Issuer = i.iss
	tkn.Subject = sub
	tkn.IssuedAt = jwt.NewNumericDate(now)
	tkn.Audience = aud

	token := jwt.NewWithClaims(jwt.SigningMethodNone, tkn)

	jwt, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", fmt.Errorf("failed to issue unsigned token: %w", err)
	}

	return jwt, nil
}

func (i *symmJwtIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl *time.Duration, nbf *time.Time) (string, error) {
	panic("unimplemented")
}

func (i *asymmJwtIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl *time.Duration, nbf *time.Time) (string, error) {
	panic("unimplemented")
}
