// Package identity carries the caller's authenticated identity on a request's context, from
// the authentication middleware that resolves it to the handlers that act on it.
package identity

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

// Identity is the caller's authenticated identity, as resolved from the request's credential.
type Identity struct {
	ID        string
	Username  string
	TenantID  string
	DeviceUID string
	APIKey    string
	APIKeyID  string
	Role      authorizer.Role
	Admin     bool
	AdminID   string
}

// WithoutUserScope returns the identity stripped of the acting user's ID and namespace scope,
// keeping the admin flag. An administrator's ID moves to AdminID, so an admin route can still
// tell who acted without the ID the guards key on.
//
// The /admin/api surface depends on this shape: the authorization middleware treats a missing
// ID alongside Admin as "this call came from the admin panel rather than from a namespace
// member", and lets it past the tenant guard. An admin hitting the regular /api surface keeps
// their user scope and stays subject to that guard.
func (i *Identity) WithoutUserScope() Identity {
	scoped := *i
	if i.Admin {
		scoped.AdminID = i.ID
	}

	scoped.ID = ""
	scoped.TenantID = ""

	return scoped
}

type contextKey struct{}

// NewContext returns a copy of ctx carrying identity. Every handler downstream trusts it as the
// authenticated caller, so only the authentication middleware calls it outside tests.
func NewContext(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, contextKey{}, identity)
}

// FromContext returns the identity [NewContext] stored on ctx, and false when the request is
// anonymous or never went through authentication.
func FromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(contextKey{}).(Identity)

	return identity, ok
}
