package jwt

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Verifier interface {
	VerifyToken(tokenStr string) (jwt.MapClaims, error)
}

type unsignedVerifier struct {
	kf jwt.Keyfunc
}

type symmVerifier struct {
	kf jwt.Keyfunc
}

type asymmVerifier struct {
	kf jwt.Keyfunc
}

func NewUnsignedVerifier() (Verifier, error) {
	return &unsignedVerifier{kf: func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodNone {
			return nil, fmt.Errorf("%w: expected none, got %T", ErrTokenVerificationFail, t.Method)
		}
		return jwt.UnsafeAllowNoneSignatureType, nil
	}}, nil
}

func NewSymmVerifier(key []byte) (Verifier, error) {
	if key == nil {
		return nil, fmt.Errorf("%w: key must not be nil", ErrConstructInstanceFail)
	}

	return &symmVerifier{kf: func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: expected HMAC, got %T", ErrTokenVerificationFail, t.Method)
		}
		return key, nil
	}}, nil
}

func NewAsymmVerifier(key *rsa.PublicKey) (Verifier, error) {
	if key == nil {
		return nil, fmt.Errorf("%w: public key must not be nil", ErrConstructInstanceFail)
	}
	return &asymmVerifier{kf: func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: expected RSA, got %T", ErrTokenVerificationFail, t.Method)
		}
		return key, nil
	}}, nil
}

func (v *unsignedVerifier) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	return verifyToken(tokenStr, v.kf)
}

func (v *symmVerifier) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	return verifyToken(tokenStr, v.kf)
}

func (v *asymmVerifier) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	return verifyToken(tokenStr, v.kf)
}

func verifyToken(tokenStr string, kf jwt.Keyfunc) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(tokenStr, claims, kf)
	if err != nil {
		return nil, fmt.Errorf("%w: parsing claims failed: %v", ErrTokenClaimsInvalid, err)
	}

	return claims, nil
}
