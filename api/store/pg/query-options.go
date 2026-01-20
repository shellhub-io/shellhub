package pg

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/internal"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// ErrQueryNotFound is returned when the query context value is not found or has the wrong type
var ErrQueryNotFound = errors.New("query not found in context")

func (pg *Pg) Options() store.QueryOptions {
	return pg.options
}

func (*queryOptions) Paginate(page *query.Paginator) store.QueryOption {
	return func(ctx context.Context) error {
		query, ok := ctx.Value("query").(*bun.SelectQuery)
		if !ok {
			return ErrQueryNotFound
		}

		query = query.Offset(page.PerPage * (page.Page - 1)).Limit(page.PerPage) //nolint:staticcheck

		return nil
	}
}

func (*queryOptions) Sort(sorter *query.Sorter) store.QueryOption {
	return func(ctx context.Context) error {
		if sorter.By == "" {
			return nil
		}

		query, ok := ctx.Value("query").(*bun.SelectQuery)
		if !ok {
			return ErrQueryNotFound
		}

		query = query.OrderExpr("? ?", bun.Ident(sorter.By), bun.Safe(strings.ToUpper(sorter.Order))) //nolint:staticcheck

		return nil
	}
}

func (*queryOptions) Match(filters *query.Filters) store.QueryOption {
	return func(ctx context.Context) error {
		if len(filters.Data) < 1 {
			return nil
		}

		bunQuery, ok := ctx.Value("query").(*bun.SelectQuery)
		if !ok {
			return ErrQueryNotFound
		}

		var filterErr error
		bunQuery = bunQuery.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery { //nolint:staticcheck
			currentOperator := "OR" //nolint:staticcheck
			firstCondition := true

			for _, filter := range filters.Data {
				switch filter.Type {
				case query.FilterTypeOperator:
					param, ok := filter.Params.(*query.FilterOperator)
					if !ok {
						return nil
					}

					op, valid := internal.ParseFilterOperator(param)
					if !valid {
						continue
					}

					currentOperator = op
				case query.FilterTypeProperty:
					param, ok := filter.Params.(*query.FilterProperty)
					if !ok {
						return nil
					}

					condition, args, valid, err := internal.ParseFilterProperty(param)
					if err != nil || !valid {
						filterErr = err

						continue
					}

					switch {
					case firstCondition: // The first condition always applies a WHERE
						q = q.Where(condition, args...)
						firstCondition = false
					case currentOperator == "AND":
						q = q.Where(condition, args...)
					case currentOperator == "OR":
						q = q.WhereOr(condition, args...)
					}
				default:
					return nil
				}
			}

			return q
		})

		if filterErr != nil {
			return filterErr
		}

		return nil
	}
}

func (*queryOptions) WithMember(userID string) store.QueryOption {
	return func(ctx context.Context) error {
		query, ok := ctx.Value("query").(*bun.SelectQuery)
		if !ok {
			return ErrQueryNotFound
		}

		query = query.Where("EXISTS (SELECT 1 FROM memberships WHERE memberships.namespace_id = namespace.id AND memberships.user_id = ?)", userID) //nolint:staticcheck

		return nil
	}
}

func (*queryOptions) InNamespace(namespaceID string) store.QueryOption {
	return func(ctx context.Context) error {
		query, ok := ctx.Value("query").(*bun.SelectQuery)
		if !ok {
			return ErrQueryNotFound
		}

		query = query.Where("namespace_id = ?", namespaceID) //nolint:staticcheck

		return nil
	}
}

func (*queryOptions) WithDeviceStatus(status models.DeviceStatus) store.QueryOption {
	return func(ctx context.Context) error {
		query, ok := ctx.Value("query").(*bun.SelectQuery)
		if !ok {
			return ErrQueryNotFound
		}

		query = query.Where("status = ?", string(status)) //nolint:staticcheck

		return nil
	}
}
