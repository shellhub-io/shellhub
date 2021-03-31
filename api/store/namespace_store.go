package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type NamespaceStore interface {
	NamespaceList(ctx context.Context, pagination paginator.Query, filters []models.Filter, export bool) ([]models.Namespace, int, error)
	NamespaceGet(ctx context.Context, namespace string) (*models.Namespace, error)
	NamespaceGetByName(ctx context.Context, namespace string) (*models.Namespace, error)
	NamespaceCreate(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error)
	NamespaceRename(ctx context.Context, namespace, name string) (*models.Namespace, error)
	NamespaceUpdate(ctx context.Context, tenant string, namespace *models.Namespace) error
	NamespaceDelete(ctx context.Context, namespace string) error
	NamespaceAddMember(ctx context.Context, namespace, ID string) (*models.Namespace, error)
	NamespaceRemoveMember(ctx context.Context, namespace, ID string) (*models.Namespace, error)
	NamespaceGetFirst(ctx context.Context, ID string) (*models.Namespace, error)
	NamespaceSetSessionRecord(ctx context.Context, sessionRecord bool, tenant string) error
	NamespaceGetSessionRecord(ctx context.Context, tenant string) (bool, error)
}
