package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	uuid "github.com/satori/go.uuid"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Parameters struct {
	Username  string `validate:"required,min=3,max=30,alphanum,ascii"`
	Namespace string `validate:"required,min=3,max=30,alphanum,ascii"`
	Password  string `validate:"required,min=5,max=30"`
	Email     string `validate:"required,email"`
	TenantID  string
}

type Service interface {
	UserCreate(Parameters) (string, error)
	NamespaceCreate(Parameters) (*models.Namespace, error)
	NamespaceAddMember(Parameters) (*models.Namespace, error)
	NamespaceDelete(Parameters) error
	UserDelete(Parameters) error
	NamespaceRemoveMember(Parameters) (*models.Namespace, error)
	UserUpdate(Parameters) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) UserCreate(data Parameters) (string, error) {
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

	if err := s.store.UserCreate(context.TODO(), &models.User{
		Name:     data.Username,
		Username: data.Username,
		Password: hashPassword(data.Password),
		Email:    data.Email,
	}); err != nil {
		return "", ErrCreateNewUser
	}

	return data.Username, nil
}

func (s *service) NamespaceCreate(data Parameters) (*models.Namespace, error) {
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

	if data.TenantID == "" {
		data.TenantID = uuid.Must(uuid.NewV4(), nil).String()
	}

	ns, err = s.store.NamespaceCreate(context.TODO(), &models.Namespace{
		Name:     data.Namespace,
		Owner:    usr.ID,
		TenantID: data.TenantID,
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

func (s *service) NamespaceAddMember(data Parameters) (*models.Namespace, error) {
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

func (s *service) NamespaceDelete(data Parameters) error {
	ns, err := s.store.NamespaceGetByName(context.TODO(), data.Namespace)
	if err != nil {
		return ErrNamespaceNotFound
	}

	if err := s.store.NamespaceDelete(context.TODO(), ns.TenantID); err != nil {
		return ErrFailedDeleteNamespace
	}

	return nil
}

func (s *service) UserDelete(data Parameters) error {
	usr, err := s.store.UserGetByUsername(context.TODO(), data.Username)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.store.UserDelete(context.TODO(), usr.ID); err != nil {
		return ErrFailedDeleteUser
	}

	return nil
}

func (s *service) NamespaceRemoveMember(data Parameters) (*models.Namespace, error) {
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

func (s *service) UserUpdate(data Parameters) error {
	if err := validateParameters(data); err != nil {
		return ErrChangePassword
	}

	usr, err := s.store.UserGetByUsername(context.TODO(), data.Username)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.store.UserUpdate(context.TODO(), usr.Name, usr.Username, usr.Email, usr.Password, hashPassword(data.Password), usr.ID); err != nil {
		return ErrFailedUpdateUser
	}

	return nil
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))

	return hex.EncodeToString(hash[:])
}
