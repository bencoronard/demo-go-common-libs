package reader

import "errors"

var (
	ErrInvalidKeyFormat = errors.New("invalid key format")
	ErrKeyTypeMismatch  = errors.New("key type assertion failure")
)
