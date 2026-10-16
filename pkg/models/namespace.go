package models

import "time"

// Namespace is a tenant: the unit devices, members and billing all hang off. Its TenantID is what
// every scoped query filters on, and its Name is what an SSH address resolves through, so the two
// are equally load-bearing and only the name may change.
type Namespace struct {
	Name     string             `json:"name"  validate:"required,hostname_rfc1123,excludes=.,lowercase"`
	Owner    string             `json:"owner"`
	TenantID string             `json:"tenant_id"`
	Members  []Member           `json:"members"`
	Settings *NamespaceSettings `json:"settings"`

	DevicesAcceptedCount int64 `json:"devices_accepted_count"`
	DevicesPendingCount  int64 `json:"devices_pending_count"`
	DevicesRejectedCount int64 `json:"devices_rejected_count"`
	DevicesRemovedCount  int64 `json:"devices_removed_count"`

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

// DeviceLimit returns the namespace's ceiling paired with the count it is measured against, so a
// caller asking both questions reads one consistent snapshot.
func (n *Namespace) DeviceLimit() NamespaceDeviceLimit {
	return NamespaceDeviceLimit{
		MaxDevices:           n.MaxDevices,
		DevicesAcceptedCount: n.DevicesAcceptedCount,
	}
}

// HasMaxDevices reports whether the namespace has a finite device ceiling at all. Community
// instances and unbilled namespaces carry -1, which is no ceiling.
func (n *Namespace) HasMaxDevices() bool {
	return n.DeviceLimit().HasMax()
}

// HasMaxDevicesReached reports whether the ceiling is used up. Ask HasMaxDevices first: with no
// ceiling this compares against -1 and answers true.
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

// NamespaceSettings is the per-namespace policy the SSH gateway reads on every connection. A nil
// Settings on a Namespace means the defaults, not "everything off".
type NamespaceSettings struct {
	SessionRecord          bool   `json:"session_record"`
	ConnectionAnnouncement string `json:"connection_announcement"`
	Legacy                 bool   `json:"legacy"`
}

// IsLegacy reports whether the namespace still runs the legacy model. It is nil-safe: nil
// settings mean a new namespace.
func (s *NamespaceSettings) IsLegacy() bool {
	return s != nil && s.Legacy
}

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
