package gateway

import (
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// CallerKind names the kind of caller an [Endpoint]'s handler takes. The route table reads it to
// refuse a row whose access admits callers the handler cannot take.
type CallerKind int

// The kinds of caller, one per [Caller] type. CallerUnstated is the kind of [Legacy] and [Raw], whose
// handler resolves the caller itself.
const (
	CallerUnstated CallerKind = iota
	CallerAnonymous
	CallerPrincipal
	CallerMember
	CallerAdmin
)

var callerKindNames = map[CallerKind]string{
	CallerUnstated:  "unstated",
	CallerAnonymous: "anonymous",
	CallerPrincipal: "principal",
	CallerMember:    "member",
	CallerAdmin:     "admin",
}

// String names k in lower case: "anonymous", "principal", "member", "admin" or "unstated".
func (k CallerKind) String() string {
	return callerKindNames[k]
}

// Caller is the set of caller types a typed handler can take. Each names one kind of caller, and
// the shapes refuse a request whose identity is not of that kind before the handler runs.
type Caller interface {
	Anonymous | Principal | Member | Admin
}

// Anonymous is the caller of a route open to anyone. It carries nothing, whether or not the request
// presented a credential.
type Anonymous struct{}

// Member is an acting member: a person whose user token names the namespace the request is bounded
// to. Role is the role authentication resolved for UserID within the scope's namespace, and
// Username is empty when the account signed in through SAML. API keys, devices and the
// administration console are never members.
type Member struct {
	UserID   string
	Username string
	Role     authorizer.Role
}

// Principal is whoever acts within the namespace scope: an acting member or an API key. Exactly one
// of UserID and APIKeyID, the API key's surrogate id, is set, and Username is empty for an API key.
// Role is the member's role, or the role the API key was issued with.
type Principal struct {
	Role     authorizer.Role
	UserID   string
	Username string
	APIKeyID string
}

// Member returns the acting member behind p, and false when an API key is behind it.
func (p Principal) Member() (Member, bool) {
	if p.UserID == "" {
		return Member{}, false
	}

	return Member{UserID: p.UserID, Username: p.Username, Role: p.Role}, true
}

// Admin is an instance administrator: a user holding the admin flag on the administration surface,
// named by UserID, or the instance API key, which names no person and sets InstanceAPIKey instead.
type Admin struct {
	UserID         string
	InstanceAPIKey bool
}

func callerKind[A Caller]() CallerKind {
	var zero A

	switch any(zero).(type) {
	case Anonymous:
		return CallerAnonymous
	case Principal:
		return CallerPrincipal
	case Member:
		return CallerMember
	case Admin:
		return CallerAdmin
	}

	return CallerUnstated
}

func callerOf[A Caller](caller identity.Identity) (A, bool) {
	var (
		zero     A
		resolved any
		ok       bool
	)

	switch any(zero).(type) {
	case Anonymous:
		resolved, ok = Anonymous{}, true
	case Principal:
		resolved, ok = principalOf(caller)
	case Member:
		resolved, ok = memberOf(caller)
	case Admin:
		resolved, ok = adminOf(caller)
	}

	typed, isA := resolved.(A)

	return typed, ok && isA
}

func memberOf(caller identity.Identity) (Member, bool) {
	if caller.ID == "" || caller.APIKey != "" || caller.DeviceUID != "" {
		return Member{}, false
	}

	return Member{UserID: caller.ID, Username: caller.Username, Role: caller.Role}, true
}

func principalOf(caller identity.Identity) (Principal, bool) {
	if member, ok := memberOf(caller); ok {
		return Principal{Role: member.Role, UserID: member.UserID, Username: member.Username}, true
	}

	if caller.APIKey == "" || caller.Admin || caller.ID != "" || caller.DeviceUID != "" {
		return Principal{}, false
	}

	return Principal{Role: caller.Role, APIKeyID: caller.APIKeyID}, true
}

func adminOf(caller identity.Identity) (Admin, bool) {
	if !caller.Admin || caller.ID != "" || caller.DeviceUID != "" {
		return Admin{}, false
	}

	switch {
	case caller.AdminID != "":
		return Admin{UserID: caller.AdminID}, true
	case strings.HasPrefix(caller.APIKey, models.InstanceAPIKeyPrefix):
		return Admin{InstanceAPIKey: true}, true
	}

	return Admin{}, false
}
