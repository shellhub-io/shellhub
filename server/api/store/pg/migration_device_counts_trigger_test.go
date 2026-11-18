package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceCountsTriggerMigration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 60)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID = "11111111-1111-4111-8111-111111111111"
		tenant  = "22222222-2222-4222-8222-222222222222"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)
	seedDevice(t, ctx, db, "33333333-3333-4333-8333-333333333331", tenant, "a", "00:00:00:00:00:01", "accepted", fixtureTime)
	seedDevice(t, ctx, db, "33333333-3333-4333-8333-333333333332", tenant, "b", "00:00:00:00:00:02", "pending", fixtureTime)
	execSQL(t, ctx, db, "UPDATE namespaces SET devices_accepted_count = 40, devices_removed_count = 7 WHERE id = ?", tenant)

	counts := func() [4]int64 {
		t.Helper()

		var c [4]int64
		require.NoError(t, db.NewRaw(`
			SELECT devices_accepted_count, devices_pending_count, devices_rejected_count, devices_removed_count
			FROM namespaces WHERE id = ?`, tenant).Scan(ctx, &c[0], &c[1], &c[2], &c[3]))

		return c
	}

	require.NoError(t, provider.ApplyNext(ctx), "061 must apply cleanly over release-shaped data")
	assert.Equal(t, [4]int64{1, 1, 0, 0}, counts(), "061 rebuilds drifted counters from the devices")

	seedDevice(t, ctx, db, "33333333-3333-4333-8333-333333333333", tenant, "c", "00:00:00:00:00:03", "rejected", fixtureTime)
	execSQL(t, ctx, db, "UPDATE devices SET status = 'removed' WHERE id = ?", "33333333-3333-4333-8333-333333333331")
	execSQL(t, ctx, db, "DELETE FROM devices WHERE id = ?", "33333333-3333-4333-8333-333333333332")
	assert.Equal(t, [4]int64{0, 0, 1, 1}, counts(), "device writes keep the counters")

	execSQL(t, ctx, db, "DELETE FROM namespaces WHERE id = ?", tenant)

	require.NoError(t, provider.Rollback(ctx), "061 must roll back cleanly")

	var triggers int
	require.NoError(t, db.NewRaw("SELECT count(*) FROM pg_trigger WHERE tgname LIKE 'devices_count_%'").Scan(ctx, &triggers))
	assert.Zero(t, triggers)
}
