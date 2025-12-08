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

func NewUnsignedJwtVerifier() JwtVerifier {
	return &unsignedJwtVerifier{}
}

func NewSymmJwtVerifier(key []byte) JwtVerifier {
	return &symmJwtVerifier{key: key}
}

func NewAsymmJwtVerifier(key *rsa.PublicKey) JwtVerifier {
	return &asymmJwtVerifier{key: key}
}

func (v *unsignedJwtVerifier) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodNone {
			return nil, fmt.Errorf("unexpected signing method: %T", t.Method)
		}
		return jwt.UnsafeAllowNoneSignatureType, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (v *symmJwtVerifier) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %T", t.Method)
		}
		return v.key, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (v *asymmJwtVerifier) VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %T", t.Method)
		}
		return v.key, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
