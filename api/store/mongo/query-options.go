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

// WithMember is a no-op in Mongo because filtering by member is done implicitly
// via gateway.IDFromContext in NamespaceList. This method exists for interface
// compatibility with the PostgreSQL implementation.
func (*queryOptions) WithMember(_ string) store.QueryOption {
	return func(_ context.Context) error {
		return nil
	}
}

func (*queryOptions) Sort(sorter *query.Sorter) store.QueryOption {
	return func(ctx context.Context) error {
		if sorter == nil || sorter.By == "" {
			return nil
		}

		pipeline, ok := ctx.Value("query").(*[]bson.M)
		if !ok {
			return errors.New("query not found in context")
		}

		options := map[string]int{query.OrderAsc: 1, query.OrderDesc: -1}
		order, ok := options[sorter.Order]
		if !ok {
			order = -1
		}

		*pipeline = append(*pipeline, bson.M{"$sort": bson.M{sorter.By: order}})

		return nil
	}
}

func (*queryOptions) Paginate(paginator *query.Paginator) store.QueryOption {
	return func(ctx context.Context) error {
		if paginator == nil || paginator.Page < 1 || paginator.PerPage < 1 {
			return nil
		}

		pipeline, ok := ctx.Value("query").(*[]bson.M)
		if !ok {
			return errors.New("query not found in context")
		}

		*pipeline = append(*pipeline, []bson.M{{"$skip": paginator.PerPage * (paginator.Page - 1)}, {"$limit": paginator.PerPage}}...)

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
