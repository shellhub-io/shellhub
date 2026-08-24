package models

import "time"

type Namespace struct {
	Name     string             `json:"name"  validate:"required,hostname_rfc1123,excludes=.,lowercase"`
	Owner    string             `json:"owner"`
	TenantID string             `json:"tenant_id"`
	Members  []Member           `json:"members"`
	Settings *NamespaceSettings `json:"settings"`
	Devices  int                `json:"-"`

	DevicesAcceptedCount int64 `json:"devices_accepted_count"`
	DevicesPendingCount  int64 `json:"devices_pending_count"`
	DevicesRejectedCount int64 `json:"devices_rejected_count"`
	DevicesRemovedCount  int64 `json:"devices_removed_count"`

	Sessions   int       `json:"-"`
	MaxDevices int       `json:"max_devices"`
	CreatedAt  time.Time `json:"created_at"`
	Billing    *Billing  `json:"billing"`
	Type       Type      `json:"type"`
}

// NamespaceDeviceLimit is a namespace's device ceiling and the accepted-device count it is
// measured against, carrying the rule so it stays defined once.
type NamespaceDeviceLimit struct {
	MaxDevices           int
	DevicesAcceptedCount int64
}

// HasMax reports whether the namespace has a finite device ceiling.
//
// Generally, a namespace has a MaxDevices value greater than 0 when the ShellHub is either in community version or
// the namespace does not have a billing plan enabled, because, in this case, we set this value to -1.
func (l NamespaceDeviceLimit) HasMax() bool {
	return l.MaxDevices > 0
}

// IsReached reports whether the accepted-device count has caught up with the ceiling.
// Only counts accepted devices. Removed devices no longer count towards the limit,
// allowing immediate slot reuse after deletion.
func (l NamespaceDeviceLimit) IsReached() bool {
	return l.DevicesAcceptedCount >= int64(l.MaxDevices)
}

func (n *Namespace) DeviceLimit() NamespaceDeviceLimit {
	return NamespaceDeviceLimit{
		MaxDevices:           n.MaxDevices,
		DevicesAcceptedCount: n.DevicesAcceptedCount,
	}
}

func (n *Namespace) HasMaxDevices() bool {
	return n.DeviceLimit().HasMax()
}

func (n *Namespace) HasMaxDevicesReached() bool {
	return n.DeviceLimit().IsReached()
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

// SSHAccessMode selects how a namespace authorizes SSH access.
const (
	// SSHAccessModeLegacy is the key/firewall model: access is a public key with
	// an ACL plus, on Cloud/Enterprise, firewall rules.
	SSHAccessModeLegacy = "legacy"
	// SSHAccessModeIdentity is the identity model: access is a ShellHub identity
	// (established by the out-of-band browser approval) plus Access Policies
	// deciding who may reach what, as which login. The legacy key ACL and
	// firewall checks are bypassed.
	SSHAccessModeIdentity = "identity"
)

type NamespaceSettings struct {
	SessionRecord          bool   `json:"session_record"`
	ConnectionAnnouncement string `json:"connection_announcement"`
	// SSHAccessMode selects the SSH authorization model for the namespace. In
	// "identity" mode every SSH login is gated on an out-of-band browser approval
	// (no device credential required) and governed by Access Policies; the legacy
	// key ACL and firewall checks are bypassed. "legacy" keeps the key/firewall
	// behavior unchanged. New namespaces are born "identity"; namespaces that
	// predate identity-first default to "legacy".
	SSHAccessMode string `json:"ssh_access_mode"`
	// SSHLegacyAllowed marks a namespace that predates identity-first
	// (grandfathered): only these may switch the SSH access mode back to legacy.
	// Namespaces born identity have it false and can never leave identity mode.
	SSHLegacyAllowed bool `json:"ssh_legacy_allowed"`
}

// IsIdentityAccess reports whether the namespace uses the identity-based SSH
// access mode. It is nil-safe so call sites can use it without a prior guard.
func (s *NamespaceSettings) IsIdentityAccess() bool {
	return s != nil && s.SSHAccessMode == SSHAccessModeIdentity
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

// NamespaceConflicts holds namespace attributes that must be unique for each document and can be utilized in queries
// to identify conflicts.
type NamespaceConflicts struct {
	Name string
}

// Distinct removes the c attributes whether it's equal to the namespace attribute.
func (c *NamespaceConflicts) Distinct(namespace *Namespace) {
	if c.Name == namespace.Name {
		c.Name = ""
	}
}
