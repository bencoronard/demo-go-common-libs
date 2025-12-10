package reader

import (
	"os"
	"testing"
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
	if err != nil {
		t.Fatalf("expected private key to parse successfully, got error: %v", err)
	}

	if err := key.Validate(); err != nil {
		t.Fatalf("expected valid private key, got error: %v", err)
	}
}

func TestReadRsaPublicKeyX509_shouldReturnPublicKey(t *testing.T) {
	key, err := ReadRsaPublicKeyX509(string(rsaPublicX509PEM))
	if err != nil {
		t.Fatalf("expected public key to parse successfully, got error: %v", err)
	}

	if key.N == nil || key.N.Sign() <= 0 || key.E <= 1 {
		t.Fatal("expected valid public key")
	}
}

func TestDerivePublicKeyFromPrivateKey_shouldMatchReadPublicKey(t *testing.T) {
	privKey, err := ReadRsaPrivateKeyPkcs8(string(rsaPrivatePkcs8PEM))
	if err != nil {
		t.Fatalf("expected private key to parse successfully, got error: %v", err)
	}

	pubKey, err := ReadRsaPublicKeyX509(string(rsaPublicX509PEM))
	if err != nil {
		t.Fatalf("expected public key to parse successfully, got error: %v", err)
	}

	if privKey.PublicKey.N.Cmp(pubKey.N) != 0 || privKey.PublicKey.E != pubKey.E {
		t.Fatal("derived public key does not match read public key")
	}
}
