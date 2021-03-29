package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type NamespaceStore interface {
	ListNamespaces(ctx context.Context, pagination paginator.Query, filters []models.Filter, export bool) ([]models.Namespace, int, error)
	GetNamespace(ctx context.Context, namespace string) (*models.Namespace, error)
	GetNamespaceByName(ctx context.Context, namespace string) (*models.Namespace, error)
	CreateNamespace(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error)
	EditNamespace(ctx context.Context, namespace, name string) (*models.Namespace, error)
	UpdateNamespace(ctx context.Context, tenant string, namespace *models.Namespace) error
	DeleteNamespace(ctx context.Context, namespace string) error
	AddNamespaceUser(ctx context.Context, namespace, ID string) (*models.Namespace, error)
	RemoveNamespaceUser(ctx context.Context, namespace, ID string) (*models.Namespace, error)
	GetSomeNamespace(ctx context.Context, ID string) (*models.Namespace, error)
	UpdateDataUserSecurity(ctx context.Context, sessionRecord bool, tenant string) error
	GetDataUserSecurity(ctx context.Context, tenant string) (bool, error)
}
