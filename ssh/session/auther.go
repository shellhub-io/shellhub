package session

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/Masterminds/semver"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/banner"
	"github.com/shellhub-io/shellhub/ssh/pkg/magickey"
	gossh "golang.org/x/crypto/ssh"
)

type authFunc func(*Session, *gossh.ClientConfig) error

type authMethod int8

const (
	AuthMethodPublicKey authMethod = iota // AuthMethodPublicKey represents a public key authentication
	AuthMethodPassword                    // AuthMethodPassword represents a password authentication
	AuthMethodEnroll                      // AuthMethodEnroll represents a browser-approved (credential-less) authentication
)

// mintEphemeralSigner asks the API for a fresh server-side keypair and sets it as
// the client credential toward the agent. The agent trusts it because the API
// holds the matching private key and vouches for it; the user's own credential
// is never forwarded. Shared by the public-key and browser-approval auth methods.
func mintEphemeralSigner(session *Session, config *gossh.ClientConfig) error {
	privateKey, err := session.api.CreatePrivateKey(context.TODO())
	if err != nil {
		return err
	}

	block, _ := pem.Decode(privateKey.Data)

	parsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	signer, err := gossh.NewSignerFromKey(parsed)
	if err != nil {
		return err
	}

	config.Auth = []gossh.AuthMethod{
		gossh.PublicKeys(signer),
	}

	return nil
}

// Auth interface defines a common interface for authenticating a session. An 'Auth'
// must have an associated [authMethod], an [authFunc] to authenticate the session, and
// an 'Evaluate' method to evaluate the session's context if necessary (e.g. the agent
// version when authenticating with public keys).
type Auth interface {
	// Method returns the associated authentication method.
	Method() authMethod

	// Auth defines the callback that must be called when authenticating the session.
	Auth() authFunc

	// Evaluate evaluates the session's context, returning an error if there's something
	// possibly broken. It's not always necessary.
	Evaluate(*Session) error
}

type publicKeyAuth struct {
	pk gliderssh.PublicKey
}

func AuthPublicKey(pk gliderssh.PublicKey) Auth {
	return &publicKeyAuth{pk: pk}
}

func (*publicKeyAuth) Method() authMethod {
	return AuthMethodPublicKey
}

func (*publicKeyAuth) Auth() authFunc {
	return mintEphemeralSigner
}

func (p *publicKeyAuth) Evaluate(session *Session) error {
	// Versions earlier than 0.6.0 do not validate the user when receiving a public key
	// authentication request. This implies that requests with invalid users are
	// treated as "authenticated" because the connection does not raise any error.
	// Moreover, the agent panics after the connection ends. To avoid this, connections
	// with public key are not permitted when agent version is 0.5.x or earlier
	if !sshconf.AllowPublickeyAccessBelow060 {
		version := session.Device.Info.Version
		if version != "latest" {
			semverVersion, err := semver.NewVersion(version)
			if err != nil {
				return ErrInvalidVersion
			}

			if semverVersion.LessThan(semver.MustParse("0.6.0")) {
				return ErrUnsuportedPublicKeyAuth
			}
		}
	}

	fingerprint := gossh.FingerprintLegacyMD5(p.pk)

	magic, err := gossh.NewPublicKey(&magickey.GetReference().PublicKey)
	if err != nil {
		return err
	}

	if gossh.FingerprintLegacyMD5(magic) != fingerprint {
		if _, err = session.api.GetPublicKey(context.TODO(), fingerprint, session.Device.TenantID); err != nil {
			return err
		}

		if ok, err := session.api.EvaluateKey(context.TODO(), fingerprint, session.Device, session.Data.Target.Username); !ok || err != nil {
			return ErrEvaluatePublicKey
		}
	}

	return err
}

type passwordAuth struct {
	pwd string
}

func AuthPassword(pwd string) Auth {
	return &passwordAuth{pwd: pwd}
}

func (*passwordAuth) Method() authMethod {
	return AuthMethodPassword
}

func (p *passwordAuth) Auth() authFunc {
	return func(_ *Session, config *gossh.ClientConfig) error {
		config.Auth = []gossh.AuthMethod{
			gossh.Password(p.pwd),
		}

		return nil
	}
}

