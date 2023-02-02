package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

func (s *service) NamespaceCreate(namespace, username, tenant string) (*models.Namespace, error) {
	ctx := context.Background()

	// tenant is optional.
	if tenant == "" {
		tenant = uuid.Generate()
	}

	user, err := s.store.UserGetByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns := &models.Namespace{
		Name:     namespace,
		Owner:    user.ID,
		TenantID: tenant,
		MaxDevices: func() int {
			if envs.IsCloud() {
				return 3
			} else if envs.IsEnterprise() {
				return -1
			}

			return 0
		}(),
		Members: []models.Member{
			{
				ID:   user.ID,
				Role: guard.RoleOwner,
			},
		},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}

	_, err = validator.ValidateStruct(ns)
	if err != nil {
		return nil, ErrNamespaceInvalid
	}

	ns, err = s.store.NamespaceCreate(ctx, ns)
	if err != nil {
		return nil, ErrDuplicateNamespace
	}

	return ns, nil
}

func (s *service) NamespaceAddMember(username, namespace, role string) (*models.Namespace, error) {
	ctx := context.Background()

	if _, err := validator.ValidateStruct(models.Member{Username: username, Role: role}); err != nil {
		return nil, ErrInvalidFormat
	}

	user, err := s.store.UserGetByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns, err := s.store.NamespaceGetByName(ctx, namespace)
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	ns, err = s.store.NamespaceAddMember(ctx, ns.TenantID, user.ID, role)
	if err != nil {
		return nil, ErrFailedNamespaceAddMember
	}

	return ns, nil
}

func (s *service) NamespaceRemoveMember(username, namespace string) (*models.Namespace, error) {
	ctx := context.Background()

	if _, err := validator.ValidateVar(username, "username"); err != nil {
		return nil, ErrInvalidFormat
	}

	user, err := s.store.UserGetByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns, err := s.store.NamespaceGetByName(ctx, namespace)
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	ns, err = s.store.NamespaceRemoveMember(ctx, ns.TenantID, user.ID)
	if err != nil {
		return nil, ErrFailedNamespaceRemoveMember
	}

	return ns, nil
}

func (s *service) NamespaceDelete(namespace string) error {
	ctx := context.Background()

	ns, err := s.store.NamespaceGetByName(ctx, namespace)
	if err != nil {
		return ErrNamespaceNotFound
	}

	if err := s.store.NamespaceDelete(ctx, ns.TenantID); err != nil {
		return ErrFailedDeleteNamespace
	}

	return nil
}
