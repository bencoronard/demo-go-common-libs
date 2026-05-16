package http

import "errors"

var (
	ErrAuthHeaderInvalid = errors.New("missing or invalid Authorization header format")
	ErrAuthTokenInvalid  = errors.New("invalid token")
)
