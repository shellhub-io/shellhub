package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) ServiceAccountList(ctx context.Context, tenantID string) ([]models.ServiceAccount, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.Membership, 0)
	if err := db.NewSelect().
		Model(&entities).
		Relation("User").
		Where("membership.namespace_id = ?", tenantID).
		OrderExpr("membership.created_at ASC").
		Scan(ctx); err != nil {
		return nil, 0, fromSQLError(err)
	}

	// A namespace holds few members, so filtering the service-typed ones here avoids
	// quoting the "user" relation alias in SQL (user is a reserved word).
	accounts := make([]models.ServiceAccount, 0, len(entities))
	for i := range entities {
		if entities[i].User == nil || entities[i].User.Type != string(models.UserTypeService) {
			continue
		}

		accounts = append(accounts, *entity.ServiceAccountFromMembership(&entities[i]))
	}

	return accounts, len(accounts), nil
}
