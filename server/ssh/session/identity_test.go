package session

import (
	"crypto/ed25519"
	"crypto/rand"
	"sync"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func newIdentitySession(apiClient internalclient.Client, mode string) *Session {
	tgt, _ := target.NewTarget("user@namespace.device") //nolint:errcheck

	return &Session{
		UID: "test-uid",
		api: apiClient,
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
		Seats:  NewSeats(),
		Agent:  &Agent{Channels: make(map[int]*AgentChannel)},
		Client: &Client{Channels: make(map[int]*ClientChannel)},
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
		apiMock := mocks.NewMockClient(t)
		apiMock.EXPECT().
			ResolveSSHIdentity(mock.Anything, "tenant-id", fingerprint).
			Return(&internalclient.SSHIdentityResolution{Found: true, Active: true, UserID: "user1", SingleUse: true}, nil).
			Once()

		sess := newIdentitySession(apiMock, models.SSHAccessModeIdentity)

		auth, err := sess.ResolveKeyAuth(newStubContext(), pubKey)
		require.NoError(t, err)
		assert.IsType(t, &identityAuth{}, auth)
		assert.Equal(t, "user1", sess.UserID)
		assert.Equal(t, fingerprint, sess.Fingerprint)
		assert.True(t, sess.SingleUse)
	})

	t.Run("dead key is rejected, not sent to enrollment", func(t *testing.T) {
		apiMock := mocks.NewMockClient(t)
		apiMock.EXPECT().
			ResolveSSHIdentity(mock.Anything, "tenant-id", fingerprint).
			Return(&internalclient.SSHIdentityResolution{Found: true, Active: false, UserID: "user1"}, nil).
			Once()

		sess := newIdentitySession(apiMock, models.SSHAccessModeIdentity)

		auth, err := sess.ResolveKeyAuth(newStubContext(), pubKey)
		require.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, auth)
		// A dead service-account key must never enroll or bind an account.
		assert.Empty(t, sess.UserID)
	})

	t.Run("unknown key opens an identity approval and yields the approval auth", func(t *testing.T) {
		apiMock := mocks.NewMockClient(t)
		apiMock.EXPECT().
			ResolveSSHIdentity(mock.Anything, "tenant-id", fingerprint).
			Return(&internalclient.SSHIdentityResolution{Found: false}, nil).
			Once()
		apiMock.EXPECT().
			CreateSSHApproval(mock.Anything, mock.MatchedBy(func(req requests.SSHApprovalCreate) bool {
				return req.Fingerprint == fingerprint &&
					req.Kind == models.SSHApprovalIdentity &&
					req.TenantID == "tenant-id"
			})).
			Return(&models.SSHApprovalCreated{Code: "AB12CD34"}, nil).
			Once()

		sess := newIdentitySession(apiMock, models.SSHAccessModeIdentity)

		auth, err := sess.ResolveKeyAuth(newStubContext(), pubKey)
		require.NoError(t, err)
		assert.IsType(t, &approvalAuth{}, auth)
		assert.Equal(t, "AB12CD34", sess.ApprovalCode)
		assert.Empty(t, sess.UserID)
	})
}
