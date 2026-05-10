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
	UnsignedIssuerConfig
	Key []byte
}

func NewSymmIssuer(cfg SymmIssuerConfig) (Issuer, error) {
	if len(cfg.Key) == 0 {
		return nil, fmt.Errorf("key must not be empty")
	}
	return &symmIssuer{issuer: cfg.Issuer, key: cfg.Key}, nil
}

type AsymmIssuerConfig struct {
	UnsignedIssuerConfig
	Key *rsa.PrivateKey
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

type UnsignedVerifierConfig struct {
	TrustedIssuer     string
	RequiredAudiences []string
}

func NewUnsignedVerifier(cfg UnsignedVerifierConfig) (Verifier, error) {
	opts := []jwt.ParserOption{
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{
			jwt.SigningMethodNone.Alg(),
		}),
	}

	if cfg.TrustedIssuer != "" {
		opts = append(opts, jwt.WithIssuer(cfg.TrustedIssuer))
	}

	if len(cfg.RequiredAudiences) > 0 {
		opts = append(opts, jwt.WithAllAudiences(cfg.RequiredAudiences...))
	}

	return &unsignedVerifier{
		parser: jwt.NewParser(opts...),
	}, nil
}

type SymmVerifierConfig struct {
	UnsignedVerifierConfig
	Key []byte
}

func NewSymmVerifier(cfg SymmVerifierConfig) (Verifier, error) {
	if len(cfg.Key) == 0 {
		return nil, fmt.Errorf("key must not be empty")
	}

	keyCopy := make([]byte, len(cfg.Key))
	copy(keyCopy, cfg.Key)

	opts := []jwt.ParserOption{
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
	}

	if cfg.TrustedIssuer != "" {
		opts = append(opts, jwt.WithIssuer(cfg.TrustedIssuer))
	}

	if len(cfg.RequiredAudiences) > 0 {
		opts = append(opts, jwt.WithAllAudiences(cfg.RequiredAudiences...))
	}

	return &symmVerifier{
		parser: jwt.NewParser(opts...),
		key:    keyCopy,
	}, nil
}

type AsymmVerifierConfig struct {
	UnsignedVerifierConfig
	Key *rsa.PublicKey
}

func NewAsymmVerifier(cfg AsymmVerifierConfig) (Verifier, error) {
	if cfg.Key == nil {
		return nil, fmt.Errorf("public key must not be nil")
	}

	opts := []jwt.ParserOption{
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{
			jwt.SigningMethodRS256.Alg(),
		}),
	}

	if cfg.TrustedIssuer != "" {
		opts = append(opts, jwt.WithIssuer(cfg.TrustedIssuer))
	}

	if len(cfg.RequiredAudiences) > 0 {
		opts = append(opts, jwt.WithAllAudiences(cfg.RequiredAudiences...))
	}

	return &asymmVerifier{
		parser: jwt.NewParser(opts...),
		key:    cfg.Key,
	}, nil
}
