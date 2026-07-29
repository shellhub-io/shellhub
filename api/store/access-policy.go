package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type AccessPolicyResolver int

const (
	AccessPolicyIDResolver AccessPolicyResolver = iota
)

type AccessPolicyStore interface {
	// AccessPolicyList retrieves a list of access policies with optional filtering and pagination.
	AccessPolicyList(ctx context.Context, opts ...QueryOption) ([]models.AccessPolicy, int, error)
	// AccessPolicyResolve retrieves an access policy by the given resolver type and value.
	AccessPolicyResolve(ctx context.Context, resolver AccessPolicyResolver, value string, opts ...QueryOption) (*models.AccessPolicy, error)
	// AccessPolicyCreate creates a new access policy and returns its ID.
	AccessPolicyCreate(ctx context.Context, accessPolicy *models.AccessPolicy) (string, error)
	// AccessPolicyUpdate updates an existing access policy.
	AccessPolicyUpdate(ctx context.Context, accessPolicy *models.AccessPolicy) error
	// AccessPolicyDelete removes an access policy.
	AccessPolicyDelete(ctx context.Context, accessPolicy *models.AccessPolicy) error
}
