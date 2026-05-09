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
	return &unsignedIssuer{issuer: cfg.Issuer}, nil
}

type SymmIssuerConfig struct {
	Issuer string
	Key    []byte
}

func NewSymmIssuer(cfg SymmIssuerConfig) (Issuer, error) {
	if len(cfg.Key) == 0 {
		return nil, fmt.Errorf("key must not be empty")
	}
	return &symmIssuer{issuer: cfg.Issuer, key: cfg.Key}, nil
}

type AsymmIssuerConfig struct {
	Issuer string
	Key    *rsa.PrivateKey
}

func NewAsymmIssuer(cfg AsymmIssuerConfig) (Issuer, error) {
	if cfg.Key == nil {
		return nil, fmt.Errorf("private key must not be nil")
	}
	return &asymmIssuer{issuer: cfg.Issuer, key: cfg.Key}, nil
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
		return nil, fmt.Errorf("key must not be empty")
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
		return nil, fmt.Errorf("public key must not be nil")
	}
	return &asymmVerifier{key: cfg.Key}, nil
}
