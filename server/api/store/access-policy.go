package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// AccessPolicyResolver names the field an access policy is looked up by.
type AccessPolicyResolver int

// The fields an access policy can be resolved by.
const (
	AccessPolicyIDResolver AccessPolicyResolver = iota
)

// AccessPolicyStore persists the rules that decide who may reach which device as whom.
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
