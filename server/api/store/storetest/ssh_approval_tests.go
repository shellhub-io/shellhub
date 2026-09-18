package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Suite) newPendingApproval(t *testing.T, tenantID, code string, expiresAt time.Time) {
	t.Helper()

	require.NoError(t, s.provider.Store().SSHApprovalCreate(context.Background(), &models.SSHApproval{
		Code:        code,
		TenantID:    tenantID,
		Kind:        models.SSHApprovalIdentity,
		SessionUID:  "session-" + code,
		SSHID:       "user@device." + tenantID,
		DeviceUID:   "device-" + code,
		DeviceName:  "device",
		Username:    "root",
		IPAddress:   "10.0.0.1",
		Fingerprint: "SHA256:" + code,
		Data:        []byte("ssh-ed25519 AAAA"),
		State:       models.SSHApprovalPending,
		RequestedAt: expiresAt.Add(-time.Minute),
		ExpiresAt:   expiresAt,
	}))
}

// TestSSHApprovalExpiryHidesTheRow verifies an approval past its expiry is reported as not found
// while the row survives, so a parked login fails closed before the cron prunes it.
func (s *Suite) TestSSHApprovalExpiryHidesTheRow(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	now := clock.Now().UTC().Truncate(time.Second)

	s.newPendingApproval(t, tenantID, "AAAA1111", now.Add(time.Minute))

	approval, err := st.SSHApprovalGet(ctx, "AAAA1111", now)
	require.NoError(t, err)
	require.NotNil(t, approval)
	assert.Equal(t, models.SSHApprovalPending, approval.State)
	assert.Equal(t, models.SSHApprovalIdentity, approval.Kind)

	_, err = st.SSHApprovalGet(ctx, "AAAA1111", now.Add(2*time.Minute))
	assert.Error(t, err)
}

// TestSSHApprovalDecideIsClaimedOnce verifies the move out of pending is the single-use claim.
func (s *Suite) TestSSHApprovalDecideIsClaimedOnce(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	userID := s.CreateUser(t)
	now := clock.Now().UTC().Truncate(time.Second)

	s.newPendingApproval(t, tenantID, "BBBB2222", now.Add(time.Minute))

	claimed, err := st.SSHApprovalDecide(ctx, "BBBB2222", models.SSHApprovalConfirmed, userID, "CONF1234", now)
	require.NoError(t, err)
	assert.True(t, claimed)

	claimedAgain, err := st.SSHApprovalDecide(ctx, "BBBB2222", models.SSHApprovalRejected, userID, "", now)
	require.NoError(t, err)
	assert.False(t, claimedAgain, "a second submit must not overwrite the decision")

	approval, err := st.SSHApprovalGet(ctx, "BBBB2222", now)
	require.NoError(t, err)
	assert.Equal(t, models.SSHApprovalConfirmed, approval.State)
	assert.Equal(t, userID, approval.DecidedBy)
	assert.Equal(t, "CONF1234", approval.ConfirmationCode,
		"the code the person types at their terminal has to survive the round trip, and this is the only seam with a real database")
	assert.NotEqual(t, approval.DecidedBy, approval.ConfirmationCode,
		"decided_by and confirmation_code are adjacent same-typed arguments, so a transposition has to fail here")
}

// TestSSHApprovalCleanupRemovesOnlyTheExpired verifies the cron prunes by expiry alone.
func (s *Suite) TestSSHApprovalCleanupRemovesOnlyTheExpired(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	now := clock.Now().UTC().Truncate(time.Second)

	s.newPendingApproval(t, tenantID, "CCCC3333", now.Add(-time.Hour))
	s.newPendingApproval(t, tenantID, "DDDD4444", now.Add(time.Hour))

	removed, err := st.SSHApprovalCleanup(ctx, now)
	require.NoError(t, err)
	assert.Equal(t, int64(1), removed)

	alive, err := st.SSHApprovalGet(ctx, "DDDD4444", now)
	require.NoError(t, err)
	assert.Equal(t, "DDDD4444", alive.Code)
}
