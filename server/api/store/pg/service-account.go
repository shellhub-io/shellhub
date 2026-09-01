package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
)

// ServiceAccountList implements [store.ServiceAccountStore].
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

	accounts := make([]models.ServiceAccount, 0, len(entities))
	for i := range entities {
		if entities[i].User == nil || entities[i].User.Type != string(models.UserTypeService) {
			continue
		}

		accounts = append(accounts, *entity.ServiceAccountFromMembership(&entities[i]))
	}

	return accounts, len(accounts), nil
}
