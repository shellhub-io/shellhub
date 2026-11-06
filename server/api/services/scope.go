package services

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
)

const reasonProvisioningKeyTenant = "enrolling with a provisioning key alone, whose digest provisioning_keys_key_digest_unique makes name exactly one namespace"

// BoundTo bounds an operation to the tenant a request carries. An absent tenant refuses the request
// rather than widening it to every namespace, matching the tenant-guard middleware's fail-closed
// behaviour at the route edge.
func BoundTo(tenantID string) (scope.Scope, error) {
	sc, err := scope.NewBounded(tenantID)
	if err != nil {
		return scope.Scope{}, NewErrForbidden(ErrForbidden, err)
	}

	return sc, nil
}

func withinScope(sc scope.Scope, tenantID string) error {
	switch {
	case !sc.IsValid():
		return NewErrForbidden(ErrForbidden, nil)
	case sc.IsBounded() && sc.TenantID() != tenantID:
		return NewErrNamespaceNotFound(tenantID, nil)
	default:
		return nil
	}
}

func anyResolveError(error) bool {
	return true
}

func missingDocument(err error) bool {
	return errors.Is(err, store.ErrNoDocuments)
}

func (s *service) resolveWithin(ctx context.Context, sc scope.Scope, tenantID string, notFound func(error) bool) (*models.Namespace, error) {
	if err := withinScope(sc, tenantID); err != nil {
		return nil, err
	}

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	switch {
	case err == nil && namespace != nil:
		return namespace, nil
	case err == nil || notFound(err):
		return nil, NewErrNamespaceNotFound(tenantID, err)
	default:
		return nil, err
	}
}
