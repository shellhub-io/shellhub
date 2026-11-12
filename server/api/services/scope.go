package services

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
)

const (
	reasonAPIKeyDigest          = "authenticating an API key by its digest, which api_keys_key_digest_unique makes name exactly one namespace"
	reasonDeviceCleanup         = "device-cleanup cron sweeps every namespace, bucketing its deletions per namespace afterwards"
	reasonEphemeralCleanup      = "the ephemeral cleanup removes expired devices of every namespace, bucketing its deletions per namespace afterwards"
	reasonHeartbeatBatch        = "a heartbeat batch spans every namespace holding a tunnel, and each beat is matched within its own"
	reasonPairingByPublicKey    = "pairing by public key: the device has not been placed in a namespace yet, and possession of the matching private key is still required"
	reasonPresenceSweep         = "the presence sweep closes the unwritten tunnel closes of every namespace"
	reasonProvisioningKeyTenant = "enrolling with a provisioning key alone, whose digest provisioning_keys_key_digest_unique makes name exactly one namespace"
	reasonSessionReadBack       = "reading back the session this call just created, by its generated UID"
	reasonTenantlessDeviceToken = "a device token minted before tokens carried a tenant names the device by UID alone, the devices table's primary key, and the server's signature vouches for that UID"
)

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
