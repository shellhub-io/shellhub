package main

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type Service interface {
	UserCreate(username, password, email string) (*models.User, error)
	UserDelete(username string) error
	UserUpdate(username, password string) error
	NamespaceCreate(namespace, username, tenant string) (*models.Namespace, error)
	NamespaceAddMember(username, namespace, role string) (*models.Namespace, error)
	NamespaceRemoveMember(username, namespace string) (*models.Namespace, error)
	NamespaceDelete(namespace string) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) UserCreate(username, password, email string) (*models.User, error) {
	ctx := context.Background()

	// returnDuplicatedField checks user's name and user's email already exist in the database.
	returnDuplicatedField := func(ctx context.Context, username, email string) error {
		list, _, err := s.store.UserList(ctx, paginator.Query{Page: -1, PerPage: -1}, nil)
		if err != nil {
			return err
		}

		var errs [2]error
		for _, item := range list {
			if item.Username == username {
				errs[0] = ErrUserNameExists
			}

			if item.Email == email {
				errs[1] = ErrUserEmailExists
			}
		}

		switch errs {
		case [2]error{ErrUserNameExists, nil}:
			return ErrUserNameExists
		case [2]error{nil, ErrUserEmailExists}:
			return ErrUserEmailExists
		case [2]error{ErrUserNameExists, ErrUserEmailExists}:
			return ErrUserNameAndEmailExists
		}

		return nil
	}

	name := normalizeField(username)
	mail := normalizeField(email)

	userData := models.UserData{
		Name:     name,
		Email:    mail,
		Username: name,
	}

	_, err := validator.ValidateStruct(userData)
	if err != nil {
		return nil, ErrUserDataInvalid
	}

	if ok := validator.ValidateFieldPassword(password); !ok {
		return nil, ErrUserPasswordInvalid
	}

	userPass := models.UserPassword{
		Password: hashPassword(password),
	}

	user := &models.User{
		UserData:     userData,
		UserPassword: userPass,
		Confirmed:    true,
		CreatedAt:    clock.Now(),
	}

	err = s.store.UserCreate(ctx, user)
	if err != nil {
		if err == store.ErrDuplicate {
			return nil, returnDuplicatedField(ctx, user.Username, user.Email)
		}

		return nil, ErrCreateNewUser
	}

	return user, nil
}

func (s *service) UserDelete(username string) error {
	ctx := context.Background()

	// Gets the user data.
	user, err := s.store.UserGetByUsername(ctx, username)
	if err != nil {
		return ErrUserNotFound
	}

	// Gets data about the namespaces what the user is either member or owner.
	detach, err := s.store.UserDetachInfo(ctx, user.ID)
	if err != nil {
		return ErrNamespaceNotFound
	}
	// Owned namespaces.
	owned := detach["owner"]
	// Joined namespaces.
	joined := detach["member"]

	// Delete all namespaces what the user is member.
	for _, o := range owned {
		if err := s.store.NamespaceDelete(ctx, o.TenantID); err != nil {
			return err
		}
	}

	// Remove user from all namespaces what it is a member.
	for _, m := range joined {
		if _, err := s.store.NamespaceRemoveMember(ctx, m.TenantID, user.ID); err != nil {
			return err
		}
	}

	// Delete the user.
	if err := s.store.UserDelete(ctx, user.ID); err != nil {
		return ErrFailedDeleteUser
	}

	return nil
}

func (s *service) UserUpdate(username, password string) error {
	ctx := context.Background()

	ok := validator.ValidateFieldPassword(password)
	if !ok {
		return ErrUserPasswordInvalid
	}

	passHash := hashPassword(password)

	user, err := s.store.UserGetByUsername(ctx, username)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.store.UserUpdatePassword(ctx, passHash, user.ID); err != nil {
		return ErrFailedUpdateUser
	}

	return nil
}

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
				Role: authorizer.MemberRoleOwner,
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

	if !validator.ValidateFieldUsername(username) {
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
