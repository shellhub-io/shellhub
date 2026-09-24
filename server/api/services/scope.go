package services

import (
	"github.com/shellhub-io/shellhub/pkg/api/scope"
)

const reasonInternalSessionMutation = "internal SSH-driven session mutation: no namespace exists anywhere in the call chain yet; bounding it changes the SSH-facing contract (see #6749)"

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
