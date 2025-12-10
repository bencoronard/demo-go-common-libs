package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	issuerName = "hireben.dev"
)

var (
	symmetricKey []byte
	privateKey   *rsa.PrivateKey
	publicKey    *rsa.PublicKey
)

func TestMain(m *testing.M) {
	var err error

	symmetricKey = make([]byte, 32)
	if _, err = rand.Read(symmetricKey); err != nil {
		panic("failed to generate symmetric key: " + err.Error())
	}

	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("failed to generate RSA key pair: " + err.Error())
	}

	publicKey = &privateKey.PublicKey

	os.Exit(m.Run())
}

func TestNewSymmJwtVerifier_withNilSymmetricKey_shouldReturnError(t *testing.T) {
	_, err := NewSymmJwtVerifier(nil)
	assert.EqualError(t, err, "symmetric key must not be nil")
}

func TestNewAsymmJwtVerifier_withNilPublicKey_shouldReturnError(t *testing.T) {
	_, err := NewAsymmJwtVerifier(nil)
	assert.EqualError(t, err, "public key must not be nil")
}

func TestNewSymmJwtIssuer_withNilSymmetricKey_shouldReturnError(t *testing.T) {
	_, err := NewSymmJwtIssuer(issuerName, nil)
	assert.EqualError(t, err, "symmetric key must not be nil")
}

func TestNewAsymmJwtIssuer_withNilPrivateKey_shouldReturnError(t *testing.T) {
	_, err := NewAsymmJwtIssuer(issuerName, nil)
	assert.EqualError(t, err, "private key must not be nil")
}
