package pg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssuedTenantsMigration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 56)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID     = "11111111-1111-4111-8111-111111111111"
		tenant      = "22222222-2222-4222-8222-222222222222"
		otherTenant = "33333333-3333-4333-8333-333333333333"
		unissued    = "44444444-4444-4444-8444-444444444444"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)
	seedNamespace(t, ctx, db, otherTenant, "other", ownerID, fixtureTime)

	require.NoError(t, provider.ApplyNext(ctx), "057 must apply over existing namespaces")

	var issued []string
	require.NoError(t, db.NewRaw("SELECT id::text FROM issued_tenants ORDER BY id").Scan(ctx, &issued))
	assert.Equal(t, []string{tenant, otherTenant}, issued, "every existing namespace's tenant is recorded as issued")

	_, err = db.ExecContext(ctx, `
		INSERT INTO namespaces (id, created_at, updated_at, scope, name, owner_id, max_devices, record_sessions)
		VALUES (?, now(), now(), 'personal', 'unissued', ?, -1, false)
	`, unissued, ownerID)
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	require.True(t, ok, "a namespace whose tenant was never issued is refused, got %v", err)
	assert.Equal(t, "23503", pgErr.Code)
	assert.Equal(t, "namespaces_issued_tenant_fkey", pgErr.ConstraintName)

	execSQL(t, ctx, db, "DELETE FROM namespaces WHERE id = ?", tenant)

	_, err = db.ExecContext(ctx, "INSERT INTO issued_tenants (id, created_at) VALUES (?, now())", tenant)
	pgErr, ok = errors.AsType[*pgconn.PgError](err)
	require.True(t, ok, "a deleted namespace's tenant stays issued, got %v", err)
	assert.Equal(t, "23505", pgErr.Code)
	assert.Equal(t, "issued_tenants_pkey", pgErr.ConstraintName)

	_, err = db.ExecContext(ctx, "DELETE FROM issued_tenants WHERE id = ?", tenant)
	pgErr, ok = errors.AsType[*pgconn.PgError](err)
	require.True(t, ok, "a deleted namespace's tenant cannot be released, got %v", err)
	assert.Equal(t, "23001", pgErr.Code)

	_, err = db.ExecContext(ctx, "UPDATE issued_tenants SET id = ? WHERE id = ?", unissued, tenant)
	pgErr, ok = errors.AsType[*pgconn.PgError](err)
	require.True(t, ok, "an issued tenant cannot be rewritten, got %v", err)
	assert.Equal(t, "23001", pgErr.Code)

	require.NoError(t, provider.Rollback(ctx), "057 must roll back cleanly")
}
