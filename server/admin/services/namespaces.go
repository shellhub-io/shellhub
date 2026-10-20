package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/admin/inputs"
	apiservices "github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/store"
)

func (s *service) maxDevices() int {
	if s.edition.IsCloud() {
		return MaxNumberDevicesLimited
	}

	return MaxNumberDevicesUnlimited
}

// NamespaceCreate initializes a new namespace, making the specified user its owner.
// The tenant defaults to a UUID if not provided.
// Max device limit is based on the service's edition.
func (s *service) NamespaceCreate(ctx context.Context, input *inputs.NamespaceCreate) (*models.Namespace, error) {
	if input.TenantID == "" {
		input.TenantID = uuid.Generate()
	}

	user, err := s.store.UserResolve(ctx, store.UserUsernameResolver, strings.ToLower(input.Owner))
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns := &models.Namespace{
		Name:                 input.Namespace,
		Owner:                user.ID,
		TenantID:             input.TenantID,
		MaxDevices:           s.maxDevices(),
		DevicesAcceptedCount: 0,
		DevicesPendingCount:  0,
		DevicesRejectedCount: 0,
		DevicesRemovedCount:  0,
		Members: []models.Member{
			{
				ID:      user.ID,
				Role:    authorizer.RoleOwner,
				AddedAt: s.clock.Now(),
			},
		},
		Settings: &models.NamespaceSettings{
			SessionRecord:            true,
			ConnectionAnnouncement:   envs.AnnouncementFor(s.edition),
			RequireVerifiedDeviceKey: s.namespaceRequireVerifiedDeviceKey,
		},
		CreatedAt: s.clock.Now(),
		Type:      models.NewDefaultType(),
	}

	if models.IsTypeTeam(input.Type) {
		ns.Type = models.TypeTeam
	} else if models.IsTypePersonal(input.Type) {
		ns.Type = models.TypePersonal
	}

	if _, err = s.store.NamespaceCreate(ctx, ns); err != nil {
		switch {
		case errors.Is(err, store.ErrNamespaceSingle):
			return nil, ErrNamespaceSingle
		case errors.Is(err, store.ErrDuplicate):
			return nil, ErrDuplicateNamespace
		default:
			return nil, fmt.Errorf("%w: %w", ErrCreateNewNamespace, err)
		}
	}

	return ns, nil
}

// NamespaceAddMember adds a new member with a specified role to a namespace.
func (s *service) NamespaceAddMember(ctx context.Context, input *inputs.MemberAdd) (*models.Namespace, error) {
	user, err := s.store.UserResolve(ctx, store.UserUsernameResolver, strings.ToLower(input.Username))
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns, err := s.store.NamespaceResolve(ctx, store.NamespaceNameResolver, strings.ToLower(input.Namespace))
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	if err = s.store.NamespaceCreateMembership(ctx, scope.MustBounded(ns.TenantID), &models.Member{
		ID:      user.ID,
		Role:    input.Role,
		AddedAt: s.clock.Now(),
	}); err != nil {
		return nil, ErrFailedNamespaceAddMember
	}

	return ns, nil
}

// NamespaceRemoveMember removes a member from a namespace, reporting ErrNamespaceRemoveOwner for
// the owner: their membership carries the access policies and SSH identities the namespace granted
// them, and removing it takes those with it.
func (s *service) NamespaceRemoveMember(ctx context.Context, input *inputs.MemberRemove) (*models.Namespace, error) {
	user, err := s.store.UserResolve(ctx, store.UserUsernameResolver, strings.ToLower(input.Username))
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns, err := s.store.NamespaceResolve(ctx, store.NamespaceNameResolver, strings.ToLower(input.Namespace))
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	member, ok := ns.FindMember(user.ID)
	if !ok {
		return nil, ErrFailedNamespaceRemoveMember
	}

	if member.Role == authorizer.RoleOwner {
		return nil, ErrNamespaceRemoveOwner
	}

	if _, err = s.store.NamespaceDepartMember(ctx, scope.MustBounded(ns.TenantID), member.ID, store.MemberDeparture{}); err != nil {
		return nil, ErrFailedNamespaceRemoveMember
	}

	return ns, nil
}

// NamespaceDelete implements [Services].
func (s *service) NamespaceDelete(ctx context.Context, input *inputs.NamespaceDelete) error {
	ns, err := s.store.NamespaceResolve(ctx, store.NamespaceNameResolver, strings.ToLower(input.Namespace))
	if err != nil {
		return ErrNamespaceNotFound
	}

	if err := apiservices.DeleteNamespacesEndingDevices(ctx, s.store, s.cache, []string{ns.TenantID}, func(ctx context.Context) error {
		return s.store.NamespaceDelete(ctx, ns)
	}); err != nil {
		if errors.Is(err, store.ErrNamespaceInstanceProtected) {
			return ErrNamespaceInstanceProtected
		}

		return ErrFailedDeleteNamespace
	}

	return nil
}

// NamespaceLimit implements [Services].
func (s *service) NamespaceLimit(ctx context.Context, input *inputs.NamespaceLimit) (*models.Namespace, error) {
	ns, err := s.store.NamespaceResolve(ctx, store.NamespaceNameResolver, strings.ToLower(input.Namespace))
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	ns.MaxDevices = input.MaxDevices
	if err := s.store.NamespaceUpdate(ctx, ns); err != nil {
		return nil, ErrFailedLimitNamespace
	}

	return ns, nil
}

// NamespaceList retrieves all namespaces available to the user
func (s *service) NamespaceList(ctx context.Context) ([]models.Namespace, error) {
	namespaces, _, err := s.store.NamespaceList(ctx)
	if err != nil {
		return nil, ErrFailedListNamespaces
	}

	return namespaces, nil
}

// NamespaceResolve retrieves a namespace using the specified resolver
func (s *service) NamespaceResolve(ctx context.Context, resolver NamespaceResolver, value string) (*models.Namespace, error) {
	var storeResolver store.NamespaceResolver

	if resolver == NamespaceResolverTenantID {
		storeResolver = store.NamespaceTenantIDResolver
	} else {
		storeResolver = store.NamespaceNameResolver
	}

	namespace, err := s.store.NamespaceResolve(ctx, storeResolver, value)
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	return namespace, nil
}

// NamespaceDeviceCounts returns the actual device counts for a namespace.
func (s *service) NamespaceDeviceCounts(ctx context.Context, tenantID string) (*models.Stats, error) {
	sc, err := scope.NewBounded(tenantID)
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	return s.store.GetStats(ctx, sc)
}
