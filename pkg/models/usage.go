package models

// InstanceUsage counts what the whole instance holds, for the telemetry heartbeat.
type InstanceUsage struct {
	Namespaces int64       `json:"namespaces"`
	Users      int64       `json:"users"`
	Devices    DeviceUsage `json:"devices"`
	// Sessions counts the sessions started since the cutoff the caller passed.
	Sessions int64 `json:"sessions"`

	SessionKinds    SessionKindUsage     `json:"session_kinds"`
	SessionActivity SessionActivityUsage `json:"session_activity"`
	NamespaceKinds  NamespaceKindUsage   `json:"namespace_kinds"`
	DeviceKinds     DeviceKindUsage      `json:"device_kinds"`
	Access          AccessUsage          `json:"access"`
	Authentication  AuthenticationUsage  `json:"authentication"`
}

// SessionKindUsage counts the sessions started since the cutoff by what they did. A session that
// did several things counts under each of them.
type SessionKindUsage struct {
	// Interactive sessions opened a shell on a seat that requested a pty.
	Interactive int64 `json:"interactive"`
	// ShellWithoutPty sessions opened a shell on a seat with no pty, as a heredoc does.
	ShellWithoutPty int64 `json:"shell_without_pty"`
	// Exec sessions ran a command other than scp.
	Exec int64 `json:"exec"`
	// SCP sessions ran scp as a command. An scp client that speaks SFTP counts under SFTP.
	SCP             int64 `json:"scp"`
	SFTP            int64 `json:"sftp"`
	PortForwarding  int64 `json:"port_forwarding"`
	AgentForwarding int64 `json:"agent_forwarding"`
	// Web sessions came from the console's web terminal.
	Web      int64 `json:"web"`
	Recorded int64 `json:"recorded"`
	// APIKey sessions were opened with an API key rather than by a user.
	APIKey int64 `json:"api_key"`
}

// SessionActivityUsage measures the sessions started since the cutoff. A session's duration runs
// from its start to the last time the server saw it, so a session still open counts the time up to
// its latest keep-alive. Durations are in whole seconds and are zero when no session authenticated.
type SessionActivityUsage struct {
	// Unauthenticated sessions never got past authentication.
	Unauthenticated int64 `json:"unauthenticated"`
	// Active counts the sessions open now, whenever they started.
	Active int64 `json:"active"`
	// Devices counts the distinct devices the authenticated sessions reached.
	Devices int64 `json:"devices"`
	// Events counts the events the sessions recorded, the ones past the timeline cap included.
	Events int64 `json:"events"`
	// DurationSeconds sums the durations of the authenticated sessions.
	DurationSeconds       int64 `json:"duration_seconds"`
	DurationMedianSeconds int64 `json:"duration_median_seconds"`
	DurationP95Seconds    int64 `json:"duration_p95_seconds"`
}

// NamespaceKindUsage counts namespaces by how they are set up.
type NamespaceKindUsage struct {
	// Legacy namespaces still use public keys and firewall rules rather than SSH identities and
	// access policies.
	Legacy                   int64 `json:"legacy"`
	Personal                 int64 `json:"personal"`
	Team                     int64 `json:"team"`
	Recording                int64 `json:"recording"`
	RequireVerifiedDeviceKey int64 `json:"require_verified_device_key"`
}

// UsageUnknownValue is the key a usage count by value gives a row whose value is empty.
const UsageUnknownValue = "unknown"

// DeviceKindUsage describes the accepted devices. Platforms, Versions, Architectures and OSes are
// keyed by the value the agent reported, and an empty value is counted under [UsageUnknownValue].
// Those values are free text from the device, so the telemetry heartbeat buckets them into known
// labels before sending them.
type DeviceKindUsage struct {
	// Online devices are accepted and seen within the online cutoff.
	Online        int64            `json:"online"`
	Ephemeral     int64            `json:"ephemeral"`
	Owned         int64            `json:"owned"`
	CustomFields  int64            `json:"custom_fields"`
	Tagged        int64            `json:"tagged"`
	KeyVerified   int64            `json:"key_verified"`
	Platforms     map[string]int64 `json:"platforms"`
	Versions      map[string]int64 `json:"versions"`
	Architectures map[string]int64 `json:"architectures"`
	OSes          map[string]int64 `json:"oses"`
	// EnrolledWith counts accepted devices by the type of provisioning key they enrolled with:
	// user, legacy or pairing.
	EnrolledWith map[string]int64 `json:"enrolled_with"`
}

// AccessUsage counts what controls and grants access.
type AccessUsage struct {
	AccessPolicies  int64 `json:"access_policies"`
	PublicKeys      int64 `json:"public_keys"`
	APIKeys         int64 `json:"api_keys"`
	InstanceAPIKeys int64 `json:"instance_api_keys"`
	Tags            int64 `json:"tags"`
	// SSHIdentities counts SSH identities by source: manual, browser or approval.
	SSHIdentities map[string]int64 `json:"ssh_identities"`
	// ProvisioningKeys counts the user provisioning keys that are not revoked, by mode.
	ProvisioningKeys map[string]int64 `json:"provisioning_keys"`
	// Enrollments counts the enrollment attempts made with a provisioning key since the cutoff,
	// each retry of a refused device included.
	Enrollments        int64 `json:"enrollments"`
	PendingInvitations int64 `json:"pending_invitations"`
}

// AuthenticationUsage describes how people sign in.
type AuthenticationUsage struct {
	LocalEnabled bool `json:"local_enabled"`
	SAMLEnabled  bool `json:"saml_enabled"`
	// ActiveUsers logged in since the cutoff.
	ActiveUsers int64 `json:"active_users"`
	// UsersByOrigin counts users by where their account comes from: local or saml.
	UsersByOrigin map[string]int64 `json:"users_by_origin"`
}

// NamespaceUsage counts what one namespace holds, for the cloud telemetry heartbeat.
type NamespaceUsage struct {
	TenantID string      `json:"-"`
	Members  int64       `json:"members"`
	Devices  DeviceUsage `json:"devices"`
	// Sessions counts the sessions started since the cutoff the caller passed.
	Sessions int64 `json:"sessions"`
}

// DeviceUsage counts devices by status. Removed devices are not counted.
type DeviceUsage struct {
	Accepted int64 `json:"accepted"`
	Pending  int64 `json:"pending"`
	Rejected int64 `json:"rejected"`
}
