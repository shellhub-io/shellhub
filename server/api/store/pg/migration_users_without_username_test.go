package pg_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestUsersWithoutUsernameMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 42)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		firstSAMLUser  = "11111111-1111-4111-8111-111111111111"
		secondSAMLUser = "22222222-2222-4222-8222-222222222222"
		owner          = "33333333-3333-4333-8333-333333333333"
		ownerClash     = "44444444-4444-4444-8444-444444444444"
	)

	seedSAMLUser(t, ctx, db, firstSAMLUser, "")
	seedUser(t, ctx, db, owner, "owner")

	require.NoError(t, provider.ApplyNext(ctx), "043 must apply cleanly")

	seedSAMLUser(t, ctx, db, secondSAMLUser, "")

	requireUsernameTaken(t, insertSAMLUser(ctx, db, ownerClash, "owner"))

	execSQL(t, ctx, db, "DELETE FROM users WHERE id = ?", secondSAMLUser)
	require.NoError(t, provider.Rollback(ctx), "043 must roll back while at most one user has no username")

	requireUsernameTaken(t, insertSAMLUser(ctx, db, secondSAMLUser, ""))
}

func TestUsersWithoutUsernameMigrationRefusesRollback(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 42)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		firstSAMLUser  = "11111111-1111-4111-8111-111111111111"
		secondSAMLUser = "22222222-2222-4222-8222-222222222222"
		thirdSAMLUser  = "33333333-3333-4333-8333-333333333333"
		owner          = "44444444-4444-4444-8444-444444444444"
		ownerClash     = "55555555-5555-4555-8555-555555555555"
	)

	require.NoError(t, provider.ApplyNext(ctx), "043 must apply cleanly")

	seedSAMLUser(t, ctx, db, firstSAMLUser, "")
	seedSAMLUser(t, ctx, db, secondSAMLUser, "")

	requireUsernameTaken(t, provider.Rollback(ctx))

	seedSAMLUser(t, ctx, db, thirdSAMLUser, "")
	seedUser(t, ctx, db, owner, "owner")
	requireUsernameTaken(t, insertSAMLUser(ctx, db, ownerClash, "owner"))
}

func requireUsernameTaken(t *testing.T, err error) {
	t.Helper()

	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	assert.Equal(t, "23505", pgErr.Code)
	assert.Equal(t, "users_username_key", pgErr.ConstraintName, "the store maps a taken username by this name")
}

func seedSAMLUser(t *testing.T, ctx context.Context, db *bun.DB, id, username string) {
	t.Helper()

	require.NoError(t, insertSAMLUser(ctx, db, id, username))
}

func insertSAMLUser(ctx context.Context, db *bun.DB, id, username string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO users
		    (id, created_at, updated_at, origin, status, name, username, email,
		     password_digest, auth_methods, namespace_ownership_limit)
		VALUES (?, now(), now(), 'saml', 'confirmed', ?, ?, ?,
		        '', ARRAY['saml']::user_auth_method[], -1)
	`, id, id, username, id+"@example.com")

	return err
}
