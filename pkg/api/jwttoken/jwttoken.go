package jwttoken

import (
	"crypto/rsa"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

var ErrMissingPrivateKey = errors.New("missing private key while trying to sign the token")

type Claims interface {
	jwt.Claims
	SetRegisteredClaims(claims jwt.RegisteredClaims)
}

type Token struct {
	registeredClaims jwt.RegisteredClaims
	privateKey       *rsa.PrivateKey
	claims           Claims
	method           jwt.SigningMethod
	raw              *jwt.Token
	str              string
}

// New creates a new Token that can be signed as a JWT token, pre-filled with default values for "jti," "iat," and "iss".
// and can be modified using the "WithExpire" method. It also provides a default signing method, RS256, which can be
// customized with the "WithMethod" method. To include non-default claims, use the "WithClaims" method. You must
// provide a valid private key using the "WithPrivateKey" method. To complete the token creation, use the "Sign" method.
func New() *Token {
	return &Token{
		str:        "",
		raw:        nil,
		claims:     nil,
		method:     jwt.SigningMethodRS256,
		privateKey: nil,
		registeredClaims: jwt.RegisteredClaims{
			ID:       uuid.Generate(),
			Issuer:   "https://cloud.shellhub.io",
			IssuedAt: jwt.NewNumericDate(clock.Now().UTC()),
		},
	}
}

// WithClaims sets the costum claims of the token. It will subscribe any predefined RegisteredClaims.
func (t *Token) WithClaims(claims Claims) *Token {
	claims.SetRegisteredClaims(t.registeredClaims)
	t.claims = claims

	return t
}

// WithExpire sets the expiration time for the JWT.
func (t *Token) WithExpire(exp time.Time) *Token {
	t.registeredClaims.ExpiresAt = jwt.NewNumericDate(exp)

	return t
}

// WithMethod sets the signing method for the JWT. Default is RS256.
func (t *Token) WithMethod(method jwt.SigningMethod) *Token {
	t.method = method

	return t
}

// WithPrivateKey sets the private key for signing the JWT.
func (t *Token) WithPrivateKey(pk *rsa.PrivateKey) *Token {
	t.privateKey = pk

	return t
}

// Sign finalizes the configuration of the JWT and signs it with the private key.
// If no custom claims have been set, it will sign an empty token with only the "jti" claim.
func (t *Token) Sign() (*Token, error) {
	if t.privateKey == nil {
		return nil, ErrMissingPrivateKey
	}

	var token *jwt.Token
	if t.claims != nil {
		token = jwt.NewWithClaims(t.method, t.claims)
	} else {
		token = jwt.NewWithClaims(t.method, t.registeredClaims)
	}

	tokenStr, err := token.SignedString(t.privateKey)
	if err != nil {
		return nil, err
	}

	t.raw = token
	t.str = tokenStr

	return t, nil
}

// String returns the string representation of the signed JWT.
func (t *Token) String() string {
	return t.str
}
