package pg_test

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevicePresenceMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 58)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		tenant  = "22222222-2222-4222-8222-222222222222"
		ownerID = "11111111-1111-4111-8111-111111111111"
	)

	seen := fixtureTime.Add(time.Hour)
	closedAt := seen.Add(time.Minute)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)

	seedDevice := func(id, mac string, disconnectedAt *time.Time) {
		t.Helper()

		execSQL(t, ctx, db, `
			INSERT INTO devices
			    (id, namespace_id, created_at, updated_at, last_seen, disconnected_at, status, status_updated_at,
			     name, mac, public_key, identifier, pretty_name, version, arch, platform)
			VALUES (?, ?, now(), now(), ?, ?, 'accepted'::device_status, now(),
			        ?, ?, 'pk', 'arch', 'Arch Linux', 'v1.2.3', 'amd64', 'docker')
		`, id, tenant, seen, disconnectedAt, id, mac)
	}

	seedDevice("online", "aa:bb:cc:dd:ee:01", nil)
	seedDevice("closed", "aa:bb:cc:dd:ee:02", &closedAt)

	require.NoError(t, provider.ApplyNext(ctx))

	type presence struct {
		LastSeen       time.Time  `bun:"last_seen"`
		DisconnectedAt *time.Time `bun:"disconnected_at"`
	}

	readPresence := func(t *testing.T, id string) presence {
		t.Helper()

		var row presence
		require.NoError(t, db.NewRaw(
			"SELECT last_seen, disconnected_at FROM device_presence WHERE device_id = ?", id,
		).Scan(ctx, &row))

		return row
	}

	online := readPresence(t, "online")
	assert.True(t, seen.Equal(online.LastSeen))
	assert.Nil(t, online.DisconnectedAt)

	closed := readPresence(t, "closed")
	assert.True(t, seen.Equal(closed.LastSeen))
	require.NotNil(t, closed.DisconnectedAt)
	assert.True(t, closedAt.Equal(*closed.DisconnectedAt))

	moved, err := db.NewSelect().
		TableExpr("information_schema.columns").
		Where("table_name = 'devices'").
		Where("column_name IN ('last_seen', 'disconnected_at')").
		Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, moved)

	require.NoError(t, provider.Rollback(ctx))

	var restored presence
	require.NoError(t, db.NewRaw(
		"SELECT last_seen, disconnected_at FROM devices WHERE id = 'closed'",
	).Scan(ctx, &restored))
	assert.True(t, seen.Equal(restored.LastSeen))
	require.NotNil(t, restored.DisconnectedAt)
	assert.True(t, closedAt.Equal(*restored.DisconnectedAt))

	require.NoError(t, provider.ApplyNext(ctx))

	execSQL(t, ctx, db, "DELETE FROM devices WHERE id = 'closed'")

	remaining, err := db.NewSelect().TableExpr("device_presence").Count(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 1, remaining, "deleting a device must delete its presence")
}
