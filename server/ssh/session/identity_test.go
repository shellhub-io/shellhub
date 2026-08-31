package session

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"sync"
	"testing"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/services"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/target"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func newIdentitySession(service services.Service, mode string) *Session {
	tgt, _ := target.NewTarget("user@namespace.device")

	return &Session{
		UID:     "test-uid",
		service: service,
		Data: Data{
			Target:       tgt,
			IPAddress:    "127.0.0.1",
			SSHID:        "user@namespace.device",
			ApprovalCode: "WXYZ2K7Q",
			Device:       &models.Device{UID: "device-uid", Name: "device", TenantID: "tenant-id"},
			Namespace: &models.Namespace{
				Name:     "namespace",
				TenantID: "tenant-id",
				Settings: &models.NamespaceSettings{SSHAccessMode: mode},
			},
		},
		once:   new(sync.Once),
		seats:  newSeats(),
		agent:  &Agent{channels: make(map[int]*AgentChannel)},
		client: &Client{channels: make(map[int]*ClientChannel)},
	}
}

func newTestSSHKey(t *testing.T) gliderssh.PublicKey {
	t.Helper()

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(priv)
	require.NoError(t, err)

	return signer.PublicKey()
}

func TestIsIdentityMode(t *testing.T) {
	require.True(t, newIdentitySession(nil, models.SSHAccessModeIdentity).IsIdentityMode())
	require.False(t, newIdentitySession(nil, models.SSHAccessModeLegacy).IsIdentityMode())
}

func TestResolveKeyAuth(t *testing.T) {
	pubKey := newTestSSHKey(t)
	fingerprint := gossh.FingerprintSHA256(pubKey)

	t.Run("enrolled active key yields the identity auth and binds the account", func(t *testing.T) {
		serviceMock := servicemocks.NewMockService(t)
		serviceMock.EXPECT().
			ResolveSSHIdentity(mock.Anything, "tenant-id", fingerprint).
			Return(&models.SSHIdentity{PrincipalID: "user1", SingleUse: true}, true, nil). //nolint:exhaustruct
			Once()

		sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)

		auth, err := sess.ResolveKeyAuth(newStubContext(), pubKey)
		require.NoError(t, err)
		assert.IsType(t, &identityAuth{}, auth)
		assert.Equal(t, "user1", sess.UserID)
		assert.Equal(t, fingerprint, sess.Fingerprint)
		assert.True(t, sess.SingleUse)
	})

	t.Run("dead key is rejected, not sent to enrollment", func(t *testing.T) {
		consumedAt := clock.Now()

		serviceMock := servicemocks.NewMockService(t)
		serviceMock.EXPECT().
			ResolveSSHIdentity(mock.Anything, "tenant-id", fingerprint).
			Return(&models.SSHIdentity{PrincipalID: "user1", ConsumedAt: &consumedAt}, true, nil). //nolint:exhaustruct
			Once()

		sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)

		auth, err := sess.ResolveKeyAuth(newStubContext(), pubKey)
		require.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, auth)
		assert.Empty(t, sess.UserID)
	})

	t.Run("an expired key is rejected too", func(t *testing.T) {
		expiredAt := clock.Now().Add(-time.Hour)

		serviceMock := servicemocks.NewMockService(t)
		serviceMock.EXPECT().
			ResolveSSHIdentity(mock.Anything, "tenant-id", fingerprint).
			Return(&models.SSHIdentity{PrincipalID: "user1", ExpiresAt: &expiredAt}, true, nil). //nolint:exhaustruct
			Once()

		sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)

		auth, err := sess.ResolveKeyAuth(newStubContext(), pubKey)
		require.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, auth)
		assert.Empty(t, sess.UserID)
	})

	t.Run("unknown key yields the approval auth without writing anything", func(t *testing.T) {
		serviceMock := servicemocks.NewMockService(t)
		serviceMock.EXPECT().
			ResolveSSHIdentity(mock.Anything, "tenant-id", fingerprint).
			Return(nil, false, nil).
			Once()

		sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)

		auth, err := sess.ResolveKeyAuth(newStubContext(), pubKey)
		require.NoError(t, err)
		assert.IsType(t, &approvalAuth{}, auth)
		assert.Empty(t, sess.UserID)

		assert.Equal(t, "WXYZ2K7Q", sess.ApprovalCode)
	})

	t.Run("offering an unknown key writes nothing", func(t *testing.T) {
		serviceMock := servicemocks.NewMockService(t)
		serviceMock.EXPECT().
			ResolveSSHIdentity(mock.Anything, "tenant-id", fingerprint).
			Return(nil, false, nil).
			Once()

		sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)

		auth, err := sess.ResolveKeyAuth(newStubContext(), pubKey)
		require.NoError(t, err)
		require.NoError(t, auth.Offer(sess))

		assert.Equal(t, "WXYZ2K7Q", sess.ApprovalCode)
	})

	t.Run("the approval is parked once the key is proven", func(t *testing.T) {
		serviceMock := servicemocks.NewMockService(t)
		serviceMock.EXPECT().
			ResolveSSHIdentity(mock.Anything, "tenant-id", fingerprint).
			Return(nil, false, nil).
			Once()
		serviceMock.EXPECT().
			CreateSSHApproval(mock.Anything, mock.MatchedBy(func(req *requests.SSHApprovalCreate) bool {
				return req.Fingerprint == fingerprint &&
					req.Kind == models.SSHApprovalIdentity &&
					req.TenantID == "tenant-id"
			})).
			Return(&models.SSHApprovalCreated{Code: "AB12CD34"}, nil).
			Once()
		serviceMock.EXPECT().
			GetSSHApprovalStatus(mock.Anything, mock.Anything).
			Return(&models.SSHApprovalStatus{State: models.SSHApprovalRejected}, nil). //nolint:exhaustruct
			Once()

		sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)

		auth, err := sess.ResolveKeyAuth(newStubContext(), pubKey)
		require.NoError(t, err)

		require.ErrorIs(t, auth.Evaluate(sess), ErrApprovalRejected)
		assert.Equal(t, "AB12CD34", sess.ApprovalCode)
	})
}

