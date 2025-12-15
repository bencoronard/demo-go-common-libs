package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JwtVerifier interface {
	VerifyToken(tokenStr string) (jwt.MapClaims, error)
}

type unsignedJwtVerifier struct {
	kf jwt.Keyfunc
}

type symmJwtVerifier struct {
	kf jwt.Keyfunc
}

type asymmJwtVerifier struct {
	kf jwt.Keyfunc
}

func NewUnsignedJwtVerifier() (JwtVerifier, error) {
	return &unsignedJwtVerifier{kf: func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodNone {
			return nil, fmt.Errorf("%w: %T", ErrTokenUnexpectedSignMethod, t.Method)
		}
		return jwt.UnsafeAllowNoneSignatureType, nil
	}}, nil
}

func NewSymmJwtVerifier(key []byte) (JwtVerifier, error) {
	if key == nil {
		return nil, errors.New("symmetric key must not be nil")
	}

	return &symmJwtVerifier{kf: func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %T", ErrTokenUnexpectedSignMethod, t.Method)
		}
		return key, nil
	}}, nil
}

func NewAsymmJwtVerifier(key *rsa.PublicKey) (JwtVerifier, error) {
	if key == nil {
		return nil, errors.New("public key must not be nil")
	}
	return &asymmJwtVerifier{kf: func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: %T", ErrTokenUnexpectedSignMethod, t.Method)
		}
		return key, nil
	}}, nil
}

func (v *unsignedJwtVerifier) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	return verifyToken(tokenStr, v.kf)
}

func (v *symmJwtVerifier) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	return verifyToken(tokenStr, v.kf)
}

func (v *asymmJwtVerifier) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
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
