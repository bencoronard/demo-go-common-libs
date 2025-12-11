package reader

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	require.NotNil(t, key)

	assert.NoError(t, key.Validate())
}

func TestReadRsaPublicKeyX509_shouldReturnPublicKey(t *testing.T) {
	key, err := ReadRsaPublicKeyX509(string(rsaPublicX509PEM))
	require.NoError(t, err)
	require.NotNil(t, key)

	assert.True(t, key.N.Sign() > 0)
	assert.True(t, key.E > 1)
}

func TestDerivePublicKeyFromPrivateKey_shouldMatchReadPublicKey(t *testing.T) {
	privKey, err := ReadRsaPrivateKeyPkcs8(string(rsaPrivatePkcs8PEM))
	require.NoError(t, err)
	require.NotNil(t, privKey)

	pubKey, err := ReadRsaPublicKeyX509(string(rsaPublicX509PEM))
	require.NoError(t, err)
	require.NotNil(t, pubKey)

	assert.Equal(t, 0, privKey.PublicKey.N.Cmp(pubKey.N))
	assert.Equal(t, pubKey.E, privKey.PublicKey.E)
}
