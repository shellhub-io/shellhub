package pg_test

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestMembershipInvitationPendingUniqueMigration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 55)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		ownerID     = "11111111-1111-4111-8111-111111111111"
		tenant      = "22222222-2222-4222-8222-222222222222"
		otherTenant = "33333333-3333-4333-8333-333333333333"
		invitee     = "44444444-4444-4444-8444-444444444444"
		otherUser   = "55555555-5555-4555-8555-555555555555"
		hiddenUser  = "77777777-7777-4777-8777-777777777777"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)
	seedNamespace(t, ctx, db, otherTenant, "other", ownerID, fixtureTime)

	invitations := []struct {
		id       string
		tenant   string
		user     string
		status   string
		created  time.Time
		expected string
		reason   string
	}{
		{invitationID("a"), tenant, invitee, "pending", fixtureTime, "cancelled", "an older pending duplicate is cancelled"},
		{invitationID("b"), tenant, invitee, "pending", fixtureTime.Add(2 * time.Hour), "pending", "the newest pending invitation stays pending"},
		{invitationID("c"), tenant, invitee, "pending", fixtureTime.Add(time.Hour), "cancelled", "a middle pending duplicate is cancelled"},
		{invitationID("d"), tenant, invitee, "cancelled", fixtureTime.Add(-time.Hour), "cancelled", "an older cancelled invitation is untouched"},
		{invitationID("e"), otherTenant, invitee, "pending", fixtureTime, "pending", "the same user in another namespace is untouched"},
		{invitationID("f"), tenant, otherUser, "pending", fixtureTime, "pending", "another user in the namespace is untouched"},
		{invitationID("1"), tenant, hiddenUser, "pending", fixtureTime, "cancelled", "a pending invitation behind a newer cancelled one is cancelled"},
		{invitationID("2"), tenant, hiddenUser, "cancelled", fixtureTime.Add(time.Hour), "cancelled", "the newer cancelled invitation is untouched"},
	}

	for _, invitation := range invitations {
		seedMembershipInvitation(t, ctx, db, invitation.id, invitation.tenant, invitation.user, ownerID, invitation.status, invitation.created)
	}

	require.NoError(t, provider.ApplyNext(ctx), "056 must apply cleanly over duplicated pending invitations")

	for _, invitation := range invitations {
		assert.Equal(t, invitation.expected, invitationStatus(t, ctx, db, invitation.id), invitation.reason)
	}

	_, err = db.ExecContext(ctx, "UPDATE membership_invitations SET status = 'pending' WHERE id = ?", invitationID("a"))
	require.Error(t, err, "a second pending invitation for the same user and namespace is refused")

	require.NoError(t, provider.Rollback(ctx), "056 must roll back cleanly")

	execSQL(t, ctx, db, "UPDATE membership_invitations SET status = 'pending' WHERE id = ?", invitationID("a"))
}

func invitationID(digit string) string {
	return "66666666-6666-4666-8666-66666666666" + digit
}

func seedMembershipInvitation(t *testing.T, ctx context.Context, db *bun.DB, id, tenantID, userID, invitedBy, status string, createdAt time.Time) {
	t.Helper()

	execSQL(t, ctx, db, `INSERT INTO membership_invitations
		(id, tenant_id, user_id, invited_by, role, status, status_updated_at, expires_at, invitations, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'observer', ?, ?, ?, 1, ?, ?)`,
		id, tenantID, userID, invitedBy, status, createdAt, createdAt.Add(7*24*time.Hour), createdAt, createdAt)
}

func invitationStatus(t *testing.T, ctx context.Context, db *bun.DB, id string) string {
	t.Helper()

	var status string
	require.NoError(t, db.NewRaw("SELECT status FROM membership_invitations WHERE id = ?", id).Scan(ctx, &status))

	return status
}
