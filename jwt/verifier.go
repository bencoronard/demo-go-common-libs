package jwt

import (
	"crypto/rsa"
	"errors"
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

const errFormat = "%w: %T"

func NewUnsignedVerifier() (Verifier, error) {
	return &unsignedVerifier{kf: func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodNone {
			return nil, fmt.Errorf(errFormat, ErrTokenUnexpectedSignMethod, t.Method)
		}
		return jwt.UnsafeAllowNoneSignatureType, nil
	}}, nil
}

func NewSymmVerifier(key []byte) (Verifier, error) {
	if key == nil {
		return nil, errors.New("symmetric key must not be nil")
	}

	return &symmVerifier{kf: func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(errFormat, ErrTokenUnexpectedSignMethod, t.Method)
		}
		return key, nil
	}}, nil
}

func NewAsymmVerifier(key *rsa.PublicKey) (Verifier, error) {
	if key == nil {
		return nil, errors.New("public key must not be nil")
	}
	return &asymmVerifier{kf: func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf(errFormat, ErrTokenUnexpectedSignMethod, t.Method)
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
		return nil, err
	}

	return claims, nil
}
