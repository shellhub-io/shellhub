package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDropSessionTermMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 38)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		tenant   = "22222222-2222-4222-8222-222222222222"
		ownerID  = "11111111-1111-4111-8111-111111111111"
		deviceID = "33333333-3333-4333-8333-333333333333"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)

	execSQL(t, ctx, db, `
		INSERT INTO devices (id, namespace_id, name, mac, public_key, status, custom_fields,
		                     remote_addr, ephemeral, ephemeral_timeout, created_at, updated_at, last_seen)
		VALUES (?, ?, 'dev', '00:00:00:00:00:01', '', 'accepted', '{}', '10.0.0.1', false, 0, now(), now(), now())
	`, deviceID, tenant)

	session := func(id, term string) {
		t.Helper()

		execSQL(t, ctx, db, `
			INSERT INTO sessions (id, namespace_id, device_id, username, ip_address, term,
			                      started_at, seen_at, created_at, updated_at)
			VALUES (?, ?, ?, 'root', '10.0.0.1', ?, now(), now(), now(), now())
		`, id, tenant, deviceID, term)
	}

	session("session-with-the-placeholder", "none")
	session("session-with-a-terminal", "xterm-256color")

	require.NoError(t, provider.ApplyNext(ctx), "039 must apply cleanly over release-shaped data")

	var columns int
	require.NoError(t, db.NewRaw(`
		SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'sessions' AND column_name = 'term'
	`).Scan(ctx, &columns))
	assert.Zero(t, columns, "sessions.term must be gone")

	var ids []string
	require.NoError(t, db.NewRaw("SELECT id FROM sessions ORDER BY id").Scan(ctx, &ids))
	assert.Equal(t, []string{"session-with-a-terminal", "session-with-the-placeholder"}, ids,
		"dropping the column must not take the sessions with it")
}
