package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/uptrace/bun"
)

func (pg *Pg) OAuthClientCreate(ctx context.Context, client *models.OAuthClient) (string, error) {
	db := pg.GetConnection(ctx)

	now := clock.Now()
	client.CreatedAt = now
	client.UpdatedAt = now

	if client.ID == "" {
		client.ID = uuid.Generate()
	}

	if _, err := db.NewInsert().Model(entity.OAuthClientFromModel(client)).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return client.ID, nil
}

func (pg *Pg) OAuthClientResolve(ctx context.Context, resolver store.OAuthClientResolver, value string, opts ...store.QueryOption) (*models.OAuthClient, error) {
	db := pg.GetConnection(ctx)

	column, err := oauthClientResolverToColumn(resolver)
	if err != nil {
		return nil, err
	}

	e := new(entity.OAuthClient)
	q := db.NewSelect().Model(e).Where("? = ?", bun.Ident(column), value)

	q, err = applyOptions(ctx, q, opts...)
	if err != nil {
		return nil, err
	}

	if err := q.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.OAuthClientToModel(e), nil
}

func (pg *Pg) OAuthClientList(ctx context.Context, opts ...store.QueryOption) ([]models.OAuthClient, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.OAuthClient, 0)
	q := db.NewSelect().Model(&entities)

	var err error
	q, err = applyOptions(ctx, q, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := q.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	clients := make([]models.OAuthClient, len(entities))
	for i, e := range entities {
		clients[i] = *entity.OAuthClientToModel(&e)
	}

	return clients, count, nil
}

func (pg *Pg) OAuthClientDelete(ctx context.Context, client *models.OAuthClient) error {
	db := pg.GetConnection(ctx)

	e := entity.OAuthClientFromModel(client)
	r, err := db.NewDelete().Model(e).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func oauthClientResolverToColumn(resolver store.OAuthClientResolver) (string, error) {
	switch resolver {
	case store.OAuthClientIDResolver:
		return "id", nil
	case store.OAuthClientClientIDResolver:
		return "client_id", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
