package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) TagCreate(ctx context.Context, tag *models.Tag) (insertedID string, err error)

func (pg *Pg) TagConflicts(ctx context.Context, tenantID string, target *models.TagConflicts) (conflicts []string, has bool, err error)

func (pg *Pg) TagList(ctx context.Context, opts ...store.QueryOption) (tags []models.Tag, totalCount int, err error)

func (pg *Pg) TagResolve(ctx context.Context, resolver store.TagResolver, value string, opts ...store.QueryOption) (tag *models.Tag, err error)

func (pg *Pg) TagUpdate(ctx context.Context, tag *models.Tag) error

func (pg *Pg) TagPushToTarget(ctx context.Context, id string, target store.TagTarget, targetID string) (err error)

func (pg *Pg) TagPullFromTarget(ctx context.Context, id string, target store.TagTarget, targetIDs ...string) (err error)

func (pg *Pg) TagDelete(ctx context.Context, tag *models.Tag) error
