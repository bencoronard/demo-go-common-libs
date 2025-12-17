package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	issuerName = "hireben.dev"
)

var (
	symmetricKey []byte
	privateKey   *rsa.PrivateKey
	publicKey    *rsa.PublicKey
)

func TestMain(m *testing.M) {
	var err error

	symmetricKey = make([]byte, 32)
	if _, err = rand.Read(symmetricKey); err != nil {
		panic("failed to generate symmetric key: " + err.Error())
	}

	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("failed to generate RSA key pair: " + err.Error())
	}

	publicKey = &privateKey.PublicKey

	os.Exit(m.Run())
}

func TestNewSymmVerifier_withNilSymmetricKey_shouldReturnError(t *testing.T) {
	_, err := NewSymmVerifier(nil)
	assert.ErrorIs(t, err, ErrConstructInstanceFail)
}

func TestNewAsymmVerifier_withNilPublicKey_shouldReturnError(t *testing.T) {
	_, err := NewAsymmVerifier(nil)
	assert.ErrorIs(t, err, ErrConstructInstanceFail)
}

func TestNewSymmIssuer_withNilSymmetricKey_shouldReturnError(t *testing.T) {
	_, err := NewSymmIssuer(issuerName, nil)
	assert.ErrorIs(t, err, ErrConstructInstanceFail)
}

func TestNewAsymmIssuer_withNilPrivateKey_shouldReturnError(t *testing.T) {
	_, err := NewAsymmIssuer(issuerName, nil)
	assert.ErrorIs(t, err, ErrConstructInstanceFail)
}

func TestIssueToken_withoutKey_withInvalidTtl_shouldReturnError(t *testing.T) {
	issuer, err := NewUnsignedIssuer(issuerName)
	require.NoError(t, err)
	require.NotNil(t, issuer)

	ttl := -1 * time.Second
	_, err = issuer.IssueToken("", nil, nil, &ttl, nil)

	assert.ErrorIs(t, err, ErrTokenIssuanceFail)
}

func TestIssueToken_withSymmKey_withInvalidTtl_shouldReturnError(t *testing.T) {
	issuer, err := NewSymmIssuer(issuerName, symmetricKey)
	require.NoError(t, err)
	require.NotNil(t, issuer)

	ttl := -1 * time.Second
	_, err = issuer.IssueToken("", nil, nil, &ttl, nil)

	assert.ErrorIs(t, err, ErrTokenIssuanceFail)
}

func TestIssueToken_withAsymmKey_withInvalidTtl_shouldReturnError(t *testing.T) {
	issuer, err := NewAsymmIssuer(issuerName, privateKey)
	require.NoError(t, err)
	require.NotNil(t, issuer)

	ttl := -1 * time.Second
	_, err = issuer.IssueToken("", nil, nil, &ttl, nil)

	assert.ErrorIs(t, err, ErrTokenIssuanceFail)
}

func TestIssueToken_withoutKey_shouldBeParsableWithUnsecuredVerifier(t *testing.T) {
	issuer, err := NewUnsignedIssuer(issuerName)
	require.NoError(t, err)
	require.NotNil(t, issuer)

	verifier, err := NewUnsignedVerifier()
	require.NoError(t, err)
	require.NotNil(t, verifier)

	token, err := issuer.IssueToken("", nil, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, token)

	claims, err := verifier.VerifyToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.NotNil(t, claims["jti"])
	assert.NotNil(t, claims["iat"])
}

func TestIssueToken_withSymmKey_shouldBeParsableWithSymmVerifier(t *testing.T) {
	issuer, err := NewSymmIssuer(issuerName, symmetricKey)
	require.NoError(t, err)
	require.NotNil(t, issuer)

	verifier, err := NewSymmVerifier(symmetricKey)
	require.NoError(t, err)
	require.NotNil(t, verifier)

	token, err := issuer.IssueToken("", nil, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, token)

	claims, err := verifier.VerifyToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.NotNil(t, claims["jti"])
	assert.NotNil(t, claims["iat"])
}

