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

func (*queryOptions) WithMember(userID string) store.QueryOption {
	return func(ctx context.Context) error {
		pipeline, ok := ctx.Value("query").(*[]bson.M)
		if !ok {
			return errors.New("query not found in context")
		}

		stage := bson.M{
			"$match": bson.M{
				"members": bson.M{
					"$elemMatch": bson.M{"id": userID},
				},
			},
		}

		*pipeline = append([]bson.M{stage}, *pipeline...)

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

		conditions := make([]bson.M, 0)
		currentOperator := "$or"

		for _, data := range filters.Data {
			switch data.Type {
			case query.FilterTypeProperty:
				param, ok := data.Params.(*query.FilterProperty)
				if !ok {
					return query.ErrFilterInvalid
				}

				if param.Name == "custom_fields" {
					condition, ok, err := internal.ParseCustomFieldsFilter(param)
					switch {
					case err != nil:
						return query.ErrFilterPropertyInvalid
					case ok:
						conditions = append(conditions, condition)
					}

					continue
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

				if operator != currentOperator && len(conditions) > 0 {
					*pipeline = append(*pipeline, bson.M{"$match": bson.M{currentOperator: conditions}})
					conditions = make([]bson.M, 0)
				}

				currentOperator = operator
			default:
				return query.ErrFilterInvalid
			}
		}

		if len(conditions) > 0 {
			*pipeline = append(*pipeline, bson.M{"$match": bson.M{currentOperator: conditions}})
		}

		return nil
	}
}
