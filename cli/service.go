package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"gopkg.in/go-playground/validator.v9"
)

var (
	ErrCreateNewUser          = errors.New("failed to create a new user")
	ErrCreateNewNamespace     = errors.New("failed to create a new namespace")
	ErrDuplicateUser          = errors.New("user already exists")
	ErrInvalidFormatNamespace = errors.New("invalid format for namespace")
	ErrDuplicateNamespace     = errors.New("namespace already exists")
	ErrInvalidFormatPassword  = errors.New("invalid format for password")
	ErrUserNotFound           = errors.New("user not found")
	ErrNamespaceNotFound      = errors.New("namespace not found")
	ErrFailedAddNamespaceUser = errors.New("failed to add the namespace for the user")
	ErrFailedDeleteUser       = errors.New("failed to delete the user")
	ErrFailedDeleteNamespace  = errors.New("failed to delete the namespace")
	ErrFailedUpdateUser       = errors.New("failed to reset the password for the user")
	ErrFailedRemoveMember     = errors.New("failed to remove member from the namespace")
)

type Parameters struct {
	Username  string
	Namespace string
	Password  string
	Email     string
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
	var errstrings []string
	validator := validator.New()

	if err := validator.Var(data.Username, "required,min=3,max=30,alphanum,ascii"); err != nil {
		errstrings = append(errstrings, fmt.Errorf("invalid format for username").Error())
	}

	if err := validator.Var(data.Email, "required,email"); err != nil {
		errstrings = append(errstrings, fmt.Errorf("email is not in a valid format").Error())
	}

	if err := validator.Var(data.Password, "required,min=5,max=30"); err != nil {
		errstrings = append(errstrings, fmt.Errorf("invalid format for password").Error())
	}

	if len(errstrings) > 0 {
		fmt.Println(fmt.Errorf(strings.Join(errstrings, "\n")))
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
	if err := validator.New().Var(data.Namespace, "required,min=3,max=30,alphanum,ascii"); err != nil {
		return nil, ErrInvalidFormatNamespace
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

	ns, err = s.store.NamespaceCreate(context.TODO(), &models.Namespace{
		Name:     data.Namespace,
		Owner:    data.Username,
		TenantID: uuid.Must(uuid.NewV4(), nil).String(),
		Members: []struct {
			ID   string
			Name string
		}{
			{usr.ID, data.Namespace},
		},
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
	if err := validator.New().Var(data.Password, "required,min=5,max=30"); err != nil {
		return ErrInvalidFormatPassword
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
