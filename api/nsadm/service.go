package nsadm

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"gopkg.in/go-playground/validator.v9"
)

var ErrUnauthorized = errors.New("unauthorized")

type Service interface {
	ListNamespaces(ctx context.Context, pagination paginator.Query) ([]models.Namespace, int, error)
	CreateNamespace(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error)
	GetNamespace(ctx context.Context, namespace string) (*models.Namespace, error)
	DeleteNamespace(ctx context.Context, namespace string) error
	EditNamespace(ctx context.Context, namespace, name string) error
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

func (s *service) CreateNamespace(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error) {
	return s.store.CreateNamespace(ctx, namespace)
}

func (s *service) GetNamespace(ctx context.Context, namespace string) (*models.Namespace, error) {
	return s.store.GetNamespace(ctx, namespace)
}

func (s *service) DeleteNamespace(ctx context.Context, namespace string) error {
	ns, _ := s.store.GetNamespace(ctx, namespace)
	if ns != nil {
		return s.store.DeleteNamespace(ctx, namespace)
	}
	return ErrUnauthorized
}

func (s *service) EditNamespace(ctx context.Context, namespace, name string) error {
	ns, _ := s.store.GetNamespace(ctx, namespace)
	validate := validator.New()
	name = strings.ToLower(name)
	if ns != nil {
		if ns.Name != name {
			ns.Name = name
			if err := validate.Struct(ns); err == nil {
				return s.store.EditNamespace(ctx, namespace, name)
			}
		}
	}

	return ErrUnauthorized
}
