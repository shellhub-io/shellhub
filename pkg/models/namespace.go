package models

import "time"

type Namespace struct {
	Name     string             `json:"name"  validate:"required,hostname_rfc1123,excludes=.,lowercase"`
	Owner    string             `json:"owner"`
	TenantID string             `json:"tenant_id" bson:"tenant_id,omitempty"`
	Members  []Member           `json:"members" bson:"members"`
	Settings *NamespaceSettings `json:"settings"`
	Devices  int                `json:"-" bson:"devices,omitempty"`

	DevicesAcceptedCount int64 `json:"devices_accepted_count" bson:"devices_accepted_count"`
	DevicesPendingCount  int64 `json:"devices_pending_count" bson:"devices_pending_count"`
	DevicesRejectedCount int64 `json:"devices_rejected_count" bson:"devices_rejected_count"`

	Sessions   int       `json:"-" bson:"sessions,omitempty"`
	MaxDevices int       `json:"max_devices" bson:"max_devices"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	Billing    *Billing  `json:"billing" bson:"billing,omitempty"`
	Type       Type      `json:"type" bson:"type"`
}

// HasMaxDevices checks if the namespace has a maximum number of devices.
//
// Generally, a namespace has a MaxDevices value greater than 0 when the ShellHub is either in community version or
// the namespace does not have a billing plan enabled, because, in this case, we set this value to -1.
func (n *Namespace) HasMaxDevices() bool {
	return n.MaxDevices > 0
}

// HasMaxDevicesReached checks if the namespace has reached the maximum number of devices.
func (n *Namespace) HasMaxDevicesReached() bool {
	return n.DevicesAcceptedCount >= int64(n.MaxDevices)
}

// HasLimitDevicesReached checks if the namespace limit was reached using the removed devices collection.
//
// This method is intended to be run only when the ShellHub instance is Cloud.
func (n *Namespace) HasLimitDevicesReached(removed int64) bool {
	return n.DevicesAcceptedCount+removed >= int64(n.MaxDevices)
}

// FindMember checks if a member with the specified ID exists in the namespace.
func (n *Namespace) FindMember(id string) (*Member, bool) {
	for _, member := range n.Members {
		if member.ID == id {
			return &member, true
		}
	}

	return nil, false
}

type NamespaceSettings struct {
	SessionRecord          bool   `json:"session_record" bson:"session_record,omitempty"`
	ConnectionAnnouncement string `json:"connection_announcement" bson:"connection_announcement"`
}

type NamespaceChanges struct {
	Name                   string  `bson:"name,omitempty"`
	SessionRecord          *bool   `bson:"settings.session_record,omitempty"`
	ConnectionAnnouncement *string `bson:"settings.connection_announcement,omitempty"`
}

// default Announcement Message for the shellhub namespace
const DefaultAnnouncementMessage = `
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
