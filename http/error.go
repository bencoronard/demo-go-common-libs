package http

import "errors"

var (
	ErrAuthHeaderInvalid = errors.New("missing or invalid Authorization header format")
	ErrTokenInvalid      = errors.New("invalid token")
)
