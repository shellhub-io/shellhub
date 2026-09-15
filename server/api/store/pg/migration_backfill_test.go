package pg_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstallKeyEventBackfillMigration covers migration 014: it writes exactly one registration event
// for every device attributed to a key but missing one, freezing the decision for already accepted or
// rejected devices while leaving pending ones open (NULL decided_status, which is what surfaces the
// accept control), and it never touches a device that already has an event or one with no key.
func TestInstallKeyEventBackfillMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 13)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID   = "11111111-1111-4111-8111-111111111111"
		tenant    = "22222222-2222-4222-8222-222222222222"
		legacyKey = "3333333333333333333333333333333333333333333333333333333333333333"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)

	execSQL(t, ctx, db, `
		INSERT INTO install_keys
		    (key_digest, namespace_id, name, reusable, mode, system, user_id, created_at, updated_at)
		VALUES (?, ?, 'legacy', true, 'manual', true, ?, now(), now())
	`, legacyKey, tenant, ownerID)

	mkDevice := func(uidSeed, mac, status, keyID string) string {
		t.Helper()

		uid := uidSeed + strings.Repeat("0", 64-len(uidSeed))

		execSQL(t, ctx, db, `
			INSERT INTO devices
			    (id, namespace_id, created_at, updated_at, last_seen, status, status_updated_at,
			     name, mac, public_key, identifier, pretty_name, version, arch, platform, install_key_id)
			VALUES (?, ?, now(), now(), now(), ?::device_status, now(),
			        ?, ?, ?, 'arch', 'Arch Linux', 'v1.2.3', 'amd64', 'docker', ?)
		`, uid, tenant, status, uidSeed, mac, "pk-"+uidSeed, sql.NullString{String: keyID, Valid: keyID != ""})

		return uid
	}

	pendingUID := mkDevice("aa", "aa:bb:cc:dd:ee:01", "pending", legacyKey)
	acceptedUID := mkDevice("bb", "aa:bb:cc:dd:ee:02", "accepted", legacyKey)
	withEventUID := mkDevice("cc", "aa:bb:cc:dd:ee:03", "accepted", legacyKey)
	keylessUID := mkDevice("dd", "aa:bb:cc:dd:ee:04", "pending", "")
	rejectedUID := mkDevice("ee", "aa:bb:cc:dd:ee:05", "rejected", legacyKey)

	execSQL(t, ctx, db, `
		INSERT INTO install_key_events
		    (id, install_key_id, namespace_id, device_uid, hostname, created_at)
		VALUES (gen_random_uuid(), ?, ?, ?, 'cc', now())
	`, legacyKey, tenant, withEventUID)

	require.NoError(t, provider.ApplyNext(ctx), "014 must apply cleanly")

	type row struct {
		DeviceUID  string `bun:"device_uid"`
		Decided    string `bun:"decided_status"`
		InfoID     string `bun:"info_id"`
		InfoPretty string `bun:"info_pretty_name"`
		InfoArch   string `bun:"info_arch"`
	}
	events := make(map[string][]row)
	var rows []row
	require.NoError(t, db.
		NewRaw("SELECT device_uid, coalesce(decided_status, '') AS decided_status, coalesce(info_id, '') AS info_id, coalesce(info_pretty_name, '') AS info_pretty_name, coalesce(info_arch, '') AS info_arch FROM install_key_events").
		Scan(ctx, &rows))
	for _, r := range rows {
		events[r.DeviceUID] = append(events[r.DeviceUID], r)
	}

	require.Len(t, events[pendingUID], 1, "a pending device without an event gets one")
	assert.Empty(t, events[pendingUID][0].Decided, "a pending device's backfilled event stays open so the accept control shows")
	assert.Equal(t, "arch", events[pendingUID][0].InfoID, "the device's distro id (identifier) is copied so the icon renders")
	assert.Equal(t, "Arch Linux", events[pendingUID][0].InfoPretty)
	assert.Equal(t, "amd64", events[pendingUID][0].InfoArch)

	require.Len(t, events[acceptedUID], 1, "an accepted device without an event gets one")
	assert.Equal(t, "accepted", events[acceptedUID][0].Decided, "an accepted device's decision is frozen on the event")
	assert.Equal(t, "arch", events[acceptedUID][0].InfoID, "OS facts ride along regardless of decision")

	require.Len(t, events[rejectedUID], 1, "a rejected device without an event gets one")
	assert.Equal(t, "rejected", events[rejectedUID][0].Decided, "a rejected device's decision is frozen too, so the accept control stays hidden")

	require.Len(t, events[withEventUID], 1, "a device that already had an event is not given a second one")

	assert.Empty(t, events[keylessUID], "a device with no key is skipped")
}