func (*passwordAuth) Evaluate(*Session) error {
	// We don't need (yet) to do any evaluation when authenticating with password.
	return nil
}

// approvalWaitTimeout bounds how long the gateway holds the SSH handshake open
// waiting for a browser approval. It must be at least the API's approval TTL so a
// user has the full window to approve; the library sets no auth-phase deadline,
// so this is the only thing that releases an abandoned login.
const approvalWaitTimeout = 90 * time.Second

// approvalPollInterval is how often the gateway polls the API for the decision.
const approvalPollInterval = 3 * time.Second

// awaitEnrollment blocks polling the API for the approval decision, returning the
// approving account id once approved. The gliderssh context is cancelled when
// the client disconnects, so an abandoned login is released at once; otherwise
// the wait is bounded by approvalWaitTimeout.
func (s *Session) awaitEnrollment(gctx gliderssh.Context) (string, error) {
	ctx, cancel := context.WithTimeout(gctx, approvalWaitTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return "", ErrEnrollmentTimeout
		case <-time.After(approvalPollInterval):
		}

		status, err := s.api.GetSSHEnrollmentStatus(ctx, s.EnrollmentCode)
		if err != nil {
			// Transient API error or the code just expired; keep polling until
			// the context bounds it rather than aborting a valid approval.
			continue
		}

		switch status.State {
		case models.SSHEnrollmentConfirmed:
			return status.UserID, nil
		case models.SSHEnrollmentRejected:
			return "", ErrEnrollmentRejected
		}
	}
}

// authorize is the hard identity-mode gate: it refuses here, before any key is
// minted, so the agent is never contacted for a login the policies deny. It
// returns the decision so the caller can honor a step-up requirement.
func (s *Session) authorize(ctx context.Context) (*models.Decision, error) {
	dec, err := s.api.AuthorizeSSHAccess(ctx, s.Namespace.TenantID, s.UserID, s.Device, s.Target.Username)
	if err != nil || dec == nil || !dec.Allowed {
		return nil, ErrAccessDenied
	}

	return dec, nil
}

// enrollmentAuth authenticates a session whose presented key is not yet enrolled
// (identity mode): it blocks until a member approves the enrollment in the
// console — which binds the key to the approving account — then authorizes the
// now-known identity and reaches the agent with the server-minted ephemeral key.
type enrollmentAuth struct {
	ctx gliderssh.Context
}

func AuthEnroll(ctx gliderssh.Context) Auth {
	return &enrollmentAuth{ctx: ctx}
}

func (*enrollmentAuth) Method() authMethod {
	return AuthMethodEnroll
}

func (*enrollmentAuth) Auth() authFunc {
	return mintEphemeralSigner
}

func (a *enrollmentAuth) Evaluate(session *Session) error {
	approver, err := session.awaitEnrollment(a.ctx)
	if err != nil {
		return err
	}

	// The accept enrolled the key to the approver; bind that identity, then let
	// Access Policies decide whether it may reach this device as this login.
	session.UserID = approver

	_, err = session.authorize(a.ctx)

	return err
}

// identityAuth authenticates a session whose presented key is already enrolled
// (identity mode): the key is the identity, so it skips the browser step,
// authorizes the resolved account, runs an optional per-policy step-up, and
// reaches the agent with the server-minted ephemeral key.
type identityAuth struct {
	ctx gliderssh.Context
}

func AuthIdentity(ctx gliderssh.Context) Auth {
	return &identityAuth{ctx: ctx}
}

func (*identityAuth) Method() authMethod {
	return AuthMethodEnroll
}

func (*identityAuth) Auth() authFunc {
	return mintEphemeralSigner
}

func (a *identityAuth) Evaluate(session *Session) error {
	dec, err := session.authorize(a.ctx)
	if err != nil {
		return err
	}

	if dec.RequireStepUp {
		// Step-up: the identity is already established by the key; require a
		// per-session browser confirmation (enroll=false, so the confirm only
		// grants and does not re-bind) before proceeding.
		if err := session.api.AttachSSHEnrollmentKey(a.ctx, session.EnrollmentCode, session.Fingerprint, session.KeyData, false); err != nil {
			return err
		}

		sendBanner(a.ctx, buildStepUpBanner(sshconf.Domain, sshconf.AutoSSL, session.EnrollmentCode))

		if _, err := session.awaitEnrollment(a.ctx); err != nil {
			return err
		}
	}

	return nil
}

