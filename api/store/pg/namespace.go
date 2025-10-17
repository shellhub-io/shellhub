package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) NamespaceList(ctx context.Context, opts ...store.QueryOption) ([]models.Namespace, int, error)

func (pg *Pg) NamespaceResolve(ctx context.Context, resolver store.NamespaceResolver, value string) (*models.Namespace, error)

func (pg *Pg) NamespaceGetPreferred(ctx context.Context, userID string) (*models.Namespace, error)

func (pg *Pg) NamespaceCreate(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error)

func (pg *Pg) NamespaceConflicts(ctx context.Context, target *models.NamespaceConflicts) (conflicts []string, has bool, err error)

func (pg *Pg) NamespaceUpdate(ctx context.Context, namespace *models.Namespace) error

func (pg *Pg) NamespaceIncrementDeviceCount(ctx context.Context, tenantID string, status models.DeviceStatus, count int64) error

func (pg *Pg) NamespaceDelete(ctx context.Context, namespace *models.Namespace) error

func (pg *Pg) NamespaceDeleteMany(ctx context.Context, tenantIDs []string) (int64, error)
