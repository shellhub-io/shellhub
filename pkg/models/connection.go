package models

import "time"

// ConnectionKind discriminates how a connection reaches its target.
type ConnectionKind string

const (
	// ConnectionKindExternal dials the target SSH endpoint directly, without the
	// agent (an external host reached by Host:Port).
	ConnectionKindExternal ConnectionKind = "external"
	// ConnectionKindDevice reaches an agent-registered device over the reverse
	// tunnel, reusing the standard device session flow.
	ConnectionKindDevice ConnectionKind = "device"
)

// Connection is a saved, reusable way to reach an SSH target, distinct from
// [Device]: it is user-provisioned inventory on top of the agent-registered
// fleet. The target is discriminated by Kind. A connection is personal (belongs
// to OwnerID); sharing one with a team is a separate Enterprise/Cloud capability.
type Connection struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	// OwnerID is the user the connection belongs to. It scopes visibility: only
	// the owner can see or use a connection.
	OwnerID string         `json:"owner_id"`
	Label   string         `json:"label"`
	Kind    ConnectionKind `json:"kind"`
	// Host and Port hold the dial target for Kind == ConnectionKindExternal.
	Host string `json:"host"`
	Port int    `json:"port"`
	// DeviceUID references the target device for Kind == ConnectionKindDevice.
	DeviceUID string `json:"device_uid"`
	Username  string `json:"username"`
	// AuthMethod is "password" or "key"; empty means none saved.
	AuthMethod string `json:"auth_method"`
	// KeyFingerprint points at the SSH key to use (resolved against the owner's
	// vault). The secret never reaches the server.
	KeyFingerprint string    `json:"key_fingerprint"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
