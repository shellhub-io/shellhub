package session

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer/dialertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type grantFixture struct {
	service *servicemocks.MockService
	tunnel  *dialertest.Stub
	sess    *Session
	ctx     *stubContext
}

func newGrantFixture(t *testing.T) *grantFixture {
	t.Helper()

	Configure(Config{ConnectTimeout: 0}) //nolint:exhaustruct

	gatewayKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	service := servicemocks.NewMockService(t)
	service.EXPECT().
		CreatePrivateKey(mock.Anything).
		Return(&models.PrivateKey{Data: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(gatewayKey)})}, nil). //nolint:exhaustruct // only the PEM is read
		Maybe()

	tunnel := &dialertest.Stub{Err: dialer.ErrNoConnection} //nolint:exhaustruct

	sess := newIdentitySession(service, false)
	sess.dialer = tunnel
	sess.Device.PublicKey = dialertest.NewAgent(t).PublicKey()
	sess.Events = NewEvents(sess.UID, tenantScope, service)

	ctx := newStubContext()
	advance(ctx, sess, StateEvaluated)

	return &grantFixture{service: service, tunnel: tunnel, sess: sess, ctx: ctx}
}

func (f *grantFixture) enroll(key gliderssh.PublicKey, identity *models.SSHIdentity) {
	f.service.EXPECT().
		ResolveSSHIdentity(mock.Anything, "tenant-id", gossh.FingerprintSHA256(key)).
		Return(identity, identity != nil, nil)
}

func (f *grantFixture) allow(principal models.Principal) {
	f.service.EXPECT().
		Authorize(mock.Anything, "tenant-id", principal, "device-uid", "user", "127.0.0.1").
		Return(&models.Decision{Allowed: true}, nil) //nolint:exhaustruct // only Allowed is read
}

func (f *grantFixture) login(t *testing.T, key gliderssh.PublicKey) error {
	t.Helper()

	auth, err := f.sess.ResolveKeyAuth(f.ctx, key)
	if err != nil {
		return err
	}

	return f.sess.Auth(f.ctx, auth)
}

func (f *grantFixture) state() State {
	_, state := ObtainSession(f.ctx)

	return state
}

func apiKeyPrincipal(id string) models.Principal {
	return models.Principal{Kind: models.PrincipalAPIKey, ID: id}
}

func userPrincipal(id string) models.Principal {
	return models.Principal{Kind: models.PrincipalUser, ID: id}
}

func TestSingleUseIdentityRedemption(t *testing.T) {
	t.Run("a login that loses the redemption never registers", func(t *testing.T) {
		f := newGrantFixture(t)
		key := newTestSSHKey(t)

		f.enroll(key, &models.SSHIdentity{PrincipalID: "key-id", PrincipalType: models.PrincipalAPIKey, SingleUse: true}) //nolint:exhaustruct
		f.allow(apiKeyPrincipal("key-id"))
		f.service.EXPECT().
			ConsumeSSHIdentity(mock.Anything, "tenant-id", gossh.FingerprintSHA256(key)).
			Return(false, nil).
			Once()

		require.ErrorIs(t, f.login(t, key), ErrAccessDenied)

		assert.Equal(t, StateEvaluated, f.state())
		assert.False(t, f.sess.registered)
		assert.Empty(t, f.tunnel.Dials())
	})

	t.Run("the identity is redeemed even when the device cannot be reached", func(t *testing.T) {
		f := newGrantFixture(t)
		key := newTestSSHKey(t)

		f.enroll(key, &models.SSHIdentity{PrincipalID: "key-id", PrincipalType: models.PrincipalAPIKey, SingleUse: true}) //nolint:exhaustruct
		f.allow(apiKeyPrincipal("key-id"))
		f.service.EXPECT().
			ConsumeSSHIdentity(mock.Anything, "tenant-id", gossh.FingerprintSHA256(key)).
			Return(true, nil).
			Once()
		f.service.EXPECT().
			CreateSession(mock.Anything, mock.MatchedBy(func(req requests.SessionCreate) bool {
				return req.APIKeyID == "key-id" && req.UserID == ""
			})).
			Return(&models.Session{}, nil). //nolint:exhaustruct // the session is registered, not read back
			Once()

		require.Error(t, f.login(t, key))
		assert.Equal(t, StateRegistered, f.state())

		spare := newTestSSHKey(t)
		f.enroll(spare, nil)

		require.ErrorIs(t, f.login(t, spare), ErrInvalidSessionState)
		assert.Len(t, f.tunnel.Dials(), 1, "another key must not take the registered session to the device")
	})
}

