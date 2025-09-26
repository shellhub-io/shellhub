package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) NamespaceCreateMemberships(ctx context.Context, tenantID string, memberships ...models.Member) error {
	entities := make([]entity.Membership, len(memberships))
	for i, m := range memberships {
		m.AddedAt = clock.Now()
		entities[i] = *entity.MembershipFromModel(tenantID, &m)
	}

	if _, err := pg.driver.NewInsert().Model(&entities).Exec(ctx); err != nil {
		return fromSqlError(err)
	}

	return nil
}

func (pg *Pg) NamespaceSaveMembership(ctx context.Context, tenantID string, member *models.Member) error {
	e := entity.MembershipFromModel(tenantID, member)
	e.UpdatedAt = clock.Now()
	_, err := pg.driver.NewUpdate().Model(e).WherePK().Exec(ctx)

	return fromSqlError(err)
}

func (pg *Pg) NamespaceDeleteMembership(ctx context.Context, tenantID string, member *models.Member) error {
	e := entity.MembershipFromModel(tenantID, member)
	_, err := pg.driver.NewDelete().Model(e).WherePK().Exec(ctx)

	return fromSqlError(err)
}
