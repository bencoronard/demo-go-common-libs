package jwt

import "errors"

var (
	ErrTokenUnexpectedSignMethod  = errors.New("token has unexpected signing method")
	ErrTokenClaimsInvalid         = errors.New("token has invalid claims")
	ErrExpiredTokenIssueAttempted = errors.New("expired token issuance attempted")
	ErrTokenMalformed             = errors.New("token is malformed")
)
