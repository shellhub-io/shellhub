package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

func (pg *Pg) ConnectionCreate(ctx context.Context, connection *models.Connection) (string, error) {
	db := pg.GetConnection(ctx)

	connection.CreatedAt = clock.Now()
	connection.UpdatedAt = clock.Now()
	if _, err := db.NewInsert().Model(entity.ConnectionFromModel(connection)).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return connection.ID, nil
}

func (pg *Pg) ConnectionList(ctx context.Context, opts ...store.QueryOption) ([]models.Connection, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.Connection, 0)

	query := db.NewSelect().Model(&entities)
	var err error
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	connections := make([]models.Connection, len(entities))
	for i, e := range entities {
		connections[i] = *entity.ConnectionToModel(&e)
	}

	return connections, count, nil
}

func (pg *Pg) ConnectionResolve(ctx context.Context, resolver store.ConnectionResolver, val string, opts ...store.QueryOption) (*models.Connection, error) {
	db := pg.GetConnection(ctx)

	column, err := connectionResolverToString(resolver)
	if err != nil {
		return nil, err
	}

	c := new(entity.Connection)
	query := db.NewSelect().Model(c).Where("? = ?", bun.Ident(column), val)
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	if err = query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.ConnectionToModel(c), nil
}

func (pg *Pg) ConnectionUpdate(ctx context.Context, connection *models.Connection) error {
	db := pg.GetConnection(ctx)

	connection.UpdatedAt = clock.Now()
	e := entity.ConnectionFromModel(connection)
	r, err := db.NewUpdate().
		Model(e).
		Column("label", "kind", "host", "port", "device_uid", "username", "auth_method", "key_fingerprint", "updated_at").
		WherePK().
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) ConnectionDelete(ctx context.Context, connection *models.Connection) error {
	db := pg.GetConnection(ctx)

	c := entity.ConnectionFromModel(connection)
	r, err := db.NewDelete().Model(c).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected, err := r.RowsAffected(); err != nil || rowsAffected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

func connectionResolverToString(resolver store.ConnectionResolver) (string, error) {
	switch resolver {
	case store.ConnectionIDResolver:
		return "id", nil
	default:
		return "", store.ErrResolverNotFound
	}
}
