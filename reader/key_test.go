package reader

import (
	"os"
	"testing"
)

func TestReadRsaPrivateKeyPkcs8_shouldReturnPrivateKey(t *testing.T) {
	pem, err := os.ReadFile("testdata/rsa-private-pkcs8.pem")
	if err != nil {
		t.Fatalf("failed to read private key fixture: %v", err)
	}

	key, err := ReadRsaPrivateKeyPkcs8(string(pem))
	if err != nil {
		t.Fatalf("expected private key to parse successfully, got error: %v", err)
	}

	if err := key.Validate(); err != nil {
		t.Fatalf("expected valid private key, got error: %v", err)
	}
}

func TestReadRsaPublicKeyX509_shouldReturnPublicKey(t *testing.T) {
	pem, err := os.ReadFile("testdata/rsa-public-x509.pem")
	if err != nil {
		t.Fatalf("failed to read public key fixture: %v", err)
	}

	key, err := ReadRsaPublicKeyX509(string(pem))
	if err != nil {
		t.Fatalf("expected public key to parse successfully, got error: %v", err)
	}

	if key.N == nil || key.N.Sign() <= 0 || key.E <= 1 {
		t.Fatal("expected valid public key")
	}
}

func TestDerivePublicKeyFromPrivateKey_shouldMatchReadPublicKey(t *testing.T) {

	privPem, err := os.ReadFile("testdata/rsa-private-pkcs8.pem")
	if err != nil {
		t.Fatalf("failed to read private key fixture: %v", err)
	}

	privKey, err := ReadRsaPrivateKeyPkcs8(string(privPem))
	if err != nil {
		t.Fatalf("expected private key to parse successfully, got error: %v", err)
	}

	pubPem, err := os.ReadFile("testdata/rsa-public-x509.pem")
	if err != nil {
		t.Fatalf("failed to read public key fixture: %v", err)
	}

	pubKey, err := ReadRsaPublicKeyX509(string(pubPem))
	if err != nil {
		t.Fatalf("expected public key to parse successfully, got error: %v", err)
	}

	if privKey.PublicKey.N.Cmp(pubKey.N) != 0 || privKey.PublicKey.E != pubKey.E {
		t.Fatal("derived public key does not match read public key")
	}
}
