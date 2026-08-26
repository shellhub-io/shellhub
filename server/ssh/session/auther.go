package session

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/Masterminds/semver/v3"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/banner"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/magickey"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type authFunc func(*Session, *gossh.ClientConfig) error

// mintEphemeralSigner asks the API for a fresh server-side keypair and sets it as
// the client credential toward the agent. The agent trusts it because the API
// holds the matching private key and vouches for it; the user's own credential
// is never forwarded. Shared by the public-key and browser-approval auth methods.
func mintEphemeralSigner(session *Session, config *gossh.ClientConfig) error {
	privateKey, err := session.service.CreatePrivateKey(context.Background())
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
// must have an [authFunc] to authenticate the session and an 'Evaluate' method to
// evaluate the session's context if necessary (e.g. the agent version when
// authenticating with public keys).
//
// The two checks are split by what the client has proven. The SSH protocol lets
// a client ask whether a key would be acceptable without signing anything, and
// x/crypto runs the publickey callback for that query, so anything reachable
// before a signature must stay cheap and free of side effects.
type Auth interface {
	// Auth defines the callback that must be called when authenticating the session.
	Auth() authFunc

	// Offer decides whether the credential is acceptable at all. It may run
	// before the client has proven it holds the key — including for a key it
	// does not hold — so it must only read.
	Offer(*Session) error

	// Evaluate runs once the credential is proven, and is where anything with a
	// cost or a side effect belongs. It's not always necessary.
	Evaluate(*Session) error
}

type publicKeyAuth struct {
	pk gliderssh.PublicKey
}

func AuthPublicKey(pk gliderssh.PublicKey) Auth {
	return &publicKeyAuth{pk: pk}
}

func (*publicKeyAuth) Auth() authFunc {
	return mintEphemeralSigner
}

func (*publicKeyAuth) Evaluate(*Session) error {
	// Everything legacy mode checks about a key is a read, so it all happens in
	// Offer and there is nothing left to do once the signature lands.
	return nil
}

func (p *publicKeyAuth) Offer(session *Session) error {
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
		ctx := context.Background()

		// The key is fetched once and evaluated twice. Over HTTP this was two
		// requests, the second fetching the key again on the other side.
		key, err := session.service.GetPublicKey(ctx, fingerprint, session.Device.TenantID)
		if err != nil {
			return err
		}

		usernameOK, err := session.service.EvaluateKeyUsername(ctx, key, session.Data.Target.Username)
		if err != nil {
			return ErrEvaluatePublicKey
		}

		filterOK, err := session.service.EvaluateKeyFilter(ctx, key, *session.Device)
		if err != nil {
			return ErrEvaluatePublicKey
		}

		if !usernameOK || !filterOK {
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

func (*passwordAuth) Offer(*Session) error {
	// A password is only ever sent proven; the protocol has no query form for it.
	return nil
}

// approvalWaitTimeout bounds how long the gateway holds the SSH handshake open
// waiting for a browser approval. It must be at least the API's approval TTL so a
// user has the full window to approve; the library sets no auth-phase deadline,
// so this is the only thing that releases an abandoned login.
const approvalWaitTimeout = 90 * time.Second

// ApprovalWaitTimeout is [approvalWaitTimeout] for callers that have to size a
// deadline around it — the SSH server's handshake timeout, above all, which
// would otherwise cut off a login the user is about to approve.
const ApprovalWaitTimeout = approvalWaitTimeout

// approvalAskInterval is the pause between two asks. The API holds each one open
// until the decision lands, so this is not what paces the wait — it only keeps a
// gateway talking to an API too old to honour that from spinning, and bounds how
// fast a failing API is retried.
const approvalAskInterval = 250 * time.Millisecond

// awaitApproval blocks until the login is decided, returning the approving
// account id once approved. Each ask is held open by the API, so the decision
// comes back as soon as it is written rather than on a poll boundary. The
// gliderssh context is cancelled when the client disconnects, so an abandoned
// login is released at once; otherwise the wait is bounded by
// approvalWaitTimeout.
func (s *Session) awaitApproval(gctx gliderssh.Context) (string, error) {
	ctx, cancel := context.WithTimeout(gctx, approvalWaitTimeout)
	defer cancel()

	for {
		// Wait asks the service to hold the request open until the login is decided,
		// which is what keeps the terminal from freezing for a poll interval after
		// the person already answered.
		status, err := s.service.GetSSHApprovalStatus(ctx, &requests.SSHApprovalStatus{
			Code: s.ApprovalCode,
			Wait: true,
		})
		if err == nil {
			switch status.State {
			case models.SSHApprovalConfirmed:
				return status.UserID, nil
			case models.SSHApprovalRejected:
				return "", ErrApprovalRejected
			}
		}
		// Still pending means the API's wait elapsed, and an error may be a
		// transient fault or a just-expired code: ask again either way, and let the
		// context bound the wait rather than aborting a valid approval.

		select {
		case <-ctx.Done():
			return "", ErrApprovalTimeout
		case <-time.After(approvalAskInterval):
		}
	}
}

// authorize is the hard identity-mode gate: it refuses here, before any key is
// minted, so the agent is never contacted for a login the policies deny. It
// returns the decision so the caller can honor a step-up requirement.
//
// A refusal is reported here rather than by the callers: this is the last point
// at which a denial, a failed evaluation and a missing decision are still
// distinguishable, and every caller refuses through it.
func (s *Session) authorize(ctx context.Context) (*models.Decision, error) {
	dec, err := s.service.Authorize(ctx, s.Namespace.TenantID, s.UserID, s.Device.UID, s.Target.Username, s.IPAddress)
	if err == nil && dec != nil && dec.Allowed {
		return dec, nil
	}

	// The log is the only place the reason can go: the sentinel below carries no
	// detail, and must not, since the reason names policies the client may not be
	// entitled to know about.
	logger := log.WithFields(s.LogFields()).WithFields(log.Fields{
		"tenant": s.Namespace.TenantID,
		"user":   s.UserID,
	})

	switch {
	case err != nil:
		logger.WithError(err).Error("failed to evaluate the access policies")
	case dec == nil:
		logger.Error("access policy evaluation returned no decision")
	default:
		logger.WithField("reason", dec.Reason).Warn("ssh access denied by the access policies")
	}

	return nil, ErrAccessDenied
}

// approvalAuth authenticates a session whose presented key is not an identity
// yet (identity mode): it blocks until a member approves in the console — which
// binds the key to the approving account — then authorizes the now-known
// identity and reaches the agent with the server-minted ephemeral key.
type approvalAuth struct {
	ctx gliderssh.Context
}

func AuthApproval(ctx gliderssh.Context) Auth {
	return &approvalAuth{ctx: ctx}
}

func (*approvalAuth) Auth() authFunc {
	return mintEphemeralSigner
}

func (*approvalAuth) Offer(*Session) error {
	// An unenrolled key is acceptable — turning it into an identity is the whole
	// point of the flow — but nothing may be parked or written until the client
	// has proven it holds the key.
	return nil
}

func (a *approvalAuth) Evaluate(session *Session) error {
	release, err := session.holdApprovalSlot()
	if err != nil {
		return err
	}

	defer release()

	if err := session.openApproval(a.ctx, models.SSHApprovalIdentity, nil); err != nil {
		return err
	}

	// Now that the key is known to be unknown and proven to be theirs, surface
	// the approval URL and hold here.
	sendBanner(a.ctx, buildAddKeyBanner(sshconf.Domain, sshconf.AutoSSL, session.ApprovalCode, session.Fingerprint))

	approver, err := session.awaitApproval(a.ctx)
	if err != nil {
		return err
	}

	// The approval bound the key to the approver; take that identity, then let
	// Access Policies decide whether it may reach this device as this login.
	session.UserID = approver

	_, err = session.authorize(a.ctx)

	return err
}

// identityAuth authenticates a session whose presented key already is an
// identity: it skips the browser step, authorizes the resolved account, runs the
// re-auth the policy may demand, and reaches the agent with the server-minted
// ephemeral key.
type identityAuth struct {
	ctx gliderssh.Context
}

func AuthIdentity(ctx gliderssh.Context) Auth {
	return &identityAuth{ctx: ctx}
}

func (*identityAuth) Auth() authFunc {
	return mintEphemeralSigner
}

func (*identityAuth) Offer(*Session) error {
	// The key resolved to a live identity, which is all that can be decided
	// before the client proves it holds it.
	return nil
}

func (a *identityAuth) Evaluate(session *Session) error {
	dec, err := session.authorize(a.ctx)
	if err != nil {
		// A web session shows the denial in the console; surface it as an
		// access-denied banner so the bridge maps it to a clear "Access denied"
		// instead of a generic auth failure. Native clients read the raw error.
		if session.Web {
			sendBanner(a.ctx, banner.Message(banner.KindAccessDenied))
		}

		return err
	}

	if dec.RequireReauth && needsReauth(session.LastReauthAt, dec.ReauthPeriod) {
		release, err := session.holdApprovalSlot()
		if err != nil {
			return err
		}

		defer release()

		// The identity is already established by the key, so the decision only
		// refreshes its window. The policy's period rides along so the console can
		// say how long that lasts.
		if err := session.openApproval(a.ctx, models.SSHApprovalReauth, dec.ReauthPeriod); err != nil {
			return err
		}

		// Native and web wait on the same decision — the login is held either way,
		// and only the way the person reaches the screen differs. A native client
		// gets a human banner with the URL to open; the web bridge gets a
		// machine-readable one carrying the code, and opens the same screen over the
		// terminal it already has.
		if session.Web {
			sendBanner(a.ctx, banner.MessageWithCode(banner.KindReauthRequired, session.ApprovalCode))
		} else {
			sendBanner(a.ctx, buildReauthBanner(sshconf.Domain, sshconf.AutoSSL, session.ApprovalCode))
		}

		if _, err := session.awaitApproval(a.ctx); err != nil {
			return err
		}
	}

	return nil
}

// needsReauth reports whether a policy requiring re-authentication must challenge
// this connection, given when the identity last re-authed and the policy's
// freshness window in seconds. A nil or zero period means "always" (every
// session); a never-re-authed identity always challenges; otherwise it is fresh
// while within the window.
func needsReauth(lastReauthAt *time.Time, period *int) bool {
	if period == nil || *period == 0 {
		return true
	}

	if lastReauthAt == nil {
		return true
	}

	return clock.Now().Sub(*lastReauthAt) >= time.Duration(*period)*time.Second
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

// ResolveKeyAuth resolves the presented key to a ShellHub identity (identity
// mode) and returns the auth to run: a hit yields the identity auth (authorize +
// mint, no browser); a miss yields the approval auth, which arranges the browser
// approval once the client has proven it holds the key.
//
// It is a lookup and nothing more. It runs for a key the client has only
// offered, so it must not write.
func (s *Session) ResolveKeyAuth(ctx gliderssh.Context, publicKey gliderssh.PublicKey) (Auth, error) {
	s.Fingerprint = gossh.FingerprintSHA256(publicKey)
	s.KeyData = gossh.MarshalAuthorizedKey(publicKey)

	identity, found, err := s.service.ResolveSSHIdentity(ctx, s.Namespace.TenantID, s.Fingerprint)
	if err != nil {
		return nil, err
	}

	if found {
		// A found-but-inactive key is a dead one: expired, or a single-use key
		// already burned. Reject the connection cleanly; it must never fall into
		// the interactive enrollment flow, which is for genuinely unknown keys.
		if !identity.Active(clock.Now()) {
			return nil, ErrAccessDenied
		}

		s.UserID = identity.PrincipalID
		s.LastReauthAt = identity.LastReauthAt
		s.SingleUse = identity.SingleUse

		return AuthIdentity(ctx), nil
	}

	// A web session registers its browser key via the API before connecting, so a
	// miss here means the key was revoked: reject cleanly. The URL approval below
	// is for native clients only — meaningless for a browser already in the
	// console, which just registers again and retries.
	if s.Web {
		return nil, ErrAccessDenied
	}

	return AuthApproval(ctx), nil
}

var (
	// ErrApprovalRejected is returned when the user rejects the login in the console.
	ErrApprovalRejected = errors.New("ssh login denied")
	// ErrAccessDenied is returned when no Access Policy grants the approved
	// identity access to the target device as the requested login.
	ErrAccessDenied = errors.New("ssh access denied by policy")
	// ErrApprovalTimeout is returned when no decision arrives before the wait
	// deadline or the client disconnects.
	ErrApprovalTimeout = errors.New("ssh login approval timed out")
)
