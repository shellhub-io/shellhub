package models

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

type Device struct {
	// Device's UID.
	UID        string          `json:"uid" example:"13b0c8ea878e61ff849db69461795006a9594c8f6a6390ce0000100b0c9d7d0a"`
	// Device's name.
	Name       string          `json:"name" bson:"name,omitempty" validate:"required,hostname_rfc1123,excludes=." example:"d0-94-00-b4-25-bc "`
	Identity   *DeviceIdentity `json:"identity"`
	// Device's information.
	Info       *DeviceInfo     `json:"info"`
	// Device's public key.
	PublicKey  string          `json:"public_key" bson:"public_key"`
	// Device's namespace tenant id.
	TenantID   string          `json:"tenant_id" bson:"tenant_id" example:"3dd0d1f8-8246-4519-b11a-a3dd33717f65"`
	// Device's last seen date.
	LastSeen   time.Time       `json:"last_seen" bson:"last_seen" example:"2020-01-01T00:00:00Z"`
	// Device's online status.
	Online     bool            `json:"online" bson:",omitempty"`
	// Device's namespace name.
	Namespace  string          `json:"namespace" bson:",omitempty" example:"examplespace"`
	// Device's status.
	Status     string          `json:"status" bson:"status,omitempty" validate:"oneof=accepted rejected pending unused" example:"accepted"`
	// Device's created date.
	CreatedAt  time.Time       `json:"created_at" bson:"created_at,omitempty" example:"2020-01-01T00:00:00Z"`
	// Device's remove address.
	RemoteAddr string          `json:"remote_addr" bson:"remote_addr" example:"127.0.0.1"`
	Position   *DevicePosition `json:"position" bson:"position"`
	// Device's tags.
	Tags       []string        `json:"tags" bson:"tags,omitempty"`
}

type DeviceAuthClaims struct {
	UID string `json:"uid"`

	AuthClaims           `mapstruct:",squash"`
	jwt.RegisteredClaims `mapstruct:",squash"`
}

type DeviceAuthRequest struct {
	Info     *DeviceInfo `json:"info"`
	Sessions []string    `json:"sessions,omitempty"`
	*DeviceAuth
}

type DeviceAuth struct {
	Hostname  string          `json:"hostname,omitempty" bson:"hostname,omitempty" validate:"required_without=Identity,omitempty,hostname_rfc1123" hash:"-"`
	Identity  *DeviceIdentity `json:"identity,omitempty" bson:"identity,omitempty" validate:"required_without=Hostname,omitempty"`
	PublicKey string          `json:"public_key"`
	TenantID  string          `json:"tenant_id"`
}

type DeviceAuthResponse struct {
	UID       string `json:"uid"`
	Token     string `json:"token"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Device's identity.
type DeviceIdentity struct {
	// Device's MAC
	MAC string `json:"mac" example:"00:00:00:00:00:00"`
}

// Device's information.
type DeviceInfo struct {
	//  Device's OS name.
	ID         string `json:"id" example:"linux"`
	// Device's OS pretty name.
	PrettyName string `json:"pretty_name" example:"Linux"`
	// Device's OS version.
	Version    string `json:"version" example:"latest"`
	// Device's OS arch.
	Arch       string `json:"arch" example:"x86_64"`
	// Device's OS platform.
	Platform   string `json:"platform" validate:"oneof=linux docker" example:"docker"`
}

type ConnectedDevice struct {
	UID      string    `json:"uid"`
	TenantID string    `json:"tenant_id" bson:"tenant_id"`
	LastSeen time.Time `json:"last_seen" bson:"last_seen"`
	Status   string    `json:"status" bson:"status"`
}

// Device's geolocation.
type DevicePosition struct {
	// Device's latitude position.
	Latitude  float64 `json:"latitude" bson:"latitude" example:"-31.7566628"`
	// Device's longitude position.
	Longitude float64 `json:"longitude" bson:"longitude" example:"-52.322474"`
}
