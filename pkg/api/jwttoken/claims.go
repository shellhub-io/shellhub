package jwttoken

import (
	"crypto/rsa"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

// TODO:
// 1. Rename [user|device]Claims.Kind JSON's tag to "kind". (BREAKING CHANGE)
// 2. Rename this package to jwt.

type (
	// claimKind represents the type of claims used in JWT tokens.
	claimKind string

	// userClaims is an auxiliary type that embeds [github.com/golang-jwt/jwt/v4.RegisteredClaims]
	// into [github.com/shellhub-io/shellhub/pkg/api/authorizer.UserClaims] to convert it into
	// [github.com/golang-jwt/jwt/v4.Claims] for use in an [encode] call.
	userClaims struct {
		Kind claimKind `json:"claims"`
		authorizer.UserClaims
		jwt.RegisteredClaims
	}

	// deviceClaims is an auxiliary type that embeds [github.com/golang-jwt/jwt/v4.RegisteredClaims]
	// into [github.com/shellhub-io/shellhub/pkg/api/authorizer.DeviceClaims] to convert it into
	// [github.com/golang-jwt/jwt/v4.Claims] for use in an [encode] call.
	deviceClaims struct {
		Kind claimKind `json:"claims"`
		authorizer.DeviceClaims
		jwt.RegisteredClaims
	}
)

const (
	kindUserClaims    claimKind = "user"
	kindDeviceClaims  claimKind = "device"
	kindUnknownClaims claimKind = "unknown"
)

// claimKindFromString converts a string to a claimKind.
func claimKindFromString(str string) claimKind {
	switch str {
	case "user":
		return kindUserClaims
	case "device":
		return kindDeviceClaims
	default:
		return kindUnknownClaims
	}
}

// EncodeUserClaims encodes the provided user claims into a signed JWT token. It returns
// the encoded token and an error, if any.
//
// The token is valid for 72 hours; tenantID is optional.
func EncodeUserClaims(claims authorizer.UserClaims, privateKey *rsa.PrivateKey) (string, error) {
	now := time.Now()
	jwtClaims := userClaims{
		Kind:       kindUserClaims,
		UserClaims: claims,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.Generate(),
			Issuer:    "", // TODO: how can we get the correct issuer?
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 72)),
		},
	}

	token, err := encodeClaims(jwtClaims, privateKey)
	if err != nil {
		return "", err
	}

	return token, nil
}

// EncodeDeviceClaims encodes the provided device claims into a signed JWT token. It returns
// the encoded token and an error, if any.
func EncodeDeviceClaims(claims authorizer.DeviceClaims, privateKey *rsa.PrivateKey) (string, error) {
	now := time.Now()
	jwtClaims := deviceClaims{
		Kind:         kindDeviceClaims,
		DeviceClaims: claims,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.Generate(),
			Issuer:    "", // TODO: how can we get the correct issuer?
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token, err := encodeClaims(jwtClaims, privateKey)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ClaimsFromBearerToken decodes the provided bearer token into either [github.com/shellhub-io/shellhub/pkg/api/authorizer.UserClaims]
// or [github.com/shellhub-io/shellhub/pkg/api/authorizer.DeviceClaims]. It returns the decoded claims and an error if any.
func ClaimsFromBearerToken(publicKey *rsa.PublicKey, bearerToken string) (interface{}, error) {
	raw := strings.ReplaceAll(bearerToken, "Bearer ", "")

	return unmarshalBearerToken(publicKey, raw)
}

func unmarshalBearerToken(publicKey *rsa.PublicKey, raw string) (interface{}, error) {
	kindAux := struct {
		Kind string `json:"claims"`
		jwt.RegisteredClaims
	}{}

	if _, err := jwt.ParseWithClaims(raw, &kindAux, evalClaims(publicKey)); err != nil {
		return nil, err
	}

	switch claimKindFromString(kindAux.Kind) {
	case kindUserClaims:
		claims := new(userClaims)
		if err := decodeClaims(publicKey, raw, claims); err != nil {
			return nil, err
		}

		return &claims.UserClaims, nil
	case kindDeviceClaims:
		claims := new(deviceClaims)
		if err := decodeClaims(publicKey, raw, claims); err != nil {
			return nil, err
		}

		return &claims.DeviceClaims, nil
	default:
		return nil, errors.New("invalid JWT's kind")
	}
}
