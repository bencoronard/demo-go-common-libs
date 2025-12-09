package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
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

	mc := jwt.MapClaims{}

	if nbf != nil {
		eff = *nbf
		mc["nbf"] = jwt.NewNumericDate(eff)
	}

	if ttl != nil {
		exp := now.Add(*ttl)
		if exp.Before(now) {
			return "", errors.New("token expiration cannot be in the past")
		}
		mc["exp"] = jwt.NewNumericDate(exp)
	}

	rand, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	mc["iss"] = i.iss
	mc["jti"] = rand.String()
	mc["sub"] = sub
	mc["iat"] = jwt.NewNumericDate(now)
	mc["aud"] = aud
	maps.Copy(mc, claims)

	token := jwt.NewWithClaims(jwt.SigningMethodNone, mc)

	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", fmt.Errorf("failed to issue unsigned token: %w", err)
	}

	return tokenStr, nil
}

func (i *symmJwtIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl *time.Duration, nbf *time.Time) (string, error) {
	now := time.Now()
	eff := now

	mc := jwt.MapClaims{}

	if nbf != nil {
		eff = *nbf
		mc["nbf"] = jwt.NewNumericDate(eff)
	}

	if ttl != nil {
		exp := now.Add(*ttl)
		if exp.Before(now) {
			return "", errors.New("token expiration cannot be in the past")
		}
		mc["exp"] = jwt.NewNumericDate(exp)
	}

	rand, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	mc["iss"] = i.iss
	mc["jti"] = rand.String()
	mc["sub"] = sub
	mc["iat"] = jwt.NewNumericDate(now)
	mc["aud"] = aud
	maps.Copy(mc, claims)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mc)

	tokenStr, err := token.SignedString(i.key)
	if err != nil {
		return "", fmt.Errorf("failed to issue symmetrically signed token: %w", err)
	}

	return tokenStr, nil
}

func (i *asymmJwtIssuer) IssueToken(sub string, aud []string, claims map[string]any, ttl *time.Duration, nbf *time.Time) (string, error) {
	now := time.Now()
	eff := now

	mc := jwt.MapClaims{}

	if nbf != nil {
		eff = *nbf
		mc["nbf"] = jwt.NewNumericDate(eff)
	}

	if ttl != nil {
		exp := now.Add(*ttl)
		if exp.Before(now) {
			return "", errors.New("token expiration cannot be in the past")
		}
		mc["exp"] = jwt.NewNumericDate(exp)
	}

	rand, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	mc["iss"] = i.iss
	mc["jti"] = rand.String()
	mc["sub"] = sub
	mc["iat"] = jwt.NewNumericDate(now)
	mc["aud"] = aud
	maps.Copy(mc, claims)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, mc)

	tokenStr, err := token.SignedString(i.key)
	if err != nil {
		return "", fmt.Errorf("failed to issue symmetrically signed token: %w", err)
	}

	return tokenStr, nil
}
