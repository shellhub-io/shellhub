package models

import (
	"time"

	jwt "github.com/golang-jwt/jwt"
)

type Device struct {
	UID        string          `json:"uid"`
	Name       string          `json:"name" bson:"name,omitempty" validate:"required,hostname_rfc1123,excludes=."`
	Identity   *DeviceIdentity `json:"identity"`
	Info       *DeviceInfo     `json:"info"`
	PublicKey  string          `json:"public_key" bson:"public_key"`
	TenantID   string          `json:"tenant_id" bson:"tenant_id"`
	LastSeen   time.Time       `json:"last_seen" bson:"last_seen"`
	Online     bool            `json:"online" bson:",omitempty"`
	Namespace  string          `json:"namespace" bson:",omitempty"`
	Status     string          `json:"status" bson:"status,omitempty" validate:"oneof=accepted rejected pending unused`
	CreatedAt  time.Time       `json:"created_at" bson:"created_at,omitempty"`
	RemoteAddr string          `json:"remote_addr" bson:"remote_addr"`
	Position   *DevicePosition `json:"position" bson:"position"`
	Tags       []string        `json:"tags" bson:"tags,omitempty"`
}

type DeviceAuthClaims struct {
	UID string `json:"uid"`

	AuthClaims         `mapstruct:",squash"`
	jwt.StandardClaims `mapstruct:",squash"`
}

type DeviceAuthRequest struct {
	Info     *DeviceInfo `json:"info"`
	Sessions []string    `json:"sessions,omitempty"`
	*DeviceAuth
}

type DeviceAuth struct {
	Hostname  string          `json:"hostname,omitempty" bson:"hostname,omitempty" validate:"omitempty,hostname_rfc1123" hash:"-"`
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
	Arch       string `json:"arch"`
	Platform   string `json:"platform"`
}

type ConnectedDevice struct {
	UID      string    `json:"uid"`
	TenantID string    `json:"tenant_id" bson:"tenant_id"`
	LastSeen time.Time `json:"last_seen" bson:"last_seen"`
	Status   string    `json:"status" bson:"status"`
}

type DevicePosition struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}
