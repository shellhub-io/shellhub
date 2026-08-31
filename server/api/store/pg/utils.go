package pg

import (
	"context"
	"database/sql"
	"errors"
	"io"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/uptrace/bun"
)

// constraintToField maps a PostgreSQL unique-constraint name on the users table to
// the user-facing field name it protects. The constraint names here are defined in
// migration 001_initial_schema — renaming those constraints silently breaks this mapping.
//
// Only constraints from the users table are listed; other tables' 23505 violations
// will return an empty string and fall through to a bare store.ErrDuplicate.
func constraintToField(constraint string) string {
	switch constraint {
	case constraintUsersEmailKey:
		return "email"
	case constraintUsersUsernameKey:
		return "username"
	default:
		return ""
	}
}

// constraintUsersEmailKey and constraintUsersUsernameKey are the PostgreSQL unique-constraint
// names for the users table, as created by migration 001_initial_schema.
// WARNING: renaming these constraints in a migration silently breaks the constraintToField
// mapping — update both together.
const (
	constraintUsersEmailKey    = "users_email_key"
	constraintUsersUsernameKey = "users_username_key"
)

// constraintSystemsInstanceTenantIDFkey is the FK from systems.instance_tenant_id to
// namespaces.id, created ON DELETE RESTRICT by migration 009. A violation means a caller
// tried to delete the namespace bound to the instance.
// WARNING: renaming it in the migration silently breaks the mapping below.
const constraintSystemsInstanceTenantIDFkey = "systems_instance_tenant_id_fkey"

func fromSQLError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, sql.ErrNoRows), errors.Is(err, io.EOF):
		return store.ErrNoDocuments
	default:
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				if field := constraintToField(pgErr.ConstraintName); field != "" {
					return errors.Join(store.ErrDuplicate, store.DuplicateFieldError{Field: field})
				}

				return store.ErrDuplicate
			}

			if (pgErr.Code == "23001" || pgErr.Code == "23503") && pgErr.ConstraintName == constraintSystemsInstanceTenantIDFkey {
				return store.ErrNamespaceInstanceProtected
			}
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		return errors.Join(err, store.ErrInternal)
	}
}

type ctxKey string

// CtxTableAlias is the context key used to pass a table alias to query options
// like the scope predicate, avoiding column ambiguity in queries with JOINs.
const CtxTableAlias ctxKey = "table_alias"

// CtxNamespaceColumn overrides the column the scope predicate bounds on. Almost every table names it
// namespace_id; membership_invitations is the exception, calling it tenant_id.
const CtxNamespaceColumn ctxKey = "namespace_column"

// queryWrapper wraps a SelectQuery pointer to allow mutations
type queryWrapper struct {
	query *bun.SelectQuery
}

func applyOptions(ctx context.Context, query *bun.SelectQuery, opts ...store.QueryOption) (*bun.SelectQuery, error) {
	wrapper := &queryWrapper{query: query}
	ctxWithQuery := context.WithValue(ctx, "query", wrapper)

	for _, opt := range opts {
		if err := opt(ctxWithQuery); err != nil {
			return wrapper.query, fromSQLError(err)
		}
	}

	return wrapper.query, nil
}

// requireBounded returns the namespace an operation must target. It rejects an unbounded scope,
// which the operations using it cannot express — there is no writing a membership, or counting a
// namespace's devices, across every namespace at once.
func requireBounded(sc scope.Scope) (string, error) {
	if !sc.IsBounded() {
		return "", store.ErrInvalidScope
	}

	return sc.TenantID(), nil
}

// applyScopedOptions applies the namespace scope ahead of the caller's options, so the bound is part
// of the query before anything else can shape it.
func applyScopedOptions(ctx context.Context, query *bun.SelectQuery, sc scope.Scope, opts ...store.QueryOption) (*bun.SelectQuery, error) {
	return applyOptions(ctx, query, append([]store.QueryOption{ScopeOption(sc)}, opts...)...)
}

// ApplyOptions is the exported version of applyOptions, allowing external packages
// (e.g. cloud store) to reuse the same query-option mechanism.
func ApplyOptions(ctx context.Context, query *bun.SelectQuery, opts ...store.QueryOption) (*bun.SelectQuery, error) {
	return applyOptions(ctx, query, opts...)
}

// ApplyScopedOptions is the exported version of applyScopedOptions, for the cloud store.
func ApplyScopedOptions(ctx context.Context, query *bun.SelectQuery, sc scope.Scope, opts ...store.QueryOption) (*bun.SelectQuery, error) {
	return applyScopedOptions(ctx, query, sc, opts...)
}
