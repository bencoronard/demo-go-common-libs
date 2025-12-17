package reader

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"
)

func ReadRsaPrivateKeyPkcs8(content string) (*rsa.PrivateKey, error) {
	clean := strings.ReplaceAll(content, "-----BEGIN PRIVATE KEY-----", "")
	clean = strings.ReplaceAll(clean, "-----END PRIVATE KEY-----", "")
	clean = strings.TrimSpace(clean)

	der, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode base64: %v", ErrKeyFormatInvalid, err)
	}

	key, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse private key: %v", ErrKeyFormatInvalid, err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%w: expected RSA private key, got %T", ErrKeyTypeMismatch, key)
	}

	return rsaKey, nil
}

func ReadRsaPublicKeyX509(content string) (*rsa.PublicKey, error) {
	clean := strings.ReplaceAll(content, "-----BEGIN PUBLIC KEY-----", "")
	clean = strings.ReplaceAll(clean, "-----END PUBLIC KEY-----", "")
	clean = strings.TrimSpace(clean)

	der, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode base64: %v", ErrKeyFormatInvalid, err)
	}

	key, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse public key: %v", ErrKeyFormatInvalid, err)
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%w: expected RSA public key, got %T", ErrKeyTypeMismatch, key)
	}

	return rsaKey, nil
}
