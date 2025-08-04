package mongo

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo/internal"
	"github.com/shellhub-io/shellhub/pkg/api/query"
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

func (*queryOptions) Match(filters *query.Filters) store.QueryOption {
	return func(ctx context.Context) error {
		if len(filters.Data) < 1 {
			return nil
		}

		pipeline, ok := ctx.Value("query").(*[]bson.M)
		if !ok {
			return errors.New("query not found in context")
		}

		conditions, stages := make([]bson.M, 0), make([]bson.M, 0)
		for _, data := range filters.Data {
			switch data.Type {
			case query.FilterTypeProperty:
				param, ok := data.Params.(*query.FilterProperty)
				if !ok {
					return query.ErrFilterInvalid
				}

				property, ok, err := internal.ParseFilterProperty(param)
				switch {
				case err != nil:
					return query.ErrFilterPropertyInvalid
				case ok:
					conditions = append(conditions, bson.M{param.Name: property})
				}
			case query.FilterTypeOperator:
				param, ok := data.Params.(*query.FilterOperator)
				if !ok {
					return query.ErrFilterInvalid
				}

				operator, ok := internal.ParseFilterOperator(param)
				if !ok {
					continue
				}

				stages = append(stages, bson.M{"$match": bson.M{operator: conditions}})
				conditions = nil
			default:
				return query.ErrFilterInvalid
			}
		}

		if len(conditions) > 0 {
			stages = append(stages, bson.M{"$match": bson.M{"$or": conditions}})
		}

		*pipeline = append(*pipeline, stages...)

		return nil
	}
}
