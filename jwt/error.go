package jwt

import "errors"

var (
	ErrConstructInstanceFail = errors.New("unable to create an instance")
	ErrTokenVerificationFail = errors.New("token verification failed")
	ErrTokenIssuanceFail     = errors.New("expired token issuance attempted")
	ErrTokenClaimsInvalid    = errors.New("token has invalid claims")
	ErrTokenMalformed        = errors.New("token is malformed")
)
