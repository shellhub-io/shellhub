package models

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Device struct {
	UID       string          `json:"uid"`
	Name      string          `json:"name" bson:"name,omitempty" validate:"required,hostname_rfc1123"`
	Identity  *DeviceIdentity `json:"identity"`
	Info      *DeviceInfo     `json:"info"`
	PublicKey string          `json:"public_key" bson:"public_key"`
	TenantID  string          `json:"tenant_id" bson:"tenant_id"`
	LastSeen  time.Time       `json:"last_seen" bson:"last_seen"`
	Online    bool            `json:"online" bson:",omitempty"`
	Namespace string          `json:"namespace" bson:",omitempty"`
	Pending   bool            `json:"pending" bson:"pending,omitempty"`
}

type DeviceAuthClaims struct {
	UID string `json:"uid"`

	AuthClaims         `json:",squash"`
	jwt.StandardClaims `json:",squash"`
}

type DeviceAuthRequest struct {
	Info     *DeviceInfo `json:"info"`
	Sessions []string    `json:"sessions,omitempty"`
	*DeviceAuth
}

type DeviceAuth struct {
	Identity  *DeviceIdentity `json:"identity"`
	PublicKey string          `json:"public_key"`
	TenantID  string          `json:"tenant_id"`
}

type DeviceAuthResponse struct {
	UID       string `json:"uid"`
	Token     string `json:"token"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type DeviceIdentity struct {
	MAC string `json:"mac"`
}

type DeviceInfo struct {
	ID         string `json:"id"`
	PrettyName string `json:"pretty_name"`
	Version    string `json:"version"`
}

type ConnectedDevice struct {
	UID      string    `json:"uid"`
	TenantID string    `json:"tenant_id" bson:"tenant_id"`
	LastSeen time.Time `json:"last_seen" bson:"last_seen"`
	Pending  bool      `json:"pending" bson:"pending"`
}
