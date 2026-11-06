package session

import (
	"context"
	"errors"
	"time"

	"github.com/Masterminds/semver/v3"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/banner"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type authFunc func(context.Context, *Session, *gossh.ClientConfig) error

func mintEphemeralSigner(ctx context.Context, session *Session, config *gossh.ClientConfig) error {
	key, err := session.service.MintDeviceSigner(ctx, models.UID(session.Device.UID))
	if err != nil {
		return err
	}

	signer, err := gossh.NewSignerFromKey(key)
	if err != nil {
		return err
	}

	config.Auth = []gossh.AuthMethod{
		gossh.PublicKeys(signer),
	}

	return nil
}

// Auth is one authentication attempt on a session. The grant a session registers and
// joins the device with comes only from an Auth's Evaluate, or from approving an Auth
// that Evaluate parked, so a session never reaches the device on an attempt nobody
// evaluated. The package's own constructors are its only implementations.
//
// The two checks are split by what the client has proven. The SSH protocol lets
// a client ask whether a key would be acceptable without signing anything, and
// x/crypto runs the publickey callback for that query, so anything reachable
// before a signature must stay cheap and free of side effects.
type Auth interface {
	// Offer decides whether the credential is acceptable at all. It may run before
	// the client has proven it holds the key, even for a key it does not hold, so
	// it must only read. Every read uses ctx, the client connection's context.
	// Only the public key auth reads anything; the others accept.
	//
	// Unless the instance allows agents below 0.6.0, the public key auth returns
	// ErrInvalidVersion for an agent version it cannot parse and
	// ErrUnsuportedPublicKeyAuth for one below 0.6.0. An agent reporting "latest"
	// skips this check. The key lookup returns ErrForbidden for an empty tenant,
	// ErrNamespaceNotFound when the namespace does not resolve, a disconnect
	// during that lookup included, ErrPublicKeyNotFound for a fingerprint with no
	// match, or the store's error. Once the client has disconnected, the lookup's
	// error wraps ctx's error. It returns ErrEvaluatePublicKey for a key the
	// username or the filter refuses or whose evaluation fails, a disconnect
	// during evaluation included.
	Offer(context.Context, *Session) error

	// Evaluate runs once the credential is proven, and is where anything with a
	// cost or a side effect belongs. On success it returns the grant the attempt
	// earned: the principal the session is recorded as, how the gateway logs in to
	// the device, and the single-use identity to redeem, if any.
	//
	// Returning ErrApprovalRequired parks the login: the credential is good but
	// a person has to decide, and the gateway asks them over a second
	// authentication method. Evaluate never waits for that decision itself. A
	// session already registered cannot park, and gets ErrInvalidSessionState
	// instead, and one with an approval of that kind already open gets
	// ErrChallengeAlreadyIssued. The identity auths return ErrAccessDenied when no
	// Access Policy grants the principal, and the error from opening the approval.
	Evaluate(*Session) (grant, error)
}

type parked interface {
	Auth
	presented() presentedKey
	approved(session *Session, approver string) (grant, error)
}

type publicKeyAuth struct {
	pk gliderssh.PublicKey
}

// AuthPublicKey authenticates the client to the server by the key pk it presented. The device
// is then reached with an ephemeral key minted by the server, not with pk itself.
func AuthPublicKey(pk gliderssh.PublicKey) Auth {
	return &publicKeyAuth{pk: pk}
}

func (*publicKeyAuth) Evaluate(session *Session) (grant, error) {
	return session.legacyGrant(mintEphemeralSigner), nil
}

func (p *publicKeyAuth) Offer(ctx context.Context, session *Session) error {
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

	return nil
}

type passwordAuth struct {
	pwd string
}

// AuthPassword authenticates to the device with pwd, which is forwarded as the client typed
// it because only the device can check it against its own accounts.
func AuthPassword(pwd string) Auth {
	return &passwordAuth{pwd: pwd}
}

func (p *passwordAuth) method(_ context.Context, _ *Session, config *gossh.ClientConfig) error {
	config.Auth = []gossh.AuthMethod{
		gossh.Password(p.pwd),
	}

	return nil
}

func (p *passwordAuth) Evaluate(session *Session) (grant, error) {
	return session.legacyGrant(p.method), nil
}

func (*passwordAuth) Offer(context.Context, *Session) error {
	return nil
}

func (s *Session) approvalDecision(ctx context.Context) (*models.SSHApprovalStatus, error) {
	status, err := s.service.GetSSHApprovalStatus(ctx, &requests.SSHApprovalStatus{Code: s.ApprovalCode})
	if err != nil {
		if errors.Is(err, services.ErrSSHApprovalCodeNotFound) {
			return nil, ErrApprovalExpired
		}

		return nil, err
	}

	switch status.State {
	case models.SSHApprovalConfirmed:
		return status, nil
	case models.SSHApprovalRejected:
		return nil, ErrApprovalRejected
	default:
		return nil, ErrApprovalPending
	}
}

func (s *Session) authorize(ctx context.Context, principal models.Principal) (*models.Decision, error) {
	dec, err := s.service.Authorize(ctx, s.Namespace.TenantID, principal, s.Device.UID, s.Target.Username, s.IPAddress)
	if err == nil && dec != nil && dec.Allowed {
		return dec, nil
	}

	logger := log.WithFields(s.LogFields()).WithFields(log.Fields{
		"tenant": s.Namespace.TenantID,
		"user":   principal.ID,
	})

	switch {
	case err != nil:
		logger.WithError(err).Error("failed to evaluate the access policies")
	case dec == nil:
		logger.Error("access policy evaluation returned no decision")
	default:
		fields := log.Fields{"reason": dec.Reason, "detail": dec.Message()}
		if dec.PolicyName != "" {
			fields["policy_name"] = dec.PolicyName
		}

		logger.WithFields(fields).Warn("ssh access denied by the access policies")
	}

	return nil, ErrAccessDenied
}

type approvalAuth struct {
	ctx gliderssh.Context
	key presentedKey
}

func (*approvalAuth) Offer(context.Context, *Session) error {
	return nil
}

func (a *approvalAuth) Evaluate(session *Session) (grant, error) {
	if err := session.beginChallenge(a.ctx, a, models.SSHApprovalIdentity, nil); err != nil {
		return grant{}, err
	}

	return grant{}, ErrApprovalRequired
}

func (a *approvalAuth) presented() presentedKey {
	return a.key
}

func (a *approvalAuth) approved(session *Session, approver string) (grant, error) {
	principal := models.Principal{Kind: models.PrincipalUser, ID: approver}

	if _, err := session.authorize(a.ctx, principal); err != nil {
		return grant{}, err
	}

	return grant{principal: principal, method: mintEphemeralSigner, redeems: ""}, nil
}

type identityAuth struct {
	ctx      gliderssh.Context
	key      presentedKey
	identity *models.SSHIdentity
}

func (*identityAuth) Offer(context.Context, *Session) error {
	return nil
}

func (a *identityAuth) Evaluate(session *Session) (grant, error) {
	g := identityGrant(a.identity, a.key)

	dec, err := session.authorize(a.ctx, g.principal)
	if err != nil {
		if session.Web {
			sendBanner(a.ctx, banner.Message(banner.KindAccessDenied))
		}

		return grant{}, err
	}

	if !dec.RequireReauth || !needsReauth(session.clock.Now(), a.identity.LastReauthAt, dec.ReauthPeriod) {
		return g, nil
	}

	if err := session.beginChallenge(a.ctx, a, models.SSHApprovalReauth, dec.ReauthPeriod); err != nil {
		return grant{}, err
	}

	return grant{}, ErrApprovalRequired
}

func (a *identityAuth) presented() presentedKey {
	return a.key
}

func (a *identityAuth) approved(*Session, string) (grant, error) {
	return identityGrant(a.identity, a.key), nil
}

func needsReauth(now time.Time, lastReauthAt *time.Time, period *int) bool {
	if period == nil || *period == 0 {
		return true
	}

	if lastReauthAt == nil {
		return true
	}

	return now.Sub(*lastReauthAt) >= time.Duration(*period)*time.Second
}

type preAuthConnCtxKey struct{}

// StorePreAuthConn stashes the x/crypto pre-auth connection so a banner can be
// sent mid-handshake, after the presented key is known. The gateway calls it
// from PreAuthConnCallback.
func StorePreAuthConn(ctx gliderssh.Context, conn gossh.ServerPreAuthConn) {
	ctx.SetValue(preAuthConnCtxKey{}, conn)
}

func sendBanner(ctx gliderssh.Context, msg string) {
	if conn, ok := ctx.Value(preAuthConnCtxKey{}).(gossh.ServerPreAuthConn); ok && conn != nil {
		_ = conn.SendAuthBanner(msg)
	}
}

// ResolveKeyAuth resolves the presented key to a ShellHub identity (identity
// mode) and returns the auth to run: a hit yields the identity auth (authorize +
// mint, no browser); a miss yields the approval auth, which arranges the browser
// approval once the client has proven it holds the key.
//
// It is a lookup and nothing more. It runs for a key the client has only
// offered, so it writes nothing to the session: what it resolved lives in the
// returned auth. It returns ErrAccessDenied for an identity that is consumed or
// expired, and for an unknown key on a web session, and the lookup's error.
func (s *Session) ResolveKeyAuth(ctx gliderssh.Context, publicKey gliderssh.PublicKey) (Auth, error) {
	key := presentKey(publicKey)

	identity, found, err := s.service.ResolveSSHIdentity(ctx, s.Namespace.TenantID, key.fingerprint)
	if err != nil {
		return nil, err
	}

	if found {
		if !identity.Active(s.clock.Now()) {
			return nil, ErrAccessDenied
		}

		return &identityAuth{ctx: ctx, key: key, identity: identity}, nil
	}

	if s.Web {
		return nil, ErrAccessDenied
	}

	return &approvalAuth{ctx: ctx, key: key}, nil
}

var (
	// ErrApprovalRejected is returned when the user rejects the login in the console.
	ErrApprovalRejected = errors.New("ssh login denied")
	// ErrAccessDenied is returned when no Access Policy grants the approved
	// identity access to the target device as the requested login.
	ErrAccessDenied = errors.New("ssh access denied by policy")
	// ErrApprovalRequired is returned by Evaluate when the credential is good
	// but a person still has to decide. It is not a failure: the caller parks
	// the login and asks the client over a second authentication method.
	ErrApprovalRequired = errors.New("ssh login needs approval")
	// ErrApprovalPending is returned when the client answered its challenge but
	// nobody has decided yet.
	ErrApprovalPending = errors.New("ssh login not approved yet")
	// ErrApprovalExpired is returned when the approval is no longer on record,
	// which is what a client that sat at the prompt past the approval's TTL
	// gets. Reconnecting mints a new one.
	ErrApprovalExpired = errors.New("ssh login approval expired")
	// ErrConfirmationMismatch is returned when the answer typed at the terminal
	// is not the confirmation code the console showed the approver.
	ErrConfirmationMismatch = errors.New("ssh login confirmation code does not match")
	// ErrInvalidSessionState is returned when a connection asks to authenticate
	// from a state the session cannot serve, which is a client reusing a context
	// the gateway has already taken past that point.
	ErrInvalidSessionState = errors.New("invalid session state")
	// ErrPromptDismissed is returned when the challenge came back with no answer
	// at all, which is a dismissed dialog or a client with nobody to ask. Unlike
	// a wrong code it is not worth asking again for.
	ErrPromptDismissed = errors.New("ssh login prompt dismissed")
	// ErrChallengeAlreadyIssued is returned when a connection asks a second time
	// about the same thing, which is how a client holding several unenrolled keys
	// is kept to a single enrollment prompt. A different kind still gets asked: a
	// key that is enrolled and needs a re-auth is another question about another
	// key, and refusing it would strand a client whose earlier prompt was
	// dismissed.
	ErrChallengeAlreadyIssued = errors.New("ssh login already prompted for approval")
)
