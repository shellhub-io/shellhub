package models

import "time"

// default Announcement Message for the shellhub namespace
const DefaultCommunityNamespaceAnnouncement = `
******************************************************************
*                                                                *
*             Welcome to ShellHub Community Edition!             *
*                                                                *
* ShellHub is a next-generation SSH server, providing a          *
* seamless, secure, and user-friendly solution for remote        *
* access management. With ShellHub, you can manage all your      *
* devices effortlessly from a single platform, ensuring optimal  *
* security and productivity.                                     *
*                                                                *
* Want to learn more about ShellHub and explore other editions?  *
* Visit: https://shellhub.io                                     *
*                                                                *
* Join our community and contribute to our open-source project:  *
* https://github.com/shellhub-io/shellhub                        *
*                                                                *
* For assistance, please contact the system administrator.       *
*                                                                *
******************************************************************
`

type NamespaceType string

const (
	NamespaceTypePersonal NamespaceType = "personal"
	NamespaceTypeTeam     NamespaceType = "team"
)

func NamespaceTypeFromString(base string) NamespaceType {
	switch base {
	case "personal":
		return NamespaceTypePersonal
	case "team":
		return NamespaceTypeTeam
	default:
		panic("invalid type") // TODO: refactor it
	}
}

type Namespace struct {
	ID          string            `json:"tenant_id" bun:"id,pk,type:uuid"`
	CreatedAt   time.Time         `json:"created_at" bun:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" bun:"updated_at"`
	Type        NamespaceType     `json:"type" bun:"scope"`
	Name        string            `json:"name" bun:"name"`
	Settings    NamespaceSettings `json:"settings" bun:"embed:"`
	Memberships []Membership      `json:"members" bun:"rel:has-many,join:id=namespace_id"`

	Billing *Billing `json:"billing" bun:"-"`
}

type NamespaceSettings struct {
	MaxDevices             int    `json:"max_devices" bun:"max_devices"`
	SessionRecord          bool   `json:"session_record" bun:"record_sessions"`
	ConnectionAnnouncement string `json:"connection_announcement" bun:"connection_announcement,type:text,nullzero"`
}

// HasMaxDevices checks if the namespace has a maximum number of devices.
//
// Generally, a namespace has a MaxDevices value greater than 0 when the ShellHub is either in community version or
// the namespace does not have a billing plan enabled, because, in this case, we set this value to -1.
func (n *Namespace) HasMaxDevices() bool {
	// return n.MaxDevices > 0
	return false
}

// HasMaxDevicesReached checks if the namespace has reached the maximum number of devices.
func (n *Namespace) HasMaxDevicesReached() bool {
	// return n.DevicesCount >= n.MaxDevices
	return false
}

// HasLimitDevicesReached checks if the namespace limit was reached using the removed devices collection.
//
// This method is intended to be run only when the ShellHub instance is Cloud.
func (n *Namespace) HasLimitDevicesReached(removed int64) bool {
	// return int64(n.DevicesCount)+removed >= int64(n.MaxDevices)
	return false
}

// FindMember checks if a member with the specified ID exists in the namespace.
func (n *Namespace) FindMember(id string) (*Membership, bool) {
	for _, member := range n.Memberships {
		if member.UserID == id {
			return &member, true
		}
	}

	return nil, false
}

// NamespaceConflicts holds user attributes that must be unique for each itam and can be utilized in queries
// to identify conflicts.
type NamespaceConflicts struct {
	Name string
}

// Distinct removes the c attributes whether it's equal to the namespace's attributes.
func (c *NamespaceConflicts) Distinct(namespace *Namespace) {
	if c.Name == namespace.Name {
		c.Name = ""
	}
}
