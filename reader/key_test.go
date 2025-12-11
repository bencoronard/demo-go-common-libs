package reader

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	rsaPrivatePkcs8PEM []byte
	rsaPublicX509PEM   []byte
)

func TestMain(m *testing.M) {
	var err error

	rsaPrivatePkcs8PEM, err = os.ReadFile("testdata/rsa-private-pkcs8.pem")
	if err != nil {
		panic("failed to read private key fixture for setup: " + err.Error())
	}

	rsaPublicX509PEM, err = os.ReadFile("testdata/rsa-public-x509.pem")
	if err != nil {
		panic("failed to read public key fixture for setup: " + err.Error())
	}

	os.Exit(m.Run())
}

func TestReadRsaPrivateKeyPkcs8_shouldReturnPrivateKey(t *testing.T) {
	key, err := ReadRsaPrivateKeyPkcs8(string(rsaPrivatePkcs8PEM))

	assert.NoError(t, err, "expected private key to parse successfully")

	if key != nil {
		assert.NoError(t, key.Validate(), "expected valid private key after parsing")
	}
}

func TestReadRsaPublicKeyX509_shouldReturnPublicKey(t *testing.T) {
	key, err := ReadRsaPublicKeyX509(string(rsaPublicX509PEM))

	assert.NoError(t, err, "expected public key to parse successfully")

	if key != nil {
		assert.NotNil(t, key.N, "public key modulus N should not be nil")
		assert.True(t, key.N.Sign() > 0, "public key modulus N should be positive")
		assert.True(t, key.E > 1, "public key exponent E should be greater than 1")
	}
}

func TestDerivePublicKeyFromPrivateKey_shouldMatchReadPublicKey(t *testing.T) {
	privKey, err := ReadRsaPrivateKeyPkcs8(string(rsaPrivatePkcs8PEM))

	if !assert.NoError(t, err, "expected private key to parse successfully") {
		return
	}

	pubKey, err := ReadRsaPublicKeyX509(string(rsaPublicX509PEM))
	if !assert.NoError(t, err, "expected public key to parse successfully") {
		return
	}

	if privKey != nil && pubKey != nil {
		assert.Equal(t, 0, privKey.PublicKey.N.Cmp(pubKey.N), "derived public key modulus (N) must match read public key modulus")
		assert.Equal(t, pubKey.E, privKey.PublicKey.E, "derived public key exponent (E) must match read public key exponent")
	}
}
