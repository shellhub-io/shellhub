package pg

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/uptrace/bun"
)

// TODO: maybe these methods can be deprecated with bun

func (pg *pg) Options() store.QueryOptions {
	return pg.options
}

func (*queryOptions) Paginate(offset, limit int) store.QueryOption {
	return func(ctx context.Context) error {
		query, ok := ctx.Value("query").(*bun.SelectQuery)
		if !ok {
			return errors.New("query not found in context")
		}

		query = query.Offset(offset).Limit(limit)

		return nil
	}
}

func (*queryOptions) Order(column, direction string) store.QueryOption {
	return func(ctx context.Context) error {
		query, ok := ctx.Value("query").(*bun.SelectQuery)
		if !ok {
			return errors.New("query not found in context")
		}

		query = query.OrderExpr("? ?", bun.Ident(column), bun.Safe(strings.ToUpper(direction)))

		return nil
	}
}

func (*queryOptions) WithMember(userID string) store.QueryOption {
	return func(ctx context.Context) error {
		query, ok := ctx.Value("query").(*bun.SelectQuery)
		if !ok {
			return errors.New("query not found in context")
		}

		query = query.Where("EXISTS (SELECT 1 FROM memberships WHERE memberships.namespace_id = namespace.id AND memberships.user_id = ?)", userID)

		return nil
	}
}

func (*queryOptions) CountAcceptedDevices() store.NamespaceQueryOption {
	return nil
}

func (*queryOptions) EnrichMembersData() store.NamespaceQueryOption {
	return nil
}
