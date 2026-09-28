package pg_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestDeviceOwnerMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 40)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID  = "11111111-1111-4111-8111-111111111111"
		memberID = "66666666-6666-4666-8666-666666666666"
		tenant   = "22222222-2222-4222-8222-222222222222"
		deviceID = "33333333-3333-4333-8333-333333333333"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedUser(t, ctx, db, memberID, "member")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)

	execSQL(t, ctx, db, `
		INSERT INTO memberships (user_id, namespace_id, created_at, updated_at, role)
		VALUES (?, ?, now(), now(), 'owner'), (?, ?, now(), now(), 'operator')
	`, ownerID, tenant, memberID, tenant)

	execSQL(t, ctx, db, `
		INSERT INTO devices (id, namespace_id, name, mac, public_key, status, custom_fields,
		                     remote_addr, ephemeral, ephemeral_timeout, created_at, updated_at, last_seen)
		VALUES (?, ?, 'dev', '00:00:00:00:00:01', '', 'accepted', '{}', '10.0.0.1', false, 0, now(), now(), now())
	`, deviceID, tenant)

	require.NoError(t, provider.ApplyNext(ctx), "041 must apply cleanly over release-shaped data")

	var owner *string
	require.NoError(t, db.NewRaw("SELECT owner_id::text FROM devices WHERE id = ?", deviceID).Scan(ctx, &owner))
	assert.Nil(t, owner, "an existing device is a team device")

	execSQL(t, ctx, db, "UPDATE devices SET owner_id = ? WHERE id = ?", memberID, deviceID)

	deleteMembership := func() error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			_, err := tx.ExecContext(ctx, "DELETE FROM memberships WHERE user_id = ? AND namespace_id = ?", memberID, tenant)

			return err
		})
	}

	require.Error(t, deleteMembership(), "a membership that still owns a device cannot be deleted")

	_, err = db.ExecContext(ctx, "UPDATE devices SET owner_id = ? WHERE id = ?", "99999999-9999-4999-9999-999999999999", deviceID)
	require.Error(t, err, "an owner that is not a member is refused")

	execSQL(t, ctx, db, "DELETE FROM namespaces WHERE id = ?", tenant)

	var remaining int
	require.NoError(t, db.NewRaw("SELECT count(*) FROM devices").Scan(ctx, &remaining))
	assert.Zero(t, remaining, "deleting a namespace with an owned device succeeds")

	require.NoError(t, provider.Rollback(ctx), "041 must roll back cleanly")

	var columns int
	require.NoError(t, db.NewRaw(`
		SELECT count(*) FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'owner_id'
	`).Scan(ctx, &columns))
	assert.Zero(t, columns)
}
