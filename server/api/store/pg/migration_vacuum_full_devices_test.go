package pg_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/server/api/store/pg/migrations"
	"github.com/shellhub-io/shellhub/server/api/store/pg/options"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration022SurvivesItsLockTimeout(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 21)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	lockTimeout := migrations.VacuumLockTimeout
	migrations.VacuumLockTimeout = 100 * time.Millisecond
	t.Cleanup(func() { migrations.VacuumLockTimeout = lockTimeout })

	holder, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = holder.Rollback() })

	_, err = holder.ExecContext(ctx, "LOCK TABLE devices IN ACCESS SHARE MODE")
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() { done <- options.Migrate()(ctx, db) }()

	recorded := assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		var recorded bool
		assert.NoError(tt, db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM bun_migrations WHERE name = '022')").Scan(&recorded))
		assert.True(tt, recorded)
	}, 30*time.Second, 50*time.Millisecond)
	if !recorded {
		require.NoError(t, holder.Rollback())
		require.FailNow(t, "022 was never recorded", "Migrate: %v", <-done)
	}

	require.NoError(t, holder.Rollback())
	require.NoError(t, <-done)

	names := recordedMigrations(t, ctx, db, "bun_migrations")
	assert.True(t, slices.Contains(names, migrations.All[len(migrations.All)-1].Name), "the migrations after 022 must be applied")
}
