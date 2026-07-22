package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type SSHIdentityResolver int

const (
	// SSHIdentityIDResolver resolves an identity by its id.
	SSHIdentityIDResolver SSHIdentityResolver = iota
	// SSHIdentityFingerprintResolver resolves an identity by its fingerprint;
	// scope it to a namespace with InNamespace since fingerprints are unique only
	// within a namespace.
	SSHIdentityFingerprintResolver
)

type SSHIdentityStore interface {
	// SSHIdentityList retrieves enrolled SSH identities with optional filtering
	// (InNamespace, WithUserID).
	SSHIdentityList(ctx context.Context, opts ...QueryOption) ([]models.SSHIdentity, int, error)
	// SSHIdentityResolve retrieves an SSH identity by the given resolver type and
	// value.
	SSHIdentityResolve(ctx context.Context, resolver SSHIdentityResolver, value string, opts ...QueryOption) (*models.SSHIdentity, error)
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
}
