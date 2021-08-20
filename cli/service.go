package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

type Arguments struct {
	Username   string `validate:"required,min=3,max=30,alphanum,ascii"`
	Namespace  string `validate:"required,min=3,max=30,alphanum,ascii"`
	Password   string `validate:"required,min=5,max=30"`
	Email      string `validate:"required,email"`
	TenantID   string
	AccessType string `validate:"required,oneof=admin operator observer`
}

type Service interface {
	UserCreate(Arguments) (string, error)
	NamespaceCreate(Arguments) (*models.Namespace, error)
	NamespaceAddMember(Arguments) (*models.Namespace, error)
	NamespaceDelete(Arguments) error
	UserDelete(Arguments) error
	NamespaceRemoveMember(Arguments) (*models.Namespace, error)
	UserUpdate(Arguments) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) UserCreate(data Arguments) (string, error) {
	if err := validateParameters(data); err != nil {
		return "", ErrCreateNewUser
	}

	password := data.Password

	if err := s.store.UserCreate(context.TODO(), &models.User{
		Name:          strings.ToLower(data.Username),
		Username:      data.Username,
		Password:      hashPassword(password),
		Email:         strings.ToLower(data.Email),
		Authenticated: true,
		CreatedAt:     clock.Now(),
	}); err != nil && err.Error() == "duplicate" {
		var errStrings []string

		usrList, _, _ := s.store.UserList(context.TODO(), paginator.Query{Page: -1, PerPage: -1}, nil)

		for _, usr := range usrList {
			if usr.Username == data.Username {
				errStrings = append(errStrings, "user already exists")
			}

			if usr.Email == data.Email {
				errStrings = append(errStrings, "email address is already in use")
			}
		}

		for _, err := range errStrings {
			fmt.Println(err) //nolint:forbidigo
		}

		return "", ErrCreateNewUser
	}

	return strings.ToLower(data.Username), nil
}

func (s *service) NamespaceCreate(data Arguments) (*models.Namespace, error) {
	if err := validateParameters(data); err != nil {
		return nil, ErrCreateNewNamespace
	}

	usr, err := s.store.UserGetByUsername(context.TODO(), data.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	var tenantID string

	if data.TenantID == "" {
		tenantID = uuid.Generate()
	} else {
		tenantID = data.TenantID
	}

	ns, err := s.store.NamespaceCreate(context.TODO(), &models.Namespace{
		Name:     data.Namespace,
		Owner:    usr.ID,
		TenantID: tenantID,
		Members:  []interface{}{&models.Member{ID: user.ID, AccessType: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	})
	if err != nil {
		return nil, ErrDuplicateNamespace
	}

	return ns, nil
}

func (s *service) NamespaceAddMember(data Arguments) (*models.Namespace, error) {
	usr, err := s.store.UserGetByUsername(context.TODO(), data.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns, err := s.store.NamespaceGetByName(context.TODO(), data.Namespace)
	if err != nil {
		return nil, ErrNamespaceNotFound
	}
member:
	&models.Member{ID: usr.ID, AccessType: data.AccessType}
	ns, err = s.store.NamespaceAddMember(context.TODO(), ns.TenantID, member)
	if err != nil {
		return nil, err
	}

	return ns, nil
}

func (s *service) NamespaceDelete(data Arguments) error {
	ns, err := s.store.NamespaceGetByName(context.TODO(), data.Namespace)
	if err != nil {
		return ErrNamespaceNotFound
	}

	if err := s.store.NamespaceDelete(context.TODO(), ns.TenantID); err != nil {
		return ErrFailedDeleteNamespace
	}

	return nil
}

func (s *service) UserDelete(data Arguments) error {
	usr, err := s.store.UserGetByUsername(context.TODO(), data.Username)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.store.UserDelete(context.TODO(), usr.ID); err != nil {
		return ErrFailedDeleteUser
	}

	return nil
}

func (s *service) NamespaceRemoveMember(data Arguments) (*models.Namespace, error) {
	usr, err := s.store.UserGetByUsername(context.TODO(), data.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns, err := s.store.NamespaceGetByName(context.TODO(), data.Namespace)
	if err != nil {
		return nil, ErrNamespaceNotFound
	}

	ns, err = s.store.NamespaceRemoveMember(context.TODO(), ns.TenantID, usr.ID)
	if err != nil {
		return nil, ErrFailedRemoveMember
	}

	return ns, nil
}

func (s *service) UserUpdate(data Arguments) error {
	if err := validateParameters(data); err != nil {
		return ErrChangePassword
	}

	usr, err := s.store.UserGetByUsername(context.TODO(), data.Username)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.store.UserUpdatePassword(context.TODO(), hashPassword(data.Password), usr.ID); err != nil {
		return ErrFailedUpdateUser
	}

	return nil
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))

	return hex.EncodeToString(hash[:])
}
