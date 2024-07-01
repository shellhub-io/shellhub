package jwttoken

import (
	"crypto/rsa"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

type ClaimsKind string

const (
	KindUserClaims    ClaimsKind = "user"
	KindDeviceClaims  ClaimsKind = "device"
	KindUnknownClaims ClaimsKind = "unknown"
)

func ClaimsKindFromString(str string) ClaimsKind {
	switch str {
	case "user":
		return KindUserClaims
	case "device":
		return KindDeviceClaims
	default:
		return KindUnknownClaims
	}
}

type UserClaims struct {
	ID       string          `json:"id"`
	TenantID string          `json:"tenant"`
	Role     authorizer.Role `json:"-"`
	Username string          `json:"name"`
	MFA      bool            `json:"mfa"`
	jwt.RegisteredClaims
}

type DeviceClaims struct {
	UID    string `json:"uid"`
	Tenant string `json:"tenant"`
	jwt.RegisteredClaims
}

func ClaimsFromBearer(publicKey *rsa.PublicKey, bearerToken string) (interface{}, error) {
	raw := strings.ReplaceAll(bearerToken, "Bearer ", "")

	kind, err := unmarshalKind(publicKey, raw)
	if err != nil {
		return nil, err
	}

	return unmarshalClaims(publicKey, kind, raw)
}

func unmarshalKind(publicKey *rsa.PublicKey, raw string) (ClaimsKind, error) {
	aux := struct {
		Kind string `json:"claims"`
		jwt.RegisteredClaims
	}{}

	if _, err := jwt.ParseWithClaims(raw, &aux, eval(publicKey), jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()})); err != nil {
		return "", err
	}

	return ClaimsKindFromString(aux.Kind), nil
}

func unmarshalClaims(publicKey *rsa.PublicKey, kind ClaimsKind, raw string) (interface{}, error) {
	switch kind {
	case KindUserClaims:
		claims := new(UserClaims)
		if err := Decode(publicKey, raw, claims); err != nil {
			return nil, err
		}

		return claims, nil
	case KindDeviceClaims:
		claims := new(DeviceClaims)
		if err := Decode(publicKey, raw, claims); err != nil {
			return nil, err
		}

		return claims, nil
	default:
		panic("foo")
	}
}