func TestIssueToken_withAsymmKey_shouldBeParsableWithAsymmVerifier(t *testing.T) {
	issuer, err := NewAsymmIssuer(issuerName, privateKey)
	require.NoError(t, err)
	require.NotNil(t, issuer)

	verifier, err := NewAsymmVerifier(publicKey)
	require.NoError(t, err)
	require.NotNil(t, verifier)

	token, err := issuer.IssueToken("", nil, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, token)

	claims, err := verifier.VerifyToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.NotNil(t, claims["jti"])
	assert.NotNil(t, claims["iat"])
}

func TestVerifyToken_withoutKey_whenTokenExpired_shouldReturnError(t *testing.T) {
	verifier, err := NewUnsignedVerifier()
	require.NoError(t, err)
	require.NotNil(t, verifier)

	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Second)),
	})
	require.NotNil(t, token)

	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = verifier.VerifyToken(tokenStr)
	assert.ErrorIs(t, err, ErrTokenClaimsInvalid)
}

func TestVerifyToken_withSymmKey_whenTokenExpired_shouldReturnError(t *testing.T) {
	verifier, err := NewSymmVerifier(symmetricKey)
	require.NoError(t, err)
	require.NotNil(t, verifier)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Second)),
	})
	require.NotNil(t, token)

	tokenStr, err := token.SignedString(symmetricKey)
	require.NoError(t, err)

	_, err = verifier.VerifyToken(tokenStr)
	assert.ErrorIs(t, err, ErrTokenClaimsInvalid)
}

func TestVerifyToken_withAsymmKey_whenTokenExpired_shouldReturnError(t *testing.T) {
	verifier, err := NewAsymmVerifier(publicKey)
	require.NoError(t, err)
	require.NotNil(t, verifier)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Second)),
	})
	require.NotNil(t, token)

	tokenStr, err := token.SignedString(privateKey)
	require.NoError(t, err)

	_, err = verifier.VerifyToken(tokenStr)
	assert.ErrorIs(t, err, ErrTokenClaimsInvalid)
}

func TestVerifyToken_withoutKey_whenTokenNotYetUsable_shouldReturnError(t *testing.T) {
	verifier, err := NewUnsignedVerifier()
	require.NoError(t, err)
	require.NotNil(t, verifier)

	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
		NotBefore: jwt.NewNumericDate(time.Now().Add(5 * time.Second)),
	})
	require.NotNil(t, token)

	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = verifier.VerifyToken(tokenStr)
	assert.ErrorIs(t, err, ErrTokenClaimsInvalid)
}

func TestVerifyToken_withSymmKey_whenTokenNotYetUsable_shouldReturnError(t *testing.T) {
	verifier, err := NewSymmVerifier(symmetricKey)
	require.NoError(t, err)
	require.NotNil(t, verifier)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		NotBefore: jwt.NewNumericDate(time.Now().Add(5 * time.Second)),
	})
	require.NotNil(t, token)

	tokenStr, err := token.SignedString(symmetricKey)
	require.NoError(t, err)

	_, err = verifier.VerifyToken(tokenStr)
	assert.ErrorIs(t, err, ErrTokenClaimsInvalid)
}

func TestVerifyToken_withAsymmKey_whenTokenNotYetUsable_shouldReturnError(t *testing.T) {
	verifier, err := NewAsymmVerifier(publicKey)
	require.NoError(t, err)
	require.NotNil(t, verifier)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		NotBefore: jwt.NewNumericDate(time.Now().Add(5 * time.Second)),
	})
	require.NotNil(t, token)

	tokenStr, err := token.SignedString(privateKey)
	require.NoError(t, err)

	_, err = verifier.VerifyToken(tokenStr)
	assert.ErrorIs(t, err, ErrTokenClaimsInvalid)
}
