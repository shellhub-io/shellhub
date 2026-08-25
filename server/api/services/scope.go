package services

import (
	"github.com/shellhub-io/shellhub/pkg/api/scope"
)

// reasonInternalSessionMutation is the justification carried by the internal session mutations the
// SSH server drives (deactivate, keep-alive, update, event recording). Their call chains carry no
// namespace at all, so bounding them means changing the SSH-facing contract, which is deliberately
// deferred to its own change.
const reasonInternalSessionMutation = "internal SSH-driven session mutation: no namespace exists anywhere in the call chain yet; bounding it changes the SSH-facing contract (see #6749)"

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
