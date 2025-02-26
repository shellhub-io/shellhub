package models

import "time"

// Taggable is an embeddable struct that adds tagging capability to other models.
//
// Example usage:
//
//	type Device struct {
//	    Taggable    // Embed the Taggable struct
//	    Name string // Other device fields
//	}
type Taggable struct {
	// TagsID contains the IDs of associated tags. It is used only for database storage
	// and relationship management. The field is not exposed in JSON responses to keep
	// the API focused on meaningful tag data rather than internal identifiers.
	TagsID []string `json:"-" bson:"tags,omitempty"`

	// Tags contains the complete Tag objects associated with this resource. This field
	// is populated from TagsID when retrieving data from the database, but is not
	// stored directly. It is used only for JSON serialization to provide clients
	// with full tag information.
	Tags []Tag `json:"tags" bson:"-"`
}

type Tag struct {
	ID        string    `json:"-" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	TenantID  string    `json:"tenant_id" bson:"tenant_id"`
}

type TagChanges struct {
	Name string `bson:"name,omitempty"`
}

type TagConflicts struct {
	Name string
}

type TagTarget int

const (
	TagTargetDevice TagTarget = iota + 1
	TagTargetPublicKey
	TagTargetFirewallRule
)

func TagTargets() []TagTarget {
	return []TagTarget{TagTargetDevice, TagTargetPublicKey, TagTargetFirewallRule}
}
