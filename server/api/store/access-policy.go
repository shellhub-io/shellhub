package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type AccessPolicyResolver int

const (
	AccessPolicyIDResolver AccessPolicyResolver = iota
)

type AccessPolicyStore interface {
	// AccessPolicyList retrieves a list of access policies scoped to a namespace.
	AccessPolicyList(ctx context.Context, sc scope.Scope, opts ...QueryOption) ([]models.AccessPolicy, int, error)
	// AccessPolicyResolve retrieves an access policy by the given resolver type and value, scoped to a namespace.
	AccessPolicyResolve(ctx context.Context, sc scope.Scope, resolver AccessPolicyResolver, value string, opts ...QueryOption) (*models.AccessPolicy, error)
	// AccessPolicyCreate creates a new access policy and returns its ID.
	AccessPolicyCreate(ctx context.Context, accessPolicy *models.AccessPolicy) (string, error)
	// AccessPolicyUpdate updates an existing access policy.
	AccessPolicyUpdate(ctx context.Context, accessPolicy *models.AccessPolicy) error
	// AccessPolicyDelete removes an access policy.
	AccessPolicyDelete(ctx context.Context, accessPolicy *models.AccessPolicy) error
}
