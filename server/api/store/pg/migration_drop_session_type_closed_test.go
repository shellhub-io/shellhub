package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDropSessionTypeClosedMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 36)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		tenant    = "22222222-2222-4222-8222-222222222222"
		ownerID   = "11111111-1111-4111-8111-111111111111"
		deviceUID = "device-1"
		liveUID   = "session-live"
		endedUID  = "session-ended"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)

	execSQL(t, ctx, db, `
		INSERT INTO devices
		    (id, namespace_id, created_at, updated_at, last_seen, status, status_updated_at,
		     name, mac, public_key, identifier, pretty_name, version, arch, platform)
		VALUES (?, ?, now(), now(), now(), 'accepted'::device_status, now(),
		        'dev', 'aa:bb:cc:dd:ee:01', 'pk', 'arch', 'Arch Linux', 'v1.2.3', 'amd64', 'docker')
	`, deviceUID, tenant)

	seedSession := func(uid string, closed bool) {
		t.Helper()

		execSQL(t, ctx, db, `
			INSERT INTO sessions
			    (id, namespace_id, device_id, username, ip_address, started_at, seen_at,
			     closed, authenticated, recorded, type, term, created_at, updated_at)
			VALUES (?, ?, ?, 'root', '127.0.0.1', now(), now(),
			        ?, true, false, 'none'::session_type, 'none', now(), now())
		`, uid, tenant, deviceUID, closed)
	}

	seedSession(liveUID, false)
	seedSession(endedUID, true)

	execSQL(t, ctx, db, `
		INSERT INTO active_sessions (session_id, seen_at, created_at)
		VALUES (?, now(), now())
	`, liveUID)

	require.NoError(t, provider.ApplyNext(ctx), "037 must apply cleanly over release-shaped data")

	t.Run("drops both columns", func(t *testing.T) {
		for _, column := range []string{"type", "closed"} {
			count, err := db.NewSelect().
				Table("information_schema.columns").
				Where("table_name = 'sessions' AND column_name = ?", column).
				Count(ctx)
			require.NoError(t, err)
			assert.Zero(t, count, "sessions.%s must be gone", column)
		}
	})

	t.Run("drops the session_type enum, which nothing else names", func(t *testing.T) {
		count, err := db.NewSelect().
			Table("pg_type").
			Where("typname = 'session_type'").
			Count(ctx)
		require.NoError(t, err)
		assert.Zero(t, count)
	})

	t.Run("keeps every session and its active-session membership", func(t *testing.T) {
		sessions, err := db.NewSelect().Table("sessions").Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, sessions, "dropping a column must take no session with it")

		var live []string
		require.NoError(t, db.NewSelect().
			Table("active_sessions").
			Column("session_id").
			Scan(ctx, &live))
		assert.Equal(t, []string{liveUID}, live,
			"membership is what says a session is running, and it is untouched")
	})
}
