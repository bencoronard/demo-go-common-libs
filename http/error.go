package http

import "errors"

var (
	ErrMissingRequestHeader = errors.New("missing request header")
)
