package vault

import "errors"

var (
	ErrSecretNotFound     = errors.New("secret not unavailable")
	ErrAuthenticationFail = errors.New("vault authentication fail")
)
