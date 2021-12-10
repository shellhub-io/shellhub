package models

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type Namespace struct {
	Name         string             `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
	Owner        string             `json:"owner"`
	Tokens       []Token            `json:"tokens" bson:"tokens"`
	TenantID     string             `json:"tenant_id" bson:"tenant_id,omitempty"`
	Members      []Member           `json:"members" bson:"members"`
	Settings     *NamespaceSettings `json:"settings"`
	Devices      int                `json:"-" bson:"devices,omitempty"`
	Sessions     int                `json:"-" bson:"sessions,omitempty"`
	MaxDevices   int                `json:"max_devices" bson:"max_devices"`
	DevicesCount int                `json:"devices_count" bson:"devices_count,omitempty"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	Billing      *Billing           `json:"billing" bson:"billing,omitempty"`
}

type NamespaceSettings struct {
	SessionRecord bool `json:"session_record" bson:"session_record,omitempty"`
}

type Member struct {
	ID       string `json:"id,omitempty" bson:"id,omitempty"`
	Username string `json:"username,omitempty" bson:"username,omitempty" validate:"min=3,max=30,alphanum,ascii"`
	Role     string `json:"role" bson:"role" validate:"required,oneof=administrator operator observer"`
}

type Token struct {
	ID       string `json:"id" bson:"id"`
	TenantID string `json:"tenant_id" bson:"tenant_id"`
	ReadOnly bool   `json:"read_only" bson:"read_only"`
}

type TokenAuthClaims struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	ReadOnly bool   `json:"read_only"`

	AuthClaims         `mapstruct:",squash"`
	jwt.StandardClaims `mapstruct:",squash"`
}

type TokenAuthRequest struct {
	TenantID string `json:"tenant_id"`
}

type TokenAuthResponse struct {
	ID       string `json:"id"`
	APIToken string `json:"api_token"`
	TenantID string `json:"tenant_id"`
	ReadOnly bool   `json:"read_only"`
}

type TokenFields struct {
	ReadOnly bool `json:"read_only"`
}

type TokenUpdate struct {
	TokenFields `bson:",inline"`
}
