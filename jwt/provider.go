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

func NewUnsignedIssuer(iss string) (Issuer, error) {
	return &unsignedIssuer{iss: iss}, nil
}

func NewSymmIssuer(iss string, key []byte) (Issuer, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("%w: key must not be empty", ErrConstructInstanceFail)
	}
	return &symmIssuer{iss: iss, key: key}, nil
}

func NewAsymmIssuer(iss string, key *rsa.PrivateKey) (Issuer, error) {
	if key == nil {
		return nil, fmt.Errorf("%w: private key must not be nil", ErrConstructInstanceFail)
	}
	return &asymmIssuer{iss: iss, key: key}, nil
}

type Verifier interface {
	VerifyToken(tokenStr string) (jwt.MapClaims, error)
}

type unsignedVerifier struct{}

type symmVerifier struct {
	key []byte
}

type asymmVerifier struct {
	key *rsa.PublicKey
}

func (v *unsignedVerifier) VerifyToken(token string) (jwt.MapClaims, error) {
	return verifyToken(token, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodNone {
			return nil, fmt.Errorf("%w: expected none, got %T", ErrTokenVerificationFail, t.Method)
		}
		return jwt.UnsafeAllowNoneSignatureType, nil
	})
}

func (v *symmVerifier) VerifyToken(token string) (jwt.MapClaims, error) {
	return verifyToken(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: expected HMAC, got %T", ErrTokenVerificationFail, t.Method)
		}
		return v.key, nil
	})
}

func (v *asymmVerifier) VerifyToken(token string) (jwt.MapClaims, error) {
	return verifyToken(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: expected RSA, got %T", ErrTokenVerificationFail, t.Method)
		}
		return v.key, nil
	})
}

func NewUnsignedVerifier() (Verifier, error) {
	return &unsignedVerifier{}, nil
}

func NewSymmVerifier(key []byte) (Verifier, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("%w: key must not be empty", ErrConstructInstanceFail)
	}

	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)

	return &symmVerifier{key: keyCopy}, nil
}

func NewAsymmVerifier(key *rsa.PublicKey) (Verifier, error) {
	if key == nil {
		return nil, fmt.Errorf("%w: public key must not be nil", ErrConstructInstanceFail)
	}
	return &asymmVerifier{key: key}, nil
}

func buildClaims(iss string, sub string, aud []string, claims map[string]any, ttl time.Duration, nbf time.Time) (jwt.MapClaims, error) {
	now := time.Now()

	rand, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate token UUID: %v", ErrTokenIssuanceFail, err)
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
		return "", fmt.Errorf("%w: failed to sign token: %v", ErrTokenIssuanceFail, err)
	}
	return tokenStr, nil
}

func verifyToken(token string, kf jwt.Keyfunc) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	if _, err := jwt.ParseWithClaims(token, claims, kf); err != nil {
		return nil, fmt.Errorf("%w: parsing claims failed: %v", ErrTokenClaimsInvalid, err)
	}

	return claims, nil
}
