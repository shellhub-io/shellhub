package pg

import (
	"context"
	"errors"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/internal"
	"github.com/uptrace/bun"
)

// ErrQueryNotFound is returned when the query context value is not found or has the wrong type
var ErrQueryNotFound = errors.New("query not found in context")

// ErrFilterNodeInvalid is returned by Match for a filter node of an unknown type, or whose Params
// is not the shape its type names. Decoding and [query.ValidateFilters] refuse both, so only a filter
// built by hand reaches it.
var ErrFilterNodeInvalid = errors.New("filter node is not valid")

// ErrLockWithoutTableAlias is returned by ForUpdate on a query that set no CtxTableAlias, since
// only an aliased query can name the one table to lock.
var ErrLockWithoutTableAlias = errors.New("for update requires a table alias")

// Options implements [store.Store].
func (pg *Pg) Options() store.QueryOptions {
	return pg.options
}

func (*queryOptions) Paginate(page *query.Paginator) store.QueryOption {
	return func(ctx context.Context) error {
		if page == nil || page.Page < 1 || page.PerPage < 1 {
			return nil
		}

		wrapper, ok := ctx.Value("query").(*queryWrapper)
		if !ok {
			return ErrQueryNotFound
		}

		wrapper.query = wrapper.query.Offset(page.PerPage * (page.Page - 1)).Limit(page.PerPage)

		return nil
	}
}

func sortDirection(order string) bun.Safe {
	if order == query.OrderAsc {
		return bun.Safe("ASC")
	}

	return bun.Safe("DESC")
}

func (*queryOptions) Sort(sorter *query.Sorter) store.QueryOption {
	return func(ctx context.Context) error {
		if sorter.By == "" {
			return nil
		}

		wrapper, ok := ctx.Value("query").(*queryWrapper)
		if !ok {
			return ErrQueryNotFound
		}

		qualifyCol := func(col string) bun.Ident {
			if alias, ok := ctx.Value(CtxTableAlias).(string); ok && alias != "" {
				return bun.Ident(alias + "." + col)
			}

			return bun.Ident(col)
		}

		direction := sortDirection(sorter.Order)

		if sorter.Tiebreak != "" {
			wrapper.query = wrapper.query.OrderExpr("? ?, ? ?", qualifyCol(sorter.By), direction, qualifyCol(sorter.Tiebreak), direction)
		} else {
			wrapper.query = wrapper.query.OrderExpr("? ?", qualifyCol(sorter.By), direction)
		}

		return nil
	}
}

func (o *queryOptions) Match(filters *query.Filters) store.QueryOption {
	return func(ctx context.Context) error {
		if len(filters.Data) < 1 {
			return nil
		}

		wrapper, ok := ctx.Value("query").(*queryWrapper)
		if !ok {
			return ErrQueryNotFound
		}

		var tableAlias string
		if alias, ok := ctx.Value(CtxTableAlias).(string); ok {
			tableAlias = alias
		}

		match := filterMatch{tableAlias: tableAlias, onlineCutoff: OnlineCutoff(o.clock)}

		var err error
		wrapper.query = wrapper.query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			q, _, err = match.where(q, filters.Data)

			return q
		})

		return err
	}
}

type filterMatch struct {
	tableAlias   string
	onlineCutoff time.Time
}

func (m filterMatch) where(q *bun.SelectQuery, nodes []query.Filter) (*bun.SelectQuery, bool, error) {
	operator := "OR"
	added := false

	for _, node := range nodes {
		switch node.Type {
		case query.FilterTypeOperator:
			params, ok := node.Params.(*query.FilterOperator)
			if !ok {
				return q, added, ErrFilterNodeInvalid
			}

			if op, valid := internal.ParseFilterOperator(params); valid {
				operator = op
			}
		case query.FilterTypeProperty:
			params, ok := node.Params.(*query.FilterProperty)
			if !ok {
				return q, added, ErrFilterNodeInvalid
			}

			condition, args, valid, err := internal.ParseFilterProperty(params, m.tableAlias, m.onlineCutoff)
			if err != nil {
				return q, added, err
			}

			if !valid {
				continue
			}

			if joinsWithOr(operator, added) {
				q = q.WhereOr(condition, args...)
			} else {
				q = q.Where(condition, args...)
			}

			added = true
		case query.FilterTypeGroup:
			params, ok := node.Params.(query.FilterGroup)
			if !ok {
				return q, added, ErrFilterNodeInvalid
			}

			var groupAdded bool
			var err error
			q, groupAdded, err = m.whereGroup(q, params, joinsWithOr(operator, added))
			if err != nil {
				return q, added, err
			}

			added = added || groupAdded
		default:
			return q, added, ErrFilterNodeInvalid
		}
	}

	return q, added, nil
}

