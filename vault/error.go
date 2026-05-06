package vault

import "errors"

var (
	ErrSecretNotFound = errors.New("secret not found")
	ErrConfig         = errors.New("")
)
