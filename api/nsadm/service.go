package nsadm

import (
	"context"
	"errors"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"gopkg.in/go-playground/validator.v9"
)

var ErrUnauthorized = errors.New("unauthorized")

type Service interface {
	ListNamespaces(ctx context.Context, pagination paginator.Query) ([]models.Namespace, int, error)
	CreateNamespace(ctx context.Context, namespace *models.Namespace, ownerUsername string) (*models.Namespace, error)
	GetNamespace(ctx context.Context, namespace string) (*models.Namespace, error)
	DeleteNamespace(ctx context.Context, namespace, ownerUsername string) error
	EditNamespace(ctx context.Context, namespace, name, ownerUsername string) error
	AddNamespaceUser(ctx context.Context, namespace, username, ownerUsername string) error
	RemoveNamespaceUser(ctx context.Context, namespace, username, ownerUsername string) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) ListNamespaces(ctx context.Context, pagination paginator.Query) ([]models.Namespace, int, error) {
	return s.store.ListNamespaces(ctx, pagination)
}

func (s *service) CreateNamespace(ctx context.Context, namespace *models.Namespace, ownerUsername string) (*models.Namespace, error) {
	user, err := s.store.GetUserByUsername(ctx, ownerUsername)
	if err != nil {
		return nil, err
	}
	namespace.Owner = user.ID
	namespace.Members = []string{user.ID}
	if namespace.TenantID == "" {
		namespace.TenantID = uuid.Must(uuid.NewV4(), nil).String()
	}
	return s.store.CreateNamespace(ctx, namespace)
}

func (s *service) GetNamespace(ctx context.Context, namespace string) (*models.Namespace, error) {
	return s.store.GetNamespace(ctx, namespace)
}

func (s *service) DeleteNamespace(ctx context.Context, namespace, ownerUsername string) error {
	if ns, _ := s.store.GetNamespace(ctx, namespace); ns != nil {
		if user, _ := s.store.GetUserByUsername(ctx, ownerUsername); user != nil {
			if ns.Owner == user.ID {
				return s.store.DeleteNamespace(ctx, namespace)
			}
		}
	}
	return ErrUnauthorized
}

func (s *service) EditNamespace(ctx context.Context, namespace, name, ownerUsername string) error {
	if ns, _ := s.store.GetNamespace(ctx, namespace); ns != nil {
		if user, _ := s.store.GetUserByUsername(ctx, ownerUsername); user != nil {
			validate := validator.New()
			name = strings.ToLower(name)
			if ns.Name != name && ns.Owner == user.ID {
				ns.Name = name
				if err := validate.Struct(ns); err == nil {
					return s.store.EditNamespace(ctx, namespace, name)
				}
			}
		}
	}
	return ErrUnauthorized
}

func (s *service) AddNamespaceUser(ctx context.Context, namespace, username, ownerUsername string) error {
	if ns, _ := s.store.GetNamespace(ctx, namespace); ns != nil {
		if OwnerUser, _ := s.store.GetUserByUsername(ctx, ownerUsername); OwnerUser != nil {
			if user, _ := s.store.GetUserByUsername(ctx, username); user != nil {
				if ns.Owner == OwnerUser.ID {
					return s.store.AddNamespaceUser(ctx, namespace, user.ID)
				}
			}
		}
	}
	return ErrUnauthorized
}
func (s *service) RemoveNamespaceUser(ctx context.Context, namespace, username, ownerUsername string) error {
	if ns, _ := s.store.GetNamespace(ctx, namespace); ns != nil {
		if OwnerUser, _ := s.store.GetUserByUsername(ctx, ownerUsername); OwnerUser != nil {
			if user, _ := s.store.GetUserByUsername(ctx, username); user != nil {
				if ns.Owner == OwnerUser.ID {
					return s.store.RemoveNamespaceUser(ctx, namespace, user.ID)
				}
			}
		}
	}
	return ErrUnauthorized
}
