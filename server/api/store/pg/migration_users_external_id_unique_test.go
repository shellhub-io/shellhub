package pg_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestUsersExternalIDUniqueMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 43)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	seedUserBoundTo(t, ctx, db, "11111111-1111-4111-8111-111111111111", "former", "shared-subject")
	seedUserBoundTo(t, ctx, db, "22222222-2222-4222-8222-222222222222", "rebound", "shared-subject")
	seedUserBoundTo(t, ctx, db, "33333333-3333-4333-8333-333333333333", "bound", "own-subject")
	seedUser(t, ctx, db, "44444444-4444-4444-8444-444444444444", "local")

	require.NoError(t, provider.ApplyNext(ctx), "044 must apply cleanly over a subject bound to two users")

	for _, id := range []string{
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222",
		"33333333-3333-4333-8333-333333333333",
	} {
		assert.False(t, externalIDOf(t, ctx, db, id).Valid,
			"every binding is cleared, since one saved from a transient NameID cannot be told apart")
	}

	execSQL(t, ctx, db,
		"UPDATE users SET external_id = 'own-subject' WHERE id = '33333333-3333-4333-8333-333333333333'")

	_, err = db.ExecContext(ctx,
		"UPDATE users SET external_id = 'own-subject' WHERE id = '44444444-4444-4444-8444-444444444444'")
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	require.True(t, ok, "a second user cannot take a bound subject, got %v", err)
	assert.Equal(t, "23505", pgErr.Code)
	assert.Equal(t, "users_external_id_key", pgErr.ConstraintName)

	require.NoError(t, provider.Rollback(ctx), "044 must roll back")

	execSQL(t, ctx, db,
		"UPDATE users SET external_id = 'own-subject' WHERE id = '44444444-4444-4444-8444-444444444444'")
}

func seedUserBoundTo(t *testing.T, ctx context.Context, db *bun.DB, id, name, externalID string) {
	t.Helper()

	seedUser(t, ctx, db, id, name)
	execSQL(t, ctx, db, "UPDATE users SET origin = 'saml', external_id = ? WHERE id = ?", externalID, id)
}

func externalIDOf(t *testing.T, ctx context.Context, db *bun.DB, id string) sql.NullString {
	t.Helper()

	var externalID sql.NullString
	require.NoError(t, db.NewRaw("SELECT external_id FROM users WHERE id = ?", id).Scan(ctx, &externalID))

	return externalID
}
