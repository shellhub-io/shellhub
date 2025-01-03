package models

import (
	"time"
)

type DeviceStatus string

const (
	DeviceStatusAccepted DeviceStatus = "accepted"
	DeviceStatusPending  DeviceStatus = "pending"
	DeviceStatusRejected DeviceStatus = "rejected"
	DeviceStatusRemoved  DeviceStatus = "removed"
	DeviceStatusUnused   DeviceStatus = "unused"
	DeviceStatusEmpty    DeviceStatus = ""
)

type Device struct {
	LastSeen         time.Time       `json:"last_seen" bson:"last_seen"`
	CreatedAt        time.Time       `json:"created_at" bson:"created_at,omitempty"`
	StatusUpdatedAt  time.Time       `json:"status_updated_at" bson:"status_updated_at,omitempty"`
	Position         *DevicePosition `json:"position" bson:"position"`
	Identity         *DeviceIdentity `json:"identity"`
	Info             *DeviceInfo     `json:"info"`
	Namespace        string          `json:"namespace" bson:",omitempty"`
	UID              string          `json:"uid"`
	Status           DeviceStatus    `json:"status" bson:"status,omitempty" validate:"oneof=accepted rejected pending unused"`
	TenantID         string          `json:"tenant_id" bson:"tenant_id"`
	PublicKey        string          `json:"public_key" bson:"public_key"`
	RemoteAddr       string          `json:"remote_addr" bson:"remote_addr"`
	Name             string          `json:"name" bson:"name,omitempty" validate:"required,device_name"`
	PublicURLAddress string          `json:"public_url_address" bson:"public_url_address,omitempty"`
	Tags             []string        `json:"tags" bson:"tags,omitempty"`
	Online           bool            `json:"online" bson:",omitempty"`
	PublicURL        bool            `json:"public_url" bson:"public_url,omitempty"`
	Acceptable       bool            `json:"acceptable" bson:"acceptable,omitempty"`
}

type DeviceAuthRequest struct {
	Info *DeviceInfo `json:"info"`
	*DeviceAuth
	Sessions []string `json:"sessions,omitempty"`
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
	LastSeen time.Time `json:"last_seen" bson:"last_seen"`
	UID      string    `json:"uid"`
	TenantID string    `json:"tenant_id" bson:"tenant_id"`
}

type DevicePosition struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

type DeviceRemoved struct {
	Device    *Device   `json:"device" bson:"device"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type DeviceTag struct {
	Tag string `validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

func NewDeviceTag(tag string) DeviceTag {
	return DeviceTag{
		Tag: tag,
	}
}
