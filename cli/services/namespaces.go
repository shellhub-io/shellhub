package services

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/cli/pkg/inputs"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

// NamespaceCreate initializes a new namespace, making the specified user its owner.
// The tenant defaults to a UUID if not provided.
// Max device limit is based on the envs.IsCloud() setting.
func (s *service) NamespaceCreate(ctx context.Context, input *inputs.NamespaceCreate) (*models.Namespace, error) {
	// tenant is optional.
	if input.TenantID == "" {
		input.TenantID = uuid.Generate()
	}

	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return nil, ErrNamespaceInvalid
	}

	user, err := s.store.UserResolve(ctx, store.UserUsernameResolver, strings.ToLower(input.Owner))
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns := &models.Namespace{
		Name:                 input.Namespace,
		Owner:                user.ID,
		TenantID:             input.TenantID,
		MaxDevices:           getMaxDevices(),
		DevicesAcceptedCount: 0,
		DevicesPendingCount:  0,
		DevicesRejectedCount: 0,
		DevicesRemovedCount:  0,
		Members: []models.Member{
			{
				ID:      user.ID,
				Role:    authorizer.RoleOwner,
				AddedAt: clock.Now(),
			},
		},
		Settings: &models.NamespaceSettings{
			SessionRecord:          true,
			ConnectionAnnouncement: models.DefaultAnnouncementMessage,
			AllowPassword:          true,
			AllowPublicKey:         true,
			AllowRoot:              true,
			AllowEmptyPasswords:    true,
			AllowTTY:               true,
			AllowTCPForwarding:     true,
			AllowWebEndpoints:      true,
			AllowSFTP:              true,
			AllowAgentForwarding:   true,
		},
		CreatedAt: clock.Now(),
		Type:      models.NewDefaultType(),
	}

	if models.IsTypeTeam(input.Type) {
		ns.Type = models.TypeTeam
	} else if models.IsTypePersonal(input.Type) {
		ns.Type = models.TypePersonal
	}

	if _, err = s.store.NamespaceCreate(ctx, ns); err != nil {
		return nil, ErrDuplicateNamespace
	}

	return ns, nil
}

// NamespaceAddMember adds a new member with a specified role to a namespace.
func (s *service) NamespaceAddMember(ctx context.Context, input *inputs.MemberAdd) (*models.Namespace, error) {
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return nil, ErrInvalidFormat
	}

	user, err := s.store.UserResolve(ctx, store.UserUsernameResolver, strings.ToLower(input.Username))
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns, err := s.store.NamespaceResolve(ctx, store.NamespaceNameResolver, strings.ToLower(input.Namespace))
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	if err = s.store.NamespaceCreateMembership(ctx, ns.TenantID, &models.Member{
		ID:      user.ID,
		Role:    input.Role,
		AddedAt: clock.Now(),
	}); err != nil {
		return nil, ErrFailedNamespaceAddMember
	}

	return ns, nil
}

// NamespaceRemoveMember removes a member from a namespace.
func (s *service) NamespaceRemoveMember(ctx context.Context, input *inputs.MemberRemove) (*models.Namespace, error) {
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return nil, ErrInvalidFormat
	}

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

	if err = s.store.NamespaceDeleteMembership(ctx, ns.TenantID, member); err != nil {
		return nil, ErrFailedNamespaceRemoveMember
	}

	return ns, nil
}

// NamespaceDelete deletes a namespace based on the provided namespace name.
func (s *service) NamespaceDelete(ctx context.Context, input *inputs.NamespaceDelete) error {
	if ok, err := s.validator.Struct(input); !ok || err != nil {
		return ErrNamespaceInvalid
	}

	ns, err := s.store.NamespaceResolve(ctx, store.NamespaceNameResolver, strings.ToLower(input.Namespace))
	if err != nil {
		return ErrNamespaceNotFound
	}

	if err := s.store.NamespaceDelete(ctx, ns); err != nil {
		return ErrFailedDeleteNamespace
	}

	return nil
}

// NamespaceList retrieves all namespaces available to the user
func (s *service) NamespaceList(ctx context.Context) ([]models.Namespace, error) {
	namespaces, _, err := s.store.NamespaceList(ctx)
	if err != nil {
		return nil, ErrFailedListNamespace
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
