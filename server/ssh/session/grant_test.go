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
		assert.Len(t, f.tunnel.Dials(), 1)
	})
}
