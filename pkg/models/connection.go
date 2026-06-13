package models

import "time"

// ConnectionKind discriminates how a connection reaches its target.
type ConnectionKind string

const (
	// ConnectionKindDirect dials the target SSH endpoint directly, without the
	// agent.
	ConnectionKindDirect ConnectionKind = "direct"
	// ConnectionKindDevice reaches an agent-registered device over the reverse
	// tunnel, reusing the standard device session flow.
	ConnectionKindDevice ConnectionKind = "device"
)

// Connection is a saved entry in the connection address book: a reusable way to
// reach an SSH target. It is intentionally distinct from [Device] (the
// agent-registered fleet); a Connection is user-provisioned inventory on top.
//
// The target is discriminated by Kind. The MVP supports only "direct" (Host:Port
// dialed directly); "device" (via the agent) and "bridge" (agent as jump) are
// planned and will reuse the same entity.
type Connection struct {
	ID       string         `json:"id" bson:"_id"`
	TenantID string         `json:"tenant_id" bson:"tenant_id"`
	Label    string         `json:"label" bson:"label"`
	Username string         `json:"username" bson:"username"`
	Kind     ConnectionKind `json:"kind" bson:"kind"`
	// Host and Port hold the dial target for Kind == ConnectionKindDirect.
	Host string `json:"host" bson:"host"`
	Port int    `json:"port" bson:"port"`
	// DeviceUID references the target device for Kind == ConnectionKindDevice.
	DeviceUID string    `json:"device_uid" bson:"device_uid"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
