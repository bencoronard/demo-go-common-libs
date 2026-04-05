package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer interface {
	IssueToken(sub string, aud []string, claims map[string]any, ttl time.Duration, nbf time.Time) (string, error)
}

type UnsignedIssuerConfig struct {
	Issuer string
}

func NewUnsignedIssuer(cfg UnsignedIssuerConfig) (Issuer, error) {
	return &unsignedIssuer{iss: cfg.Issuer}, nil
}

type SymmIssuerConfig struct {
	Issuer string
	Key    []byte
}

func NewSymmIssuer(cfg SymmIssuerConfig) (Issuer, error) {
	if len(cfg.Key) == 0 {
		return nil, fmt.Errorf("%w: key must not be empty", ErrConstructInstanceFail)
	}
	return &symmIssuer{iss: cfg.Issuer, key: cfg.Key}, nil
}

type AsymmIssuerConfig struct {
	Issuer string
	Key    *rsa.PrivateKey
}

func NewAsymmIssuer(cfg AsymmIssuerConfig) (Issuer, error) {
	if cfg.Key == nil {
		return nil, fmt.Errorf("%w: private key must not be nil", ErrConstructInstanceFail)
	}
	return &asymmIssuer{iss: cfg.Issuer, key: cfg.Key}, nil
}

type Verifier interface {
	VerifyToken(tokenStr string) (jwt.MapClaims, error)
}

func NewUnsignedVerifier() (Verifier, error) {
	return &unsignedVerifier{}, nil
}

type SymmVerifierConfig struct {
	Key []byte
}

func NewSymmVerifier(cfg SymmVerifierConfig) (Verifier, error) {
	if len(cfg.Key) == 0 {
		return nil, fmt.Errorf("%w: key must not be empty", ErrConstructInstanceFail)
	}

	keyCopy := make([]byte, len(cfg.Key))
	copy(keyCopy, cfg.Key)

	return &symmVerifier{key: keyCopy}, nil
}

type AsymmVerifierConfig struct {
	Key *rsa.PublicKey
}

func NewAsymmVerifier(cfg AsymmVerifierConfig) (Verifier, error) {
	if cfg.Key == nil {
		return nil, fmt.Errorf("%w: public key must not be nil", ErrConstructInstanceFail)
	}
	return &asymmVerifier{key: cfg.Key}, nil
}
