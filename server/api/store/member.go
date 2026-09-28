package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// MemberDeparture shapes what [MemberStore.NamespaceDepartMember] leaves behind.
type MemberDeparture struct {
	// KeepMembership leaves the membership in place, which is how a demotion departs.
	KeepMembership bool
	// KeepDevices names devices the member owns that become team devices instead of leaving.
	KeepDevices []string
}

// MemberStore persists namespace membership: who belongs to a namespace, and in what role.
type MemberStore interface {
	NamespaceCreateMembership(ctx context.Context, sc scope.Scope, member *models.Member) error
	NamespaceUpdateMembership(ctx context.Context, sc scope.Scope, member *models.Member) error
	NamespaceDeleteMembership(ctx context.Context, sc scope.Scope, member *models.Member) error

	// NamespaceDepartMember ends a member's standing in the namespace as one unit: the accepted
	// devices they own are removed with their owner cleared, the API keys they created are deleted
	// and, unless departure keeps it, the membership is deleted. It joins the transaction the
	// context carries and opens its own otherwise.
	//
	// It returns [ErrDeviceNotOwned] when departure keeps a device the member does not own, and
	// [ErrNoDocuments] when the membership to delete does not exist; either changes nothing.
	NamespaceDepartMember(ctx context.Context, sc scope.Scope, memberID string, departure MemberDeparture) (*models.MemberDeparted, error)
}