// captureLogs collects what the session logs on the standard logger, at debug
// level so entries below the default threshold are recorded too.
func captureLogs(t *testing.T) *test.Hook {
	t.Helper()

	hook := test.NewGlobal()
	level := log.GetLevel()
	log.SetLevel(log.DebugLevel)

	t.Cleanup(func() {
		log.SetLevel(level)
		hook.Reset()
	})

	return hook
}

func TestAuthorize(t *testing.T) {
	ctx := context.Background()

	t.Run("a policy denial is reported with the reason the policies gave", func(t *testing.T) {
		hook := captureLogs(t)

		serviceMock := servicemocks.NewMockService(t)
		serviceMock.EXPECT().
			Authorize(mock.Anything, "tenant-id", "user-id", "device-uid", "user", "127.0.0.1").
			Return(&models.Decision{Allowed: false, Reason: models.ReasonNoGrant, Login: "user"}, nil).
			Once()

		sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)
		sess.UserID = "user-id"

		dec, err := sess.authorize(ctx)

		require.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, dec)

		require.Len(t, hook.AllEntries(), 1)

		entry := hook.LastEntry()
		assert.Equal(t, log.WarnLevel, entry.Level)
		assert.Equal(t, "ssh access denied by the access policies", entry.Message)
		assert.Equal(t, models.ReasonNoGrant, entry.Data["reason"])
		assert.Equal(t, `no policy grants "user" on this device`, entry.Data["detail"])
		assert.NotContains(t, entry.Data, "policy_name")
		assert.Equal(t, "tenant-id", entry.Data["tenant"])
		assert.Equal(t, "user-id", entry.Data["user"])
		assert.Equal(t, "device-uid", entry.Data["device"])
		assert.Equal(t, "user", entry.Data["username"])
		assert.Equal(t, "127.0.0.1", entry.Data["ip"])
		assert.Equal(t, "test-uid", entry.Data["session"])
	})

	t.Run("a denial by a named policy carries the policy on the entry", func(t *testing.T) {
		hook := captureLogs(t)

		serviceMock := servicemocks.NewMockService(t)
		serviceMock.EXPECT().
			Authorize(mock.Anything, "tenant-id", "user-id", "device-uid", "user", "127.0.0.1").
			Return(&models.Decision{Allowed: false, Reason: models.ReasonDeniedByPolicy, PolicyName: "block contractors"}, nil).
			Once()

		sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)
		sess.UserID = "user-id"

		_, err := sess.authorize(ctx)
		require.ErrorIs(t, err, ErrAccessDenied)

		require.Len(t, hook.AllEntries(), 1)

		entry := hook.LastEntry()
		assert.Equal(t, "ssh access denied by the access policies", entry.Message)
		assert.Equal(t, models.ReasonDeniedByPolicy, entry.Data["reason"])
		assert.Equal(t, "block contractors", entry.Data["policy_name"])
		assert.Equal(t, `denied by policy "block contractors"`, entry.Data["detail"])
	})

	t.Run("a failure to evaluate the policies is reported as a malfunction", func(t *testing.T) {
		hook := captureLogs(t)

		serviceMock := servicemocks.NewMockService(t)
		serviceMock.EXPECT().
			Authorize(mock.Anything, "tenant-id", "user-id", "device-uid", "user", "127.0.0.1").
			Return(nil, errors.New("the store is unreachable")).
			Once()

		sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)
		sess.UserID = "user-id"

		dec, err := sess.authorize(ctx)

		require.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, dec)

		require.Len(t, hook.AllEntries(), 1)

		entry := hook.LastEntry()
		assert.Equal(t, log.ErrorLevel, entry.Level)
		loggedErr, ok := entry.Data[log.ErrorKey].(error)
		require.True(t, ok)
		require.EqualError(t, loggedErr, "the store is unreachable")
		assert.NotContains(t, entry.Data, "reason")
	})

	t.Run("an absent decision is reported as a malfunction", func(t *testing.T) {
		hook := captureLogs(t)

		serviceMock := servicemocks.NewMockService(t)
		serviceMock.EXPECT().
			Authorize(mock.Anything, "tenant-id", "user-id", "device-uid", "user", "127.0.0.1").
			Return(nil, nil).
			Once()

		sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)
		sess.UserID = "user-id"

		dec, err := sess.authorize(ctx)

		require.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, dec)

		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, log.ErrorLevel, hook.LastEntry().Level)
	})

	t.Run("an allowed decision is returned without a log line", func(t *testing.T) {
		hook := captureLogs(t)

		serviceMock := servicemocks.NewMockService(t)
		serviceMock.EXPECT().
			Authorize(mock.Anything, "tenant-id", "user-id", "device-uid", "user", "127.0.0.1").
			Return(&models.Decision{Allowed: true, RequireReauth: true}, nil).
			Once()

		sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)
		sess.UserID = "user-id"

		dec, err := sess.authorize(ctx)

		require.NoError(t, err)
		require.NotNil(t, dec)
		assert.True(t, dec.RequireReauth)
		assert.Empty(t, hook.AllEntries())
	})
}
