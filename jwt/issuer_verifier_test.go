package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
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

func TestHelloWorld(t *testing.T) {

}
