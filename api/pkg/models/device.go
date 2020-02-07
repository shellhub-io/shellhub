package models

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"gopkg.in/mgo.v2/bson"
)

type Device struct {
	ID         bson.ObjectId     `json:"-" bson:"_id,omitempty"`
	UID        string            `json:"uid"`
	Name       string            `json:"name" bson:"name,omitempty"`
	Identity   map[string]string `json:"identity"`
	Attributes map[string]string `json:"attributes"`
	PublicKey  string            `json:"public_key" bson:"public_key"`
	TenantID   string            `json:"tenant_id" bson:"tenant_id"`
	LastSeen   time.Time         `json:"last_seen"`
	Online     bool              `json:"online"`
	Namespace  string            `json:"namespace"`
}

type DeviceAuthClaims struct {
	UID string `json:"uid"`

	jwt.StandardClaims
}

type DeviceAuthRequest struct {
	Attributes map[string]string `json:"attributes"`
	Sessions   []string          `json:"sessions,omitempty"`
	*DeviceAuth
}

type DeviceAuth struct {
	Identity  map[string]string `json:"identity"`
	PublicKey string            `json:"public_key"`
	TenantID  string            `json:"tenant_id"`
}

type DeviceAuthResponse struct {
	UID       string `json:"uid"`
	Token     string `json:"token"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type ConnectedDevice struct {
	ID       bson.ObjectId `json:"-" bson:"_id,omitempty"`
	UID      string        `json:"uid"`
	TenantID string        `json:"tenant_id" bson:"tenant_id"`
	LastSeen time.Time     `json:"last_seen" bson:"last_seen"`
}
