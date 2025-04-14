package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/internal/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/uptrace/bun"
)

func (pg *pg) NamespaceCreate(ctx context.Context, namespace *models.Namespace) (string, error) {
	if namespace.TenantID == "" {
		namespace.TenantID = uuid.Generate()
	}

	namespace.CreatedAt = clock.Now()
	namespace.UpdatedAt = clock.Now()

	if _, err := pg.driver.NewInsert().Model(entity.NamespaceFromModel(namespace)).Exec(ctx); err != nil {
		return "", fromSqlError(err)
	}

	return namespace.TenantID, nil
}

func (pg *pg) NamespaceCreateMemberships(ctx context.Context, tenantID string, memberships ...models.Member) error {
	entities := make([]entity.Membership, len(memberships))
	for i, m := range memberships {
		m.AddedAt = clock.Now()
		m.UpdatedAt = clock.Now()
		entities[i] = *entity.MembershipFromModel(&m, tenantID)
	}

	if _, err := pg.driver.NewInsert().Model(&entities).Exec(ctx); err != nil {
		return fromSqlError(err)
	}

	return nil
}

func (pg *pg) NamespaceConflicts(ctx context.Context, target *models.NamespaceConflicts) ([]string, bool, error) {
	namespaces := make([]map[string]any, 0)
	if err := pg.driver.NewSelect().Model((*entity.Namespace)(nil)).Column("name").Where("name = ?", target.Name).Scan(ctx, &namespaces); err != nil {
		return nil, false, fromSqlError(err)
	}

	conflicts := make([]string, 0)
	for _, user := range namespaces {
		if user["name"] == target.Name {
			conflicts = append(conflicts, "name")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (pg *pg) NamespaceList(ctx context.Context, opts ...store.QueryOption) ([]models.Namespace, int, error) {
	entities := make([]entity.Namespace, 0)
	query := pg.driver.NewSelect().Model(&entities).Relation("Memberships.User")
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSqlError(err)
	}

	count, err := query.ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fromSqlError(err)
	}

	namespaces := make([]models.Namespace, len(entities))
	for i, e := range entities {
		namespaces[i] = *entity.NamespaceToModel(&e)
	}

	return namespaces, count, nil
}

func (pg *pg) NamespaceGet(ctx context.Context, ident store.NamespaceIdent, val string, opts ...store.QueryOption) (*models.Namespace, error) {
	ns := new(entity.Namespace)

	query := pg.driver.NewSelect().Model(ns).Relation("Memberships.User").Where("? = ?", bun.Ident(ident), val)
	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, fromSqlError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.NamespaceToModel(ns), nil
}

func (pg *pg) NamespaceSave(ctx context.Context, namespace *models.Namespace) error {
	n := entity.NamespaceFromModel(namespace)
	n.UpdatedAt = clock.Now()

	_, err := pg.driver.NewUpdate().Model(n).WherePK().Exec(ctx)

	return fromSqlError(err)
}

func (pg *pg) NamespaceSaveMembership(ctx context.Context, tenantID string, member *models.Member) error {
	e := entity.MembershipFromModel(member, tenantID)
	e.UpdatedAt = clock.Now()
	_, err := pg.driver.NewUpdate().Model(e).WherePK().Exec(ctx)

	return fromSqlError(err)
}

func (pg *pg) NamespaceDelete(ctx context.Context, namespace *models.Namespace) error {
	n := entity.NamespaceFromModel(namespace)
	_, err := pg.driver.NewDelete().Model(n).WherePK().Exec(ctx)

	return fromSqlError(err)
}

func (pg *pg) NamespaceDeleteMembership(ctx context.Context, tenantID string, member *models.Member) error {
	e := entity.MembershipFromModel(member, tenantID)
	_, err := pg.driver.NewDelete().Model(e).WherePK().Exec(ctx)

	return fromSqlError(err)
}
