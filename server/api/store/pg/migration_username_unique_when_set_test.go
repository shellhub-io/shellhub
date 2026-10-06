package pg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestUsernameUniqueWhenSetMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 42)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	seedUser(t, ctx, db, "11111111-1111-4111-8111-111111111111", "taken")

	require.NoError(t, provider.ApplyNext(ctx), "043 must apply cleanly")

	seedUserWithoutUsername(t, ctx, db, "22222222-2222-4222-8222-222222222222", "first@example.com")
	seedUserWithoutUsername(t, ctx, db, "33333333-3333-4333-8333-333333333333", "second@example.com")

	_, err = db.ExecContext(ctx, `
		INSERT INTO users
		    (id, created_at, updated_at, origin, status, name, username, email,
		     password_digest, auth_methods, namespace_ownership_limit)
		VALUES ('44444444-4444-4444-8444-444444444444', now(), now(), 'local', 'confirmed',
		        'taken', 'taken', 'other@example.com', 'hash', ARRAY['local']::user_auth_method[], -1)
	`)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	require.True(t, ok, "a second user cannot take a set username, got %v", err)
	assert.Equal(t, "23505", pgErr.Code, "a set username stays unique")
	assert.Equal(t, "users_username_key", pgErr.ConstraintName, "the store maps this name to the username field")

	execSQL(t, ctx, db, "DELETE FROM users WHERE id = '33333333-3333-4333-8333-333333333333'")
	require.NoError(t, provider.Rollback(ctx), "043 must roll back once one user without a username is left")

	_, err = db.ExecContext(ctx, `
		INSERT INTO users
		    (id, created_at, updated_at, origin, status, name, username, email,
		     password_digest, auth_methods, namespace_ownership_limit)
		VALUES ('55555555-5555-4555-8555-555555555555', now(), now(), 'saml', 'confirmed',
		        'third@example.com', '', 'third@example.com', '', ARRAY['saml']::user_auth_method[], -1)
	`)
	pgErr, ok = errors.AsType[*pgconn.PgError](err)
	require.True(t, ok, "after the rollback a second user cannot go without a username, got %v", err)
	assert.Equal(t, "users_username_key", pgErr.ConstraintName)
}

func TestUsernameUniqueWhenSetRollbackWithUsersWithoutUsername(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 42)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	require.NoError(t, provider.ApplyNext(ctx), "043 must apply cleanly")

	seedUserWithoutUsername(t, ctx, db, "22222222-2222-4222-8222-222222222222", "first@example.com")
	seedUserWithoutUsername(t, ctx, db, "33333333-3333-4333-8333-333333333333", "second@example.com")

	pgErr, ok := errors.AsType[*pgconn.PgError](provider.Rollback(ctx))
	require.True(t, ok, "043 cannot roll back while two users have no username")
	assert.Equal(t, "23505", pgErr.Code, "the restored constraint refuses the duplicate empty username")
}

func seedUserWithoutUsername(t *testing.T, ctx context.Context, db *bun.DB, id, email string) {
	t.Helper()

	execSQL(t, ctx, db, `
		INSERT INTO users
		    (id, created_at, updated_at, origin, status, name, username, email,
		     password_digest, auth_methods, namespace_ownership_limit)
		VALUES (?, now(), now(), 'saml', 'confirmed', ?, '', ?,
		        '', ARRAY['saml']::user_auth_method[], -1)
	`, id, email, email)
}
