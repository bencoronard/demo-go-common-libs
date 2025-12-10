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
}

type symmJwtVerifier struct {
	key []byte
}

type asymmJwtVerifier struct {
	key *rsa.PublicKey
}

func NewUnsignedJwtVerifier() (JwtVerifier, error) {
	return &unsignedJwtVerifier{}, nil
}

func NewSymmJwtVerifier(key []byte) (JwtVerifier, error) {
	if key == nil {
		return nil, errors.New("symmetric key must not be nil")
	}
	return &symmJwtVerifier{key: key}, nil
}

func NewAsymmJwtVerifier(key *rsa.PublicKey) (JwtVerifier, error) {
	if key == nil {
		return nil, errors.New("public key must not be nil")
	}
	return &asymmJwtVerifier{key: key}, nil
}

func (v *unsignedJwtVerifier) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	return verifyToken(tokenStr, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodNone {
			return nil, fmt.Errorf("%w: %T", ErrTokenUnexpectedSignMethod, t.Method)
		}
		return jwt.UnsafeAllowNoneSignatureType, nil
	})
}

func (v *symmJwtVerifier) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	return verifyToken(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %T", ErrTokenUnexpectedSignMethod, t.Method)
		}
		return v.key, nil
	})
}

func (v *asymmJwtVerifier) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	return verifyToken(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: %T", ErrTokenUnexpectedSignMethod, t.Method)
		}
		return v.key, nil
	})
}

func verifyToken(tokenStr string, keyFunc jwt.Keyfunc) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, keyFunc)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrTokenClaimsInvalid
	}

	return claims, nil
}
