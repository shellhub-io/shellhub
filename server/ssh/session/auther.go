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
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/banner"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type authFunc func(*Session, *gossh.ClientConfig) error

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
	//
	// Returning ErrApprovalRequired parks the login: the credential is good but
	// a person has to decide, and the gateway asks them over a second
	// authentication method. Evaluate never waits for that decision itself.
	Evaluate(*Session) error

	// Approved finishes what Evaluate parked, and runs only after Evaluate
	// returned ErrApprovalRequired and the approval came back confirmed.
	// approver is who confirmed it.
	Approved(session *Session, approver string) error
}

type publicKeyAuth struct {
	pk gliderssh.PublicKey
}

// AuthPublicKey authenticates the client to the server by the key pk it presented. The device
// is then reached with an ephemeral key minted by the server, not with pk itself.
func AuthPublicKey(pk gliderssh.PublicKey) Auth {
	return &publicKeyAuth{pk: pk}
}

func (*publicKeyAuth) Auth() authFunc {
	return mintEphemeralSigner
}

func (*publicKeyAuth) Evaluate(*Session) error {
	return nil
}

func (*publicKeyAuth) Approved(*Session, string) error {
	return nil
}

func (p *publicKeyAuth) Offer(session *Session) error {
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

	ctx := context.Background()
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

func (p *passwordAuth) Auth() authFunc {
	return func(_ *Session, config *gossh.ClientConfig) error {
		config.Auth = []gossh.AuthMethod{
			gossh.Password(p.pwd),
		}

		return nil
	}
}

func (*passwordAuth) Evaluate(*Session) error {
	return nil
}

func (*passwordAuth) Approved(*Session, string) error {
	return nil
}

func (*passwordAuth) Offer(*Session) error {
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

func (s *Session) authorize(ctx context.Context) (*models.Decision, error) {
	principal := models.Principal{Kind: s.PrincipalKind, ID: s.UserID}
	if principal.Kind == "" {
		principal.Kind = models.PrincipalUser
	}

	dec, err := s.service.Authorize(ctx, s.Namespace.TenantID, principal, s.Device.UID, s.Target.Username, s.IPAddress)
	if err == nil && dec != nil && dec.Allowed {
		return dec, nil
	}

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
}

// AuthApproval authenticates a session that an administrator has approved out of band,
// reading the approval from ctx.
func AuthApproval(ctx gliderssh.Context) Auth {
	return &approvalAuth{ctx: ctx}
}

func (*approvalAuth) Auth() authFunc {
	return mintEphemeralSigner
}

func (*approvalAuth) Offer(*Session) error {
	return nil
}

func (a *approvalAuth) Evaluate(session *Session) error {
	if err := session.beginChallenge(a.ctx, a, models.SSHApprovalIdentity, nil); err != nil {
		return err
	}

	return ErrApprovalRequired
}

func (a *approvalAuth) Approved(session *Session, approver string) error {
	session.UserID = approver

	_, err := session.authorize(a.ctx)

	return err
}

type identityAuth struct {
	ctx gliderssh.Context
}

// AuthIdentity authenticates a session as a stored SSH identity, reading it from ctx.
func AuthIdentity(ctx gliderssh.Context) Auth {
	return &identityAuth{ctx: ctx}
}

func (*identityAuth) Auth() authFunc {
	return mintEphemeralSigner
}

func (*identityAuth) Offer(*Session) error {
	return nil
}

func (a *identityAuth) Evaluate(session *Session) error {
	dec, err := session.authorize(a.ctx)
	if err != nil {
		if session.Web {
			sendBanner(a.ctx, banner.Message(banner.KindAccessDenied))
		}

		return err
	}

	if !dec.RequireReauth || !needsReauth(session.LastReauthAt, dec.ReauthPeriod) {
		return nil
	}

	if err := session.beginChallenge(a.ctx, a, models.SSHApprovalReauth, dec.ReauthPeriod); err != nil {
		return err
	}

	return ErrApprovalRequired
}

func (*identityAuth) Approved(*Session, string) error {
	return nil
}

func needsReauth(lastReauthAt *time.Time, period *int) bool {
	if period == nil || *period == 0 {
		return true
	}

	if lastReauthAt == nil {
		return true
	}

	return clock.Now().Sub(*lastReauthAt) >= time.Duration(*period)*time.Second
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
// offered, so it must not write.
func (s *Session) ResolveKeyAuth(ctx gliderssh.Context, publicKey gliderssh.PublicKey) (Auth, error) {
	s.Fingerprint = gossh.FingerprintSHA256(publicKey)
	s.KeyData = gossh.MarshalAuthorizedKey(publicKey)

	identity, found, err := s.service.ResolveSSHIdentity(ctx, s.Namespace.TenantID, s.Fingerprint)
	if err != nil {
		return nil, err
	}

	if found {
		if !identity.Active(clock.Now()) {
			return nil, ErrAccessDenied
		}

		s.UserID = identity.PrincipalID
		s.PrincipalKind = identity.PrincipalType
		s.LastReauthAt = identity.LastReauthAt
		s.SingleUse = identity.SingleUse

		return AuthIdentity(ctx), nil
	}

	if s.Web {
		return nil, ErrAccessDenied
	}

	s.UserID = ""
	s.PrincipalKind = ""
	s.LastReauthAt = nil
	s.SingleUse = false

	return AuthApproval(ctx), nil
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