// preAuthConnCtxKey keys the pre-auth connection stashed on the gliderssh
// context by the gateway's PreAuthConnCallback.
type preAuthConnCtxKey struct{}

// StorePreAuthConn stashes the x/crypto pre-auth connection so a banner can be
// sent mid-handshake, after the presented key is known. The gateway calls it
// from PreAuthConnCallback.
func StorePreAuthConn(ctx gliderssh.Context, conn gossh.ServerPreAuthConn) {
	ctx.SetValue(preAuthConnCtxKey{}, conn)
}

// sendBanner delivers a mid-handshake SSH banner to the client. It is how the
// enrollment URL reaches a stock OpenSSH client, sent only once the gateway
// knows a browser step is actually needed, so an enrolled key never sees it. A
// no-op when the connection is absent (e.g. tests).
func sendBanner(ctx gliderssh.Context, msg string) {
	if conn, ok := ctx.Value(preAuthConnCtxKey{}).(gossh.ServerPreAuthConn); ok && conn != nil {
		_ = conn.SendAuthBanner(msg) //nolint:errcheck
	}
}

// webIdentityAuth authenticates a web-terminal session in identity mode. The
// browser user is already authenticated in the console (their id is bound to the
// session), so there is no key to resolve and no enrollment step: it just
// authorizes the identity against Access Policies and reaches the agent with the
// server-minted ephemeral key. No device credential is involved.
type webIdentityAuth struct {
	ctx gliderssh.Context
}

func AuthWebIdentity(ctx gliderssh.Context) Auth {
	return &webIdentityAuth{ctx: ctx}
}

func (*webIdentityAuth) Method() authMethod {
	return AuthMethodEnroll
}

func (*webIdentityAuth) Auth() authFunc {
	return mintEphemeralSigner
}

func (a *webIdentityAuth) Evaluate(session *Session) error {
	if _, err := session.authorize(a.ctx); err != nil {
		// Surface the denial as an access-denied banner so the web bridge maps it
		// to a clear "Access denied" in the console. Without it, a policy denial
		// reaches the browser as a generic auth failure ("username or password is
		// incorrect"), which is misleading in identity mode.
		sendBanner(a.ctx, banner.Message(banner.KindAccessDenied))

		return err
	}

	return nil
}

// ResolveKeyAuth resolves the presented key to a ShellHub identity (identity
// mode) and returns the auth to run: a hit yields the identity auth (authorize +
// mint, no browser); a miss arranges enrollment (attaching the key to the
// pending code, surfacing the enrollment URL to the client, and having the
// console page bind it) and yields the enrollment auth (poll + mint).
func (s *Session) ResolveKeyAuth(ctx gliderssh.Context, publicKey gliderssh.PublicKey) (Auth, error) {
	s.Fingerprint = gossh.FingerprintSHA256(publicKey)
	s.KeyData = gossh.MarshalAuthorizedKey(publicKey)

	resolution, err := s.api.ResolveSSHIdentity(ctx, s.Namespace.TenantID, s.Fingerprint)
	if err != nil {
		return nil, err
	}

	if resolution.Found {
		s.UserID = resolution.UserID

		return AuthIdentity(ctx), nil
	}

	if err := s.api.AttachSSHEnrollmentKey(ctx, s.EnrollmentCode, s.Fingerprint, s.KeyData, true); err != nil {
		return nil, err
	}

	// Now that the key is known to be unenrolled, surface the enrollment URL and
	// hold here. An enrolled key returns above and never sees this.
	sendBanner(ctx, buildEnrollmentBanner(sshconf.Domain, sshconf.AutoSSL, s.EnrollmentCode))

	return AuthEnroll(ctx), nil
}

var (
	// ErrEnrollmentRejected is returned when the user rejects the login in the console.
	ErrEnrollmentRejected = errors.New("ssh login denied")
	// ErrAccessDenied is returned when no Access Policy grants the approved
	// identity access to the target device as the requested login.
	ErrAccessDenied = errors.New("ssh access denied by policy")
	// ErrEnrollmentTimeout is returned when no decision arrives before the wait
	// deadline or the client disconnects.
	ErrEnrollmentTimeout = errors.New("ssh login approval timed out")
)
