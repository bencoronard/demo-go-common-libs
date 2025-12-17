package reader

import "errors"

var (
	ErrKeyFormatInvalid = errors.New("invalid key format")
	ErrKeyTypeMismatch  = errors.New("key type assertion failure")
)
