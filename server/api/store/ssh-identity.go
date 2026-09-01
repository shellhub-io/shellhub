package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// SSHIdentityResolver names the field an SSH identity is looked up by.
type SSHIdentityResolver int

const (
	// SSHIdentityIDResolver resolves an identity by its id.
	SSHIdentityIDResolver SSHIdentityResolver = iota
	// SSHIdentityFingerprintResolver resolves an identity by its fingerprint.
	// Fingerprints are unique only within a namespace, so the scope must be bounded.
	SSHIdentityFingerprintResolver
)

// SSHIdentityStore persists the enrolled public keys that identify a person to a device.
type SSHIdentityStore interface {
	// SSHIdentityList retrieves enrolled SSH identities scoped to a namespace.
	SSHIdentityList(ctx context.Context, sc scope.Scope, opts ...QueryOption) ([]models.SSHIdentity, int, error)
	// SSHIdentityResolve retrieves an SSH identity by the given resolver type and value, scoped to a namespace.
	SSHIdentityResolve(ctx context.Context, sc scope.Scope, resolver SSHIdentityResolver, value string, opts ...QueryOption) (*models.SSHIdentity, error)
	// SSHIdentityCreate creates a new SSH identity and returns its id.
	SSHIdentityCreate(ctx context.Context, identity *models.SSHIdentity) (string, error)
	// SSHIdentityUpdate renames an existing SSH identity scoped to its namespace.
	SSHIdentityUpdate(ctx context.Context, identity *models.SSHIdentity) error
	// SSHIdentityDelete removes an SSH identity scoped to its namespace.
	SSHIdentityDelete(ctx context.Context, identity *models.SSHIdentity) error
	// SSHIdentityTouchLastUsed stamps the last-used time of the identity matching
	// the namespace and fingerprint. A miss is not an error.
	SSHIdentityTouchLastUsed(ctx context.Context, tenantID, fingerprint string) error
	// SSHIdentityTouchReauth stamps the last-reauth time of the identity matching
	// the namespace and fingerprint, on a successful re-authentication. A miss is
	// not an error.
	SSHIdentityTouchReauth(ctx context.Context, tenantID, fingerprint string) error
	// SSHIdentityConsume atomically burns a single-use identity, stamping
	// consumed_at only if it was still null. It returns true when this call won
	// the burn (and false when the key was already consumed or absent), so
	// concurrent single-use sessions resolve to exactly one winner.
	SSHIdentityConsume(ctx context.Context, tenantID, fingerprint string) (bool, error)
}
