package jwt

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

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

func verifyToken(token string, kf jwt.Keyfunc) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	if _, err := jwt.ParseWithClaims(token, claims, kf); err != nil {
		return nil, fmt.Errorf("%w: parsing claims failed: %v", ErrTokenClaimsInvalid, err)
	}

	return claims, nil
}
