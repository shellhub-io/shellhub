package services

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/cli/pkg/inputs"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

// NamespaceCreate initializes a new namespace, making the specified user its owner.
// The tenant defaults to a UUID if not provided.
// Max device limit is based on the envs.IsCloud() setting.
func (s *service) NamespaceCreate(ctx context.Context, input *inputs.NamespaceCreate) (*models.Namespace, error) {
	user, err := s.store.UserGet(ctx, store.UserIdentUsername, input.Owner)
	if err != nil {
		return nil, ErrUserNotFound
	}

	input.Namespace = strings.ToLower(input.Namespace)

	if _, has, err := s.store.NamespaceConflicts(ctx, &models.NamespaceConflicts{Name: input.Namespace}); err != nil || has {
		return nil, ErrDuplicateNamespace
	}

	if input.TenantID == "" {
		input.TenantID = uuid.Generate()
	}

	ns := &models.Namespace{
		ID:   input.TenantID,
		Type: models.NamespaceTypeFromString(input.Type),
		Name: input.Namespace,
		Settings: models.NamespaceSettings{
			SessionRecord:          true,
			ConnectionAnnouncement: models.DefaultCommunityNamespaceAnnouncement,
			MaxDevices:             getMaxDevices(),
		},
		Memberships: []models.Membership{
			{
				UserID:      user.ID,
				NamespaceID: input.TenantID,
				Status:      models.MembershipStatusAccepted,
				Role:        authorizer.RoleOwner,
			},
		},
	}

	// TODO: transaction

	if _, err := s.store.NamespaceCreate(ctx, ns); err != nil {
		return nil, err
	}

	if err := s.store.NamespaceCreateMemberships(ctx, ns.Memberships); err != nil {
		return nil, err
	}

	return ns, nil
}

// NamespaceDelete deletes a namespace based on the provided namespace name.
func (s *service) NamespaceDelete(ctx context.Context, input *inputs.NamespaceDelete) error {
	n, err := s.store.NamespaceGet(ctx, store.NamespaceIdentName, input.Namespace)
	if err != nil {
		return ErrNamespaceNotFound
	}

	if err := s.store.Delete(ctx, n.Memberships, n); err != nil {
		return ErrFailedDeleteNamespace
	}

	return nil
}

// NamespaceAddMember adds a new member with a specified role to a namespace.
func (s *service) NamespaceAddMember(ctx context.Context, input *inputs.MemberAdd) (*models.Namespace, error) {
	u, err := s.store.UserGet(ctx, store.UserIdentUsername, input.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	n, err := s.store.NamespaceGet(ctx, store.NamespaceIdentName, input.Namespace)
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	membership := models.Membership{
		UserID:      u.ID,
		NamespaceID: n.ID,
		Status:      models.MembershipStatusAccepted,
		Role:        input.Role,
	}

	if err = s.store.NamespaceCreateMemberships(ctx, []models.Membership{membership}); err != nil {
		return nil, ErrFailedNamespaceAddMember
	}

	return n, nil
}

// NamespaceRemoveMember removes a member from a namespace.
func (s *service) NamespaceRemoveMember(ctx context.Context, input *inputs.MemberRemove) (*models.Namespace, error) {
	// user, err := s.store.UserGet(ctx, store.UserIdentUsername, input.Username)
	// if err != nil {
	// 	return nil, ErrUserNotFound
	// }
	//
	// ns, err := s.store.NamespaceGet(ctx, store.NamespaceIdentName, input.Namespace)
	// if err != nil {
	// 	return nil, ErrNamespaceNotFound
	// }
	//
	// if err = s.store.NamespaceRemoveMember(ctx, ns.ID, user.ID); err != nil {
	// 	return nil, ErrFailedNamespaceRemoveMember
	// }

	return nil, nil
}
