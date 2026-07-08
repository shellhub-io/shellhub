package session

import (
	"crypto/ed25519"
	"crypto/rand"
	"sync"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
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
			Target:         tgt,
			IPAddress:      "127.0.0.1",
			SSHID:          "user@namespace.device",
			EnrollmentCode: "WXYZ2K7Q",
			Device:         &models.Device{UID: "device-uid", Name: "device", TenantID: "tenant-id"},
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

	t.Run("enrolled key yields the identity auth and binds the account", func(t *testing.T) {
		apiMock := mocks.NewMockClient(t)
		apiMock.EXPECT().
			ResolveSSHIdentity(mock.Anything, "tenant-id", fingerprint).
			Return(&internalclient.SSHIdentityResolution{Found: true, UserID: "user1"}, nil).
			Once()

		sess := newIdentitySession(apiMock, models.SSHAccessModeIdentity)

		auth, err := sess.ResolveKeyAuth(newStubContext(), pubKey)
		require.NoError(t, err)
		assert.IsType(t, &identityAuth{}, auth)
		assert.Equal(t, "user1", sess.UserID)
		assert.Equal(t, fingerprint, sess.Fingerprint)
	})

	t.Run("unknown key attaches enrollment and yields the approval auth", func(t *testing.T) {
		apiMock := mocks.NewMockClient(t)
		apiMock.EXPECT().
			ResolveSSHIdentity(mock.Anything, "tenant-id", fingerprint).
			Return(&internalclient.SSHIdentityResolution{Found: false}, nil).
			Once()
		apiMock.EXPECT().
			AttachSSHEnrollmentKey(mock.Anything, "WXYZ2K7Q", fingerprint, mock.Anything, true).
			Return(nil).
			Once()

		sess := newIdentitySession(apiMock, models.SSHAccessModeIdentity)

		auth, err := sess.ResolveKeyAuth(newStubContext(), pubKey)
		require.NoError(t, err)
		assert.IsType(t, &enrollmentAuth{}, auth)
		assert.Empty(t, sess.UserID)
	})
}
