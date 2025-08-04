package mongo

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) Options() store.QueryOptions {
	return s.options
}

func (*queryOptions) InNamespace(tenantID string) store.QueryOption {
	return func(ctx context.Context) error {
		query, ok := ctx.Value("query").(*[]bson.M)
		if !ok {
			return errors.New("query not found in context")
		}

		*query = append(*query, bson.M{
			"$match": bson.M{
				"tenant_id": tenantID,
			},
		})

		return nil
	}
}

func (*queryOptions) WithDeviceStatus(status models.DeviceStatus) store.QueryOption {
	return func(ctx context.Context) error {
		query, ok := ctx.Value("query").(*[]bson.M)
		if !ok {
			return errors.New("query not found in context")
		}

		*query = append(*query, bson.M{
			"$match": bson.M{
				"status": status,
			},
		})

		return nil
	}
}
