package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecountDeviceCountsMigration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 62)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID       = "11111111-1111-4111-8111-111111111111"
		driftedTenant = "22222222-2222-4222-8222-222222222221"
		correctTenant = "22222222-2222-4222-8222-222222222222"
		emptyTenant   = "22222222-2222-4222-8222-222222222223"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, driftedTenant, "drifted", ownerID, fixtureTime)
	seedNamespace(t, ctx, db, correctTenant, "correct", ownerID, fixtureTime)
	seedNamespace(t, ctx, db, emptyTenant, "empty", ownerID, fixtureTime)

	seedDevice(t, ctx, db, "33333333-3333-4333-8333-333333333331", driftedTenant, "a", "00:00:00:00:00:01", "accepted", fixtureTime)
	seedDevice(t, ctx, db, "33333333-3333-4333-8333-333333333332", driftedTenant, "b", "00:00:00:00:00:02", "accepted", fixtureTime)
	seedDevice(t, ctx, db, "33333333-3333-4333-8333-333333333333", driftedTenant, "c", "00:00:00:00:00:03", "pending", fixtureTime)
	seedDevice(t, ctx, db, "33333333-3333-4333-8333-333333333334", driftedTenant, "d", "00:00:00:00:00:04", "removed", fixtureTime)
	seedDevice(t, ctx, db, "33333333-3333-4333-8333-333333333335", correctTenant, "e", "00:00:00:00:00:05", "rejected", fixtureTime)

	execSQL(t, ctx, db, `
		UPDATE namespaces SET devices_accepted_count = 4, devices_pending_count = 0,
		                      devices_rejected_count = 3, devices_removed_count = 5
		WHERE id = ?`, driftedTenant)
	execSQL(t, ctx, db, "UPDATE namespaces SET devices_pending_count = 2 WHERE id = ?", emptyTenant)

	counts := func(tenant string) [4]int64 {
		t.Helper()

		var c [4]int64
		require.NoError(t, db.NewRaw(`
			SELECT devices_accepted_count, devices_pending_count, devices_rejected_count, devices_removed_count
			FROM namespaces WHERE id = ?`, tenant).Scan(ctx, &c[0], &c[1], &c[2], &c[3]))

		return c
	}

	rowVersion := func(tenant string) string {
		t.Helper()

		var xmin string
		require.NoError(t, db.NewRaw("SELECT xmin::text FROM namespaces WHERE id = ?", tenant).Scan(ctx, &xmin))

		return xmin
	}

	require.Equal(t, [4]int64{0, 0, 1, 0}, counts(correctTenant), "the triggers count the seeded devices")
	correctVersion := rowVersion(correctTenant)

	require.NoError(t, provider.ApplyNext(ctx), "063 must apply cleanly")
	assert.Equal(t, [4]int64{2, 1, 0, 1}, counts(driftedTenant), "063 rebuilds drifted counters from the devices")
	assert.Equal(t, [4]int64{0, 0, 1, 0}, counts(correctTenant), "063 keeps correct counters")
	assert.Equal(t, correctVersion, rowVersion(correctTenant), "063 does not rewrite a namespace whose counters are correct")
	assert.Equal(t, [4]int64{0, 0, 0, 0}, counts(emptyTenant), "063 zeroes the counters of a namespace without devices")

	require.NoError(t, provider.Rollback(ctx), "063 must roll back cleanly")
	assert.Equal(t, [4]int64{2, 1, 0, 1}, counts(driftedTenant), "rolling 063 back keeps the recounted counters")
}
