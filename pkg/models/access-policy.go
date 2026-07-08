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

// AccessPolicy is a namespace-scoped, default-deny authorization grant for the
// identity-based SSH access mode: it grants a subject (user, role, or all
// members) the ability to reach the devices selected by Filter, as the unix
// logins listed in Logins. Policies can only grant; there are no deny rules and
// no priority. Access is authorized iff at least one policy grants it.
type AccessPolicy struct {
	ID       string          `json:"id"`
	TenantID string          `json:"-"`
	Name     string          `json:"name"`
	Subject  PolicySubject   `json:"subject"`
	Filter   PublicKeyFilter `json:"filter"`
	// Logins are the unix logins this policy grants: exact names, or ["*"] for
	// any login.
	Logins []string `json:"logins"`
	// RequireStepUp gates access granted by this policy on an extra per-session
	// browser approval (JIT step-up), even when the connecting key is already
	// enrolled. Off by default; enrollment alone is the norm.
	RequireStepUp bool `json:"require_step_up"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Decision is the outcome of an Access Policy authorization check.
type Decision struct {
	Allowed bool `json:"allowed"`
	// RequireStepUp is set when access is allowed by a policy that carries the
	// step-up flag; the gateway must run a per-session browser approval before
	// proceeding.
	RequireStepUp bool   `json:"require_step_up"`
	Reason        string `json:"reason"`
}
