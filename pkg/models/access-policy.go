package models

import "time"

// PolicySubjectType enumerates who an Access Policy grants access to.
type PolicySubjectType string

const (
	// PolicySubjectUser grants a single user, identified by user id in Value.
	PolicySubjectUser PolicySubjectType = "user"
	// PolicySubjectRole grants every member holding a role, named in Value.
	PolicySubjectRole PolicySubjectType = "role"
	// PolicySubjectAllMembers grants every member of the namespace; Value is empty.
	PolicySubjectAllMembers PolicySubjectType = "all-members"
)

// PolicySubject identifies who an Access Policy grants access to.
type PolicySubject struct {
	Type  PolicySubjectType `json:"type"`
	Value string            `json:"value"`
}

// PolicyAction is whether an Access Policy grants access (allow) or blocks it
// (deny).
type PolicyAction string

const (
	// PolicyActionAllow grants access to the subject; the default.
	PolicyActionAllow PolicyAction = "allow"
	// PolicyActionDeny blocks access. Deny is evaluated before allow and wins
	// over any allow, however specific: it is a subtractive blocklist carved out
	// of the broad grants, not a base layer (default-deny already blocks the rest).
	PolicyActionDeny PolicyAction = "deny"
)

// AccessPolicy is a namespace-scoped authorization rule for the identity-based
// SSH access mode: for a subject (user, role, or all members) reaching the
// devices selected by Filter as the unix logins listed in Logins, it either
// grants (Effect allow) or blocks (Effect deny) access. Evaluation is
// default-deny and deny-first: a matching deny wins over any allow, and access
// is authorized iff some allow grants it and no deny blocks it.
type AccessPolicy struct {
	ID       string          `json:"id"`
	TenantID string          `json:"-"`
	Name     string          `json:"name"`
	Subject  PolicySubject   `json:"subject"`
	Filter   PublicKeyFilter `json:"filter"`
	// Logins are the unix logins this policy covers: exact names, or ["*"] for
	// any login.
	Logins []string `json:"logins"`
	// SourceIP restricts the policy to connections from these CIDRs (a client IP
	// in any of them matches). Empty matches any IP. A single host is a /32 (or
	// /128 for IPv6).
	SourceIP []string `json:"source_ip"`
	// Action is whether this policy grants (allow) or blocks (deny) the covered
	// access. Defaults to allow.
	Action PolicyAction `json:"action"`
	// RequireReauth gates access granted by this policy on a fresh per-session
	// re-authentication (an out-of-band confirmation), even when the connecting
	// key is already enrolled. Off by default; enrollment alone is the norm.
	RequireReauth bool `json:"require_reauth"`
	// ReauthPeriod is the freshness window for RequireReauth, in seconds: a
	// re-authentication is only demanded when the identity has not re-authed
	// within it. nil or 0 means "always" (every session). Only meaningful when
	// RequireReauth is set.
	ReauthPeriod *int `json:"reauth_period"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewOwnerAccessPolicy is the starter policy for the identity access mode: it
// grants the namespace owner every login on every device. Seeded when a
// namespace is born identity (creation) or switches to identity with no
// policies (legacy toggle), so default-deny never locks the owner out while
// every other member starts with no access.
func NewOwnerAccessPolicy(tenantID, ownerID string) *AccessPolicy {
	return &AccessPolicy{
		TenantID: tenantID,
		Name:     "Owner access",
		Subject:  PolicySubject{Type: PolicySubjectUser, Value: ownerID},
		Filter:   PublicKeyFilter{},
		Logins:   []string{"*"},
		SourceIP: []string{},
		Action:   PolicyActionAllow,
	}
}

// Decision is the outcome of an Access Policy authorization check.
type Decision struct {
	Allowed bool `json:"allowed"`
	// RequireReauth is set when access is allowed by a policy that carries the
	// re-auth flag; the gateway must run a fresh per-session re-authentication
	// before proceeding, subject to ReauthPeriod.
	RequireReauth bool `json:"require_reauth"`
	// ReauthPeriod is the matched policy's freshness window in seconds (nil/0 =
	// always). The gateway skips the re-auth when the identity re-authed within
	// it. Only meaningful when RequireReauth is set.
	ReauthPeriod *int   `json:"reauth_period"`
	Reason       string `json:"reason"`
}
