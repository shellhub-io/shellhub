package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg/internal/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *pg) PrivateKeyCreate(ctx context.Context, key *models.PrivateKey) error {
	key.CreatedAt = clock.Now()
	key.UpdatedAt = clock.Now()

	if _, err := pg.driver.NewInsert().Model(&entity.PrivateKey{PrivateKey: *key}).Exec(ctx); err != nil {
		return fromSqlError(err)
	}

	return nil
}

func (pg *pg) PrivateKeyGet(ctx context.Context, fingerprint string) (*models.PrivateKey, error) {
	// TODO: private keys are now saved only in the frontend and this can be removedV
	return nil, nil
}
