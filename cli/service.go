package main

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type Service interface {
	UserCreate(username, password, email string) (*models.User, error)
	UserDelete(username string) error
	UserUpdate(username string, password string) error
	NamespaceCreate(namespace, username, tenantID string) (*models.Namespace, error)
	NamespaceAddMember(username string, namespace string) (*models.Namespace, error)
	NamespaceRemoveMember(username string, namespace string) (*models.Namespace, error)
	NamespaceDelete(namespace string) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) UserCreate(username, password, email string) (*models.User, error) {
	// When UserCreate, a store function, return ErrDuplicated, checks which already exists in the database.
	findConflict := func() (*models.User, error) {
		userList, _, err := s.store.UserList(context.Background(), paginator.Query{Page: -1, PerPage: -1}, nil)
		if err != nil {
			return nil, err
		}

		var errs [2]error
		for _, userItem := range userList {
			if userItem.Username == username {
				errs[0] = ErrUserNameExists
			}

			if userItem.Email == email {
				errs[1] = ErrUserEmailExists
			}
		}
		if errs == [2]error{ErrUserNameExists, ErrUserEmailExists} {
			return nil, ErrUserNameAndEmailExists
		}

		for _, err := range errs {
			if err != nil {
				return nil, err
			}
		}

		return nil, err
	}

	username = normalizeString(username)
	email = normalizeString(email)

	userData := models.UserData{
		Name:     username,
		Email:    email,
		Username: username,
	}
	_, err := validator.ValidateStruct(userData)
	if err != nil {
		return nil, ErrUserDataInvalid
	}

	userPassword := models.UserPassword{
		Password: password,
	}
	_, err = validator.ValidateStruct(userPassword)
	if err != nil {
		return nil, ErrUserPasswordInvalid
	}

	userPassword.Password = hashPassword(userPassword.Password)

	user := &models.User{
		UserData:     userData,
		UserPassword: userPassword,
		Confirmed:    true,
		CreatedAt:    clock.Now(),
	}

	err = s.store.UserCreate(context.Background(), user)
	if err != nil {
		if err == store.ErrDuplicate {
			return findConflict()
		}

		return nil, ErrCreateNewUser
	}

	return user, nil
}

func (s *service) UserDelete(username string) error {
	user, err := s.store.UserGetByUsername(context.Background(), username)
	if err != nil {
		return ErrUserNotFound
	}

	namespaces, err := s.store.UserDetachInfo(context.Background(), user.ID)
	if err != nil {
		return ErrNamespaceNotFound
	}

	// It is all namespaces what the user is owner.
	ownedNamespaces := namespaces["owner"]
	// It is all namespaces what the user is member.
	memberNamespaces := namespaces["member"]

	// Delete all namespaces that the user is owner.
	for _, ownedNamespace := range ownedNamespaces {
		if err := s.store.NamespaceDelete(context.Background(), ownedNamespace.TenantID); err != nil {
			return err
		}
	}

	// Remove user from all namespaces it is a member.
	for _, memberNamespace := range memberNamespaces {
		if _, err := s.store.NamespaceRemoveMember(context.Background(), memberNamespace.TenantID, user.ID); err != nil {
			return err
		}
	}

	if err := s.store.UserDelete(context.Background(), user.ID); err != nil {
		return ErrFailedDeleteUser
	}

	return nil
}

func (s *service) UserUpdate(username string, password string) error {
	passwordData := models.UserPassword{
		Password: password,
	}
	_, err := validator.ValidateStruct(passwordData)
	if err != nil {
		return ErrPasswordInvalid
	}

	passwordData.Password = hashPassword(password)

	user, err := s.store.UserGetByUsername(context.TODO(), username)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.store.UserUpdatePassword(context.TODO(), passwordData.Password, user.ID); err != nil {
		return ErrFailedUpdateUser
	}

	return nil
}

func (s *service) NamespaceCreate(namespace, username, tenantID string) (*models.Namespace, error) {
	// tenantID is optional.
	if tenantID == "" {
		tenantID = uuid.Generate()
	}

	user, err := s.store.UserGetByUsername(context.Background(), username)
	if err != nil {
		return nil, ErrUserNotFound
	}
	ns := &models.Namespace{
		Name:     namespace,
		Owner:    user.ID,
		TenantID: tenantID,
		Members:  []interface{}{user.ID},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}
	_, err = validator.ValidateStruct(ns)
	if err != nil {
		return nil, ErrNamespaceInvalid
	}

	ns, err = s.store.NamespaceCreate(context.Background(), ns)
	if err != nil {
		return nil, ErrDuplicateNamespace
	}

	return ns, nil
}

func (s *service) NamespaceAddMember(username string, namespace string) (*models.Namespace, error) {
	user, err := s.store.UserGetByUsername(context.Background(), username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns, err := s.store.NamespaceGetByName(context.Background(), namespace)
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	ns, err = s.store.NamespaceAddMember(context.Background(), ns.TenantID, user.ID)
	if err != nil {
		return nil, ErrFailedNamespaceAddMember
	}

	return ns, nil
}

func (s *service) NamespaceRemoveMember(username string, namespace string) (*models.Namespace, error) {
	usr, err := s.store.UserGetByUsername(context.TODO(), username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns, err := s.store.NamespaceGetByName(context.TODO(), namespace)
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	ns, err = s.store.NamespaceRemoveMember(context.TODO(), ns.TenantID, usr.ID)
	if err != nil {
		return nil, ErrFailedNamespaceRemoveMember
	}

	return ns, nil
}

func (s *service) NamespaceDelete(namespace string) error {
	ns, err := s.store.NamespaceGetByName(context.Background(), namespace)
	if err != nil {
		return ErrNamespaceNotFound
	}

	if err := s.store.NamespaceDelete(context.TODO(), ns.TenantID); err != nil {
		return ErrFailedDeleteNamespace
	}

	return nil
}
