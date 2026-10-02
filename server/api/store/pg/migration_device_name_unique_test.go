package pg_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestDeviceAcceptedNameUniqueMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 41)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID     = "11111111-1111-4111-8111-111111111111"
		tenant      = "22222222-2222-4222-8222-222222222222"
		otherTenant = "33333333-3333-4333-8333-333333333333"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)
	seedNamespace(t, ctx, db, otherTenant, "other", ownerID, fixtureTime)

	long := strings.Repeat("a", 54) + "-" + strings.Repeat("b", 9)

	devices := []struct {
		id       string
		tenant   string
		name     string
		status   string
		created  time.Time
		seen     time.Time
		expected string
		reason   string
	}{
		{deviceID("a1"), tenant, "web", "accepted", fixtureTime, fixtureTime, "web-a1a1a1a1", "an accepted device seen earlier is renamed"},
		{
			deviceID("b2"), tenant, "web", "accepted", fixtureTime.Add(time.Hour), fixtureTime.Add(5 * time.Hour),
			"web", "the accepted device seen last keeps the name",
		},
		{
			deviceID("c3"), tenant, "web", "pending", fixtureTime.Add(2 * time.Hour), fixtureTime.Add(6 * time.Hour),
			"web", "a pending device is not renamed and does not take the name",
		},
		{
			deviceID("d4"), otherTenant, "web", "accepted", fixtureTime.Add(3 * time.Hour), fixtureTime,
			"web", "another namespace is not renamed",
		},
		{deviceID("e5"), tenant, long, "accepted", fixtureTime, fixtureTime.Add(time.Hour), long, "the accepted device seen last keeps a long name"},
		{
			deviceID("f6"), tenant, long, "accepted", fixtureTime, fixtureTime,
			strings.Repeat("a", 54) + "-f6f6f6f6", "a renamed device fits the name column and drops the hyphen the cut leaves",
		},
		{deviceID("17"), tenant, "db", "accepted", fixtureTime, fixtureTime.Add(time.Hour), "db", "the accepted device seen last keeps a name"},
		{deviceID("6c"), tenant, "db-28282828", "accepted", fixtureTime, fixtureTime, "db-28282828", "a device already holding the generated name keeps it"},
		{
			"28282828" + strings.Repeat("39", 28), tenant, "db", "accepted", fixtureTime, fixtureTime,
			"db-2828282839393939", "a generated name another accepted device holds is lengthened until it is free",
		},
		{
			deviceID("5b"), tenant, "tie", "accepted", fixtureTime.Add(time.Hour), fixtureTime,
			"tie-5b5b5b5b", "on equal last seen times the newer device is renamed",
		},
		{deviceID("4a"), tenant, "tie", "accepted", fixtureTime, fixtureTime, "tie", "on equal last seen times the older device keeps the name"},
		{deviceID("7d"), tenant, "same", "accepted", fixtureTime, fixtureTime, "same", "on equal times the lower id keeps the name"},
		{deviceID("8e"), tenant, "same", "accepted", fixtureTime, fixtureTime, "same-8e8e8e8e", "on equal times the higher id is renamed"},
	}

	for i, device := range devices {
		seedDevice(t, ctx, db, device.id, device.tenant, device.name, fmt.Sprintf("00:00:00:00:01:%02x", i), device.status, device.created)
		execSQL(t, ctx, db, "UPDATE devices SET last_seen = ? WHERE id = ?", device.seen, device.id)
	}

	require.NoError(t, provider.ApplyNext(ctx), "042 must apply cleanly over duplicated names")

	for _, device := range devices {
		assert.Equal(t, device.expected, deviceName(t, ctx, db, device.id), device.reason)
	}

	_, err = db.ExecContext(ctx, "UPDATE devices SET name = 'web' WHERE id = ?", deviceID("a1"))
	require.Error(t, err, "a second accepted device cannot take a name")

	_, err = db.ExecContext(ctx, "UPDATE devices SET status = 'accepted' WHERE id = ?", deviceID("c3"))
	require.Error(t, err, "a pending device cannot be accepted under a taken name")

	require.NoError(t, provider.Rollback(ctx), "042 must roll back cleanly")

	execSQL(t, ctx, db, "UPDATE devices SET name = 'web' WHERE id = ?", deviceID("a1"))
}

func deviceID(prefix string) string {
	return strings.Repeat(prefix, 32)
}

func deviceName(t *testing.T, ctx context.Context, db *bun.DB, id string) string {
	t.Helper()

	var name string
	require.NoError(t, db.NewRaw("SELECT name FROM devices WHERE id = ?", id).Scan(ctx, &name))

	return name
}