func (m filterMatch) whereGroup(q *bun.SelectQuery, group query.FilterGroup, orWithPrevious bool) (*bun.SelectQuery, bool, error) {
	separator := " AND "
	if orWithPrevious {
		separator = " OR "
	}

	var added bool
	var err error
	q = q.WhereGroup(separator, func(q *bun.SelectQuery) *bun.SelectQuery {
		q, added, err = m.where(q, group)

		return q
	})

	return q, added, err
}

func joinsWithOr(operator string, added bool) bool {
	return operator == "OR" && added
}

func (*queryOptions) WithMember(userID string) store.QueryOption {
	return func(ctx context.Context) error {
		wrapper, ok := ctx.Value("query").(*queryWrapper)
		if !ok {
			return ErrQueryNotFound
		}

		wrapper.query = wrapper.query.Where("EXISTS (SELECT 1 FROM memberships WHERE memberships.namespace_id = namespace.id AND memberships.user_id = ?)", userID)

		return nil
	}
}

func (*queryOptions) WithUserID(userID string) store.QueryOption {
	return func(ctx context.Context) error {
		wrapper, ok := ctx.Value("query").(*queryWrapper)
		if !ok {
			return ErrQueryNotFound
		}

		col := "user_id"
		if alias, ok := ctx.Value(CtxTableAlias).(string); ok && alias != "" {
			col = alias + ".user_id"
		}

		wrapper.query = wrapper.query.Where(col+" = ?", userID)

		return nil
	}
}

func (*queryOptions) WithAPIKeyID(apiKeyID string) store.QueryOption {
	return func(ctx context.Context) error {
		wrapper, ok := ctx.Value("query").(*queryWrapper)
		if !ok {
			return ErrQueryNotFound
		}

		wrapper.query = wrapper.query.Where("api_key_id = ?", apiKeyID)

		return nil
	}
}

func (*queryOptions) WithoutAPIKeyOwner() store.QueryOption {
	return func(ctx context.Context) error {
		wrapper, ok := ctx.Value("query").(*queryWrapper)
		if !ok {
			return ErrQueryNotFound
		}

		wrapper.query = wrapper.query.Where("api_key_id IS NULL")

		return nil
	}
}

func (*queryOptions) ForUpdate() store.QueryOption {
	return func(ctx context.Context) error {
		wrapper, ok := ctx.Value("query").(*queryWrapper)
		if !ok {
			return ErrQueryNotFound
		}

		alias, ok := ctx.Value(CtxTableAlias).(string)
		if !ok || alias == "" {
			return ErrLockWithoutTableAlias
		}

		wrapper.query = wrapper.query.For("UPDATE OF ?", bun.Ident(alias))

		return nil
	}
}

// ScopeOption turns a namespace scope into the query predicate that enforces it. A bounded scope
// becomes a namespace_id predicate; an unbounded scope adds nothing, which is the whole point of it
// having to carry a reason. A scope that was never constructed is rejected.
//
// Exported so the cloud store, which builds queries against the driver directly, enforces a scope
// the same way rather than reimplementing the predicate.
func ScopeOption(sc scope.Scope) store.QueryOption {
	return func(ctx context.Context) error {
		wrapper, ok := ctx.Value("query").(*queryWrapper)
		if !ok {
			return ErrQueryNotFound
		}

		name := "namespace_id"
		if override, ok := ctx.Value(CtxNamespaceColumn).(string); ok && override != "" {
			name = override
		}

		column := name
		if alias, ok := ctx.Value(CtxTableAlias).(string); ok && alias != "" {
			column = alias + "." + name
		}

		query, err := whereInScope(wrapper.query, sc, column)
		if err != nil {
			return err
		}

		wrapper.query = query

		return nil
	}
}

func (*queryOptions) WithDeviceStatus(status models.DeviceStatus) store.QueryOption {
	return func(ctx context.Context) error {
		wrapper, ok := ctx.Value("query").(*queryWrapper)
		if !ok {
			return ErrQueryNotFound
		}

		col := "status"
		if alias, ok := ctx.Value(CtxTableAlias).(string); ok && alias != "" {
			col = alias + ".status"
		}

		wrapper.query = wrapper.query.Where(col+" = ?", string(status))

		return nil
	}
}
