package models

import "time"

type TagTarget int

const (
	TagTargetDevice TagTarget = iota + 1
	TagTargetPublicKey
	TagTargetFirewallRule
)

func TagTargets() []TagTarget {
	return []TagTarget{TagTargetDevice, TagTargetPublicKey, TagTargetFirewallRule}
}

type Tag struct {
	ID        string    `json:"-" bson:"_id"`
	CreatedAt time.Time `json:"-" bson:"created_at"`
	UpdatedAt time.Time `json:"-" bson:"updated_at"`
	Name      string    `json:"name" bson:"name"`
	TenantID  string    `json:"-" bson:"tenant_id"`
}

type TagChanges struct {
	Name string `bson:"name,omitempty"`
}

type TagConflicts struct {
	Name string
}
