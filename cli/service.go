package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	uuid "github.com/satori/go.uuid"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Arguments struct {
	Username  string `validate:"required,min=3,max=30,alphanum,ascii"`
	Namespace string `validate:"required,min=3,max=30,alphanum,ascii"`
	Password  string `validate:"required,min=5,max=30"`
	Email     string `validate:"required,email"`
	TenantID  string
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

	usr, err := s.store.UserGetByUsername(context.TODO(), data.Username)
	if err == store.ErrDuplicateEmail {
		return "", store.ErrDuplicateEmail
	}

	if usr != nil {
		return "", ErrDuplicateUser
	}

	password := data.Password

	if err := s.store.UserCreate(context.TODO(), &models.User{
		Name:     data.Username,
		Username: data.Username,
		Password: hashPassword(password),
		Email:    data.Email,
	}); err != nil {
		return "", ErrCreateNewUser
	}

	return data.Username, nil
}

func (s *service) NamespaceCreate(data Arguments) (*models.Namespace, error) {
	if err := validateParameters(data); err != nil {
		return nil, ErrCreateNewNamespace
	}

	usr, err := s.store.UserGetByUsername(context.TODO(), data.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	ns, err := s.store.NamespaceGetByName(context.TODO(), data.Namespace)
	if err == store.ErrNamespaceNoDocuments {
		return nil, ErrNamespaceNotFound
	}

	if ns != nil {
		return nil, ErrDuplicateNamespace
	}

	var tenantID string

	if data.TenantID == "" {
		tenantID = uuid.Must(uuid.NewV4(), nil).String()
	} else {
		tenantID = data.TenantID
	}

	ns, err = s.store.NamespaceCreate(context.TODO(), &models.Namespace{
		Name:     data.Namespace,
		Owner:    usr.ID,
		TenantID: tenantID,
		Members:  []interface{}{usr.ID},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
	})
	if err != nil {
		return nil, ErrCreateNewNamespace
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

	ns, err = s.store.NamespaceAddMember(context.TODO(), ns.TenantID, usr.ID)
	if err != nil {
		return nil, ErrFailedAddNamespaceUser
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