func TestRegisteredSessionRejoin(t *testing.T) {
	t.Run("another key of the registered principal reaches the device again", func(t *testing.T) {
		f := newGrantFixture(t)
		first, second := newTestSSHKey(t), newTestSSHKey(t)

		f.enroll(first, &models.SSHIdentity{PrincipalID: "user1", PrincipalType: models.PrincipalUser})  //nolint:exhaustruct
		f.enroll(second, &models.SSHIdentity{PrincipalID: "user1", PrincipalType: models.PrincipalUser}) //nolint:exhaustruct
		f.allow(userPrincipal("user1"))
		f.service.EXPECT().
			CreateSession(mock.Anything, mock.MatchedBy(func(req requests.SessionCreate) bool { return req.UserID == "user1" })).
			Return(&models.Session{}, nil). //nolint:exhaustruct // the session is registered, not read back
			Once()

		require.Error(t, f.login(t, first))
		require.Equal(t, StateRegistered, f.state())

		err := f.login(t, second)
		require.Error(t, err)
		require.NotErrorIs(t, err, ErrInvalidSessionState)
		assert.Len(t, f.tunnel.Dials(), 2)
	})

	t.Run("a key of another principal is refused without reaching the device", func(t *testing.T) {
		f := newGrantFixture(t)
		first, other := newTestSSHKey(t), newTestSSHKey(t)

		f.enroll(first, &models.SSHIdentity{PrincipalID: "user1", PrincipalType: models.PrincipalUser}) //nolint:exhaustruct
		f.enroll(other, &models.SSHIdentity{PrincipalID: "user2", PrincipalType: models.PrincipalUser}) //nolint:exhaustruct
		f.allow(userPrincipal("user1"))
		f.allow(userPrincipal("user2"))
		f.service.EXPECT().
			CreateSession(mock.Anything, mock.MatchedBy(func(req requests.SessionCreate) bool { return req.UserID == "user1" })).
			Return(&models.Session{}, nil). //nolint:exhaustruct // the session is registered, not read back
			Once()

		require.Error(t, f.login(t, first))
		require.Equal(t, StateRegistered, f.state())

		require.ErrorIs(t, f.login(t, other), ErrInvalidSessionState)
		assert.Len(t, f.tunnel.Dials(), 1)
	})

	t.Run("a key that needs an approval is refused without opening one", func(t *testing.T) {
		f := newGrantFixture(t)
		first, unenrolled := newTestSSHKey(t), newTestSSHKey(t)

		f.enroll(first, &models.SSHIdentity{PrincipalID: "user1", PrincipalType: models.PrincipalUser}) //nolint:exhaustruct
		f.enroll(unenrolled, nil)
		f.allow(userPrincipal("user1"))
		f.service.EXPECT().
			CreateSession(mock.Anything, mock.Anything).
			Return(&models.Session{}, nil). //nolint:exhaustruct // the session is registered, not read back
			Once()

		require.Error(t, f.login(t, first))
		require.Equal(t, StateRegistered, f.state())

		require.ErrorIs(t, f.login(t, unenrolled), ErrInvalidSessionState)
		assert.Len(t, f.tunnel.Dials(), 1)
	})
}
