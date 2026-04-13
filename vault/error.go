package vault

import "errors"

var (
	ErrConfigUnset        = errors.New("required vault config unset")
	ErrSecretNotFound     = errors.New("secret not unavailable")
	ErrAuthenticationFail = errors.New("vault authentication fail")
)
