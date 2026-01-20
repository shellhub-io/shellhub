package entity

import (
	"time"

	"github.com/uptrace/bun"
)

type Tunnel struct {
	bun.BaseModel `bun:"table:tunnels"`

	ID          string    `bun:"id,pk,type:uuid"`
	NamespaceID string    `bun:"namespace_id"`
	DeviceID    string    `bun:"device_id"`
	Address     string    `bun:"address"`
	Host        string    `bun:"host"`
	Port        int       `bun:"port"`
	TLSEnabled  bool      `bun:"tls_enabled"`
	TLSVerify   bool      `bun:"tls_verify"`
	TLSDomain   string    `bun:"tls_domain,nullzero"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`

	Namespace *Namespace `bun:"rel:belongs-to,join:namespace_id=id"`
	Device    *Device    `bun:"rel:belongs-to,join:device_id=id"`
}
