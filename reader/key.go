package reader

import "crypto/rsa"

func ReadRsaPrivateKeyPkcs8(content string) (*rsa.PrivateKey, error) {
	var key rsa.PrivateKey
	return &key, nil
}

func ReadRsaPublicKeyX509(content string) (*rsa.PublicKey, error) {
	var key rsa.PublicKey
	return &key, nil
}
