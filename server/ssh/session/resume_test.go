package session

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer/dialertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

// TestResumeFinishesAParkedLoginWithoutEvaluatingAgain pins the split Auth and
// Resume share: a login parked on an approval is finished from the challenge
// callback through the same register, connect and authenticate tail a login
// that needed no approval takes, and the evaluation that parked it is not run
// a second time. The approver becomes the principal the access policy is asked
// about, which is how an enrollment names who is logging in. The mocks are
// strict, so a repeated lookup or approval fails the test on its own.
func TestResumeFinishesAParkedLoginWithoutEvaluatingAgain(t *testing.T) {
	Configure(Config{ConnectTimeout: 0}) //nolint:exhaustruct

	pubKey := newTestSSHKey(t)
	fingerprint := gossh.FingerprintSHA256(pubKey)

	gatewayKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	serviceMock := servicemocks.NewMockService(t)
	serviceMock.EXPECT().
		CreatePrivateKey(mock.Anything).
		Return(&models.PrivateKey{Data: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(gatewayKey)})}, nil). //nolint:exhaustruct // only the PEM is read
		Once()
	serviceMock.EXPECT().
		ResolveSSHIdentity(mock.Anything, "tenant-id", fingerprint).
		Return(nil, false, nil).
		Once()
	serviceMock.EXPECT().
		CreateSSHApproval(mock.Anything, mock.Anything).
		Return(&models.SSHApprovalCreated{Code: "AB12CD34"}, nil). //nolint:exhaustruct // only the code is read
		Once()
	serviceMock.EXPECT().
		Authorize(mock.Anything, "tenant-id", models.Principal{Kind: models.PrincipalUser, ID: "user-id"}, "device-uid", "user", "127.0.0.1").
		Return(&models.Decision{Allowed: true}, nil). //nolint:exhaustruct // only Allowed is read
		Once()
	serviceMock.EXPECT().
		CreateSession(mock.Anything, mock.Anything).
		Return(&models.Session{}, nil). //nolint:exhaustruct // the session is registered, not read back
		Once()
	serviceMock.EXPECT().
		UpdateSession(mock.Anything, models.UID("test-uid"), mock.Anything).
		Return(nil)

	sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)
	sess.dialer = dialertest.NewAgent(t)
	sess.Events = NewEvents(sess.UID, serviceMock)

	ctx := newStubContext()
	advance(ctx, sess, StateEvaluated)

	auth, err := sess.ResolveKeyAuth(ctx, pubKey)
	require.NoError(t, err)
	require.ErrorIs(t, sess.Auth(ctx, auth), ErrApprovalRequired, "the key needs approval, so Auth parks the login")

	_, state := ObtainSession(ctx)
	require.Equal(t, StateChallenged, state, "a parked login sits in StateChallenged until the client answers")

	require.NoError(t, sess.Resume(ctx, "user-id"))

	_, state = ObtainSession(ctx)
	assert.Equal(t, StateFinished, state, "Resume must carry the login through to the same end Auth reaches")
	assert.True(t, sess.registered, "Resume must register the session, or its teardown has nothing to deactivate")
}

// TestAuthRetriesAfterTheDialToTheDeviceFailed is the retry a client gets for
// free: one failed attempt must not cost the whole connection. finish advances
// to StateRegistered before it reaches the device, and nothing rolls that back
// when reaching it fails, so a second attempt arrives on a session already
// registered and has to re-enter at the join rather than be turned away as an
// invalid state. The dial is what fails here, which is an offline device; a
// credential the device rejects lands on the same state the same way.
func TestAuthRetriesAfterTheDialToTheDeviceFailed(t *testing.T) {
	Configure(Config{ConnectTimeout: 0}) //nolint:exhaustruct

	serviceMock := servicemocks.NewMockService(t)
	serviceMock.EXPECT().
		CreateSession(mock.Anything, mock.Anything).
		Return(&models.Session{}, nil). //nolint:exhaustruct // the session is registered, not read back
		Once()

	sess := newIdentitySession(serviceMock, models.SSHAccessModeIdentity)
	sess.dialer = &dialertest.Stub{Err: dialer.ErrNoConnection} //nolint:exhaustruct
	sess.Events = NewEvents(sess.UID, serviceMock)

	ctx := newStubContext()
	advance(ctx, sess, StateEvaluated)

	auth := AuthPassword("wrong")

	require.Error(t, sess.Auth(ctx, auth), "the device cannot be reached, so the first attempt fails")

	_, state := ObtainSession(ctx)
	require.Equal(t, StateRegistered, state, "the session registered before it reached the device")

	err := sess.Auth(ctx, auth)

	require.Error(t, err, "the device is still unreachable, so the retry fails too")
	require.NotErrorIs(t, err, ErrInvalidSessionState,
		"a retry on a registered session must reach the device again, not be refused by the state machine")
}
