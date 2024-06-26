package jwttoken

import (
	"crypto/rsa"
	"encoding/json"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

const (
	KindUserClaims    = "user"
	KindDeviceClaims  = "device"
	KindUnknownClaims = "unknown"
)

type UserClaims struct {
	ID       string          `json:"id"`
	TenantID string          `json:"tenant"`
	Role     authorizer.Role `json:"-"`
	Username string          `json:"name"`
	MFA      bool            `json:"mfa"`
}

type DeviceClaims struct {
	UID    string `json:"uid"`
	Tenant string `json:"tenant"`
}

type Claims struct {
	Kind         string       `json:"kind"`
	UserClaims   UserClaims   `json:"-"`
	DeviceClaims DeviceClaims `json:"-"`
	jwt.RegisteredClaims
}

func ClaimsFromBearer(publicKey *rsa.PublicKey, bearer string) *Claims {
	raw := strings.ReplaceAll(bearer, "Bearer ", "")

	claims := new(Claims)
	if err := Decode(publicKey, raw, claims); err != nil {
		claims.Kind = KindUnknownClaims
	}

	return claims
}

func (c *Claims) UnmarshalJSON(data []byte) error {
	aux := struct {
		Kind string `json:"claims"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch aux.Kind {
	case KindUserClaims:
		userClaims := new(UserClaims)
		if err := json.Unmarshal(data, userClaims); err != nil {
			return err
		}

		c.Kind = KindUserClaims
		c.UserClaims = *userClaims
	case KindDeviceClaims:
		deviceClaims := new(DeviceClaims)
		if err := json.Unmarshal(data, &deviceClaims); err != nil {
			return err
		}

		c.Kind = KindDeviceClaims
		c.DeviceClaims = *deviceClaims
	default:
		c.Kind = KindUnknownClaims
	}

	return nil
}

func (c *Claims) Headers() map[string]string {
	switch c.Kind {
	case KindUserClaims:
		return map[string]string{
			"X-ID":        c.UserClaims.ID,
			"X-Username":  c.UserClaims.Username,
			"X-Tenant-ID": c.UserClaims.TenantID,
			"X-Role":      c.UserClaims.Role.String(),
		}
	case KindDeviceClaims:
		return map[string]string{
			"X-Device-UID": c.DeviceClaims.UID,
			"X-Tenant-ID":  c.DeviceClaims.Tenant,
		}
	default:
		panic("invalid claim kind")
	}
}
