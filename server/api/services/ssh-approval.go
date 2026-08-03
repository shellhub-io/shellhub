package services

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

// sshApprovalTTL is how long an approval code remains valid. It is a
// just-in-time gate: the SSH gateway holds the login open for this long while
// the user decides in the console. It must not outlive the gateway's own wait
// (approvalWaitTimeout), so a code is never decidable after the connection
// waiting on it is gone.
const sshApprovalTTL = 90 * time.Second

type SSHApprovalService interface {
	// CreateSSHApproval stores a pending approval and returns a short-lived code
	// the SSH gateway embeds in the terminal banner. The code itself is the
	// secret; it expires with a short TTL.
	CreateSSHApproval(ctx context.Context, req *requests.SSHApprovalCreate) (*models.SSHApprovalCreated, error)

	// GetSSHApprovalStatus reports the decision to the gateway waiting on it. An
	// unknown or expired code returns not found.
	GetSSHApprovalStatus(ctx context.Context, req *requests.SSHApprovalStatus) (*models.SSHApprovalStatus, error)

	// GetSSHApproval returns the request details the console renders, for a
	// member of the namespace the login targeted. It is not-found for everyone
	// else, so a code alone never discloses the request.
	GetSSHApproval(ctx context.Context, userID, code string) (*models.SSHApprovalRequest, error)

	// ConfirmSSHApproval confirms a pending approval and binds the approving user
	// to it. The user must be a member of the target's namespace with the session
	// approve permission.
	ConfirmSSHApproval(ctx context.Context, userID string, req *requests.SSHApprovalConfirm) error

	// RejectSSHApproval rejects a pending approval. Same authorization as
	// confirm.
	RejectSSHApproval(ctx context.Context, userID string, req *requests.SSHApprovalReject) error
}

func (s *service) CreateSSHApproval(ctx context.Context, req *requests.SSHApprovalCreate) (*models.SSHApprovalCreated, error) {
	code, err := pairingcode.New(pairingcode.DeviceCodeLength)
	if err != nil {
		return nil, err
	}

	now := clock.Now()
	if err := s.store.SSHApprovalCreate(ctx, &models.SSHApproval{
		Code:         code,
		TenantID:     req.TenantID,
		Kind:         req.Kind,
		SessionUID:   req.SessionUID,
		SSHID:        req.SSHID,
		DeviceUID:    req.DeviceUID,
		DeviceName:   req.DeviceName,
		Username:     req.Username,
		IPAddress:    req.IPAddress,
		Fingerprint:  req.Fingerprint,
		Data:         req.Data,
		ReauthPeriod: req.ReauthPeriod,
		State:        models.SSHApprovalPending,
		RequestedAt:  now,
		ExpiresAt:    now.Add(sshApprovalTTL),
	}); err != nil {
		return nil, err
	}

	return &models.SSHApprovalCreated{
		Code:      code,
		ExpiresIn: int(sshApprovalTTL.Seconds()),
	}, nil
}

// statusWaitTimeout bounds how long a single waiting status request is held
// open. It is well under the gateway's own wait so a login is asked about
// several times while it is held, which is what lets a decision written by
// another API replica be picked up without any cross-replica signalling.
const statusWaitTimeout = 20 * time.Second

// statusWaitInterval is how often a waiting request re-reads the decision, and
// so is the floor on how long someone stares at a frozen terminal after they
// have already proven who they are. The read is a primary-key lookup on a table
// that only holds logins currently in flight, so it stays cheap at this rate.
const statusWaitInterval = 200 * time.Millisecond

func (s *service) GetSSHApprovalStatus(ctx context.Context, req *requests.SSHApprovalStatus) (*models.SSHApprovalStatus, error) {
	code := pairingcode.Normalize(req.Code)

	// The budget is a timer rather than a deadline on ctx so that expiring it
	// never cuts a read in flight, which would surface as an unknown code. ctx
	// stays the request's, so a gateway that hangs up still ends the wait.
	budget := time.NewTimer(statusWaitTimeout)
	defer budget.Stop()

	for {
		approval, err := s.store.SSHApprovalGet(ctx, code, clock.Now())
		if err != nil {
			return nil, NewErrSSHApprovalCodeNotFound(code, err)
		}

		status := &models.SSHApprovalStatus{
			State:  approval.State,
			UserID: approval.DecidedBy,
		}

		if !req.Wait || approval.State != models.SSHApprovalPending {
			return status, nil
		}

		// Answering "still pending" when the budget elapses is not a failure: the
		// gateway asks again, and that is also how it notices a client that hung up.
		select {
		case <-ctx.Done():
			return status, nil
		case <-budget.C:
			return status, nil
		case <-time.After(statusWaitInterval):
		}
	}
}

func (s *service) GetSSHApproval(ctx context.Context, userID, code string) (*models.SSHApprovalRequest, error) {
	code = pairingcode.Normalize(code)

	now := clock.Now()

	approval, err := s.store.SSHApprovalGet(ctx, code, now)
	if err != nil {
		return nil, NewErrSSHApprovalCodeNotFound(code, err)
	}

	// Only a member of the target namespace may read the request. Without this a
	// code holder from outside the namespace could learn the key fingerprint,
	// device, login and source IP of someone else's pending login. Deciding is
	// already gated (see decideSSHApproval); this closes the read side, and
	// reports not-found so a non-member cannot confirm the code exists.
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, approval.TenantID)
	if err != nil {
		return nil, NewErrSSHApprovalCodeNotFound(code, err)
	}

	if _, ok := namespace.FindMember(userID); !ok {
		return nil, NewErrSSHApprovalCodeNotFound(code, nil)
	}

	return &models.SSHApprovalRequest{
		SSHID:        approval.SSHID,
		DeviceName:   approval.DeviceName,
		Username:     approval.Username,
		IPAddress:    approval.IPAddress,
		RequestedAt:  approval.RequestedAt,
		State:        approval.State,
		Code:         code,
		Fingerprint:  approval.Fingerprint,
		Kind:         approval.Kind,
		ReauthPeriod: approval.ReauthPeriod,
		ExpiresIn:    int(approval.ExpiresAt.Sub(now).Seconds()),
		Namespace:    namespace.Name,
	}, nil
}

func (s *service) ConfirmSSHApproval(ctx context.Context, userID string, req *requests.SSHApprovalConfirm) error {
	return s.decideSSHApproval(ctx, userID, req.Code, models.SSHApprovalConfirmed, req.ExpiresIn)
}

func (s *service) RejectSSHApproval(ctx context.Context, userID string, req *requests.SSHApprovalReject) error {
	return s.decideSSHApproval(ctx, userID, req.Code, models.SSHApprovalRejected, nil)
}

// decideSSHApproval writes the confirm/reject decision after checking that the
// user is a member of the target's namespace with the approve permission. A
// reject carries no TTL, since it creates nothing.
func (s *service) decideSSHApproval(ctx context.Context, userID, code string, decision models.SSHApprovalState, expiresIn *int) error {
	code = pairingcode.Normalize(code)
	if !pairingcode.IsValid(code, pairingcode.DeviceCodeLength) {
		return NewErrSSHApprovalCodeNotFound(code, nil)
	}

	now := clock.Now()

	approval, err := s.store.SSHApprovalGet(ctx, code, now)
	if err != nil {
		return NewErrSSHApprovalCodeNotFound(code, err)
	}

	// The console session may be scoped to another namespace, so the gateway's
	// permission middleware cannot cover this route; check the user's role in the
	// target namespace explicitly.
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, approval.TenantID)
	if err != nil {
		return NewErrNamespaceNotFound(approval.TenantID, err)
	}

	member, ok := namespace.FindMember(userID)
	if !ok {
		return NewErrNamespaceMemberNotFound(userID, nil)
	}

	if !member.Role.HasPermission(authorizer.SessionApprove) {
		return NewErrRoleForbidden()
	}

	if decision == models.SSHApprovalConfirmed {
		// A re-auth is only satisfied by proving a factor, which happens on the
		// step-up route and releases the login there (see StampWebReauth).
		// Accepting a bare confirm here would be a way to skip the factor entirely.
		if approval.Kind == models.SSHApprovalReauth {
			return NewErrForbidden(ErrForbidden, nil)
		}

		// Binding a key to an account is a bigger act than releasing a login;
		// require the add permission on top of the approve permission.
		if !member.Role.HasPermission(authorizer.SSHIdentityAdd) {
			return NewErrRoleForbidden()
		}
	}

	// The claim and the effect share a transaction, so the gateway's poll never
	// reads a confirmation whose effect has not landed, and a failed effect rolls
	// the claim back — leaving the code decidable again instead of stuck.
	return s.store.WithTransaction(ctx, func(ctx context.Context) error {
		claimed, err := s.store.SSHApprovalDecide(ctx, code, decision, userID, now)
		if err != nil {
			return err
		}

		if !claimed {
			return NewErrSSHApprovalCodeNotFound(code, nil)
		}

		if decision != models.SSHApprovalConfirmed {
			return nil
		}

		return s.applySSHApproval(ctx, userID, approval, expiresIn)
	})
}

// applySSHApproval performs a confirmation's durable effect: binding the key as
// an identity. Only the identity kind ever gets here — a re-auth is refused
// above and released on the step-up route instead, once a factor is proved (see
// StampWebReauth). expiresIn is the TTL in days the person chose while
// confirming.
func (s *service) applySSHApproval(ctx context.Context, userID string, approval *models.SSHApproval, expiresIn *int) error {
	if approval.Kind != models.SSHApprovalIdentity {
		return nil
	}

	// A confirmation can be replayed, so this enrolls through the tolerant path.
	// An empty name lets the identity service generate a default from the key.
	// The source is the one the server asserts for itself: it saw the key offered
	// on a login it was holding open.
	if _, err := s.reenrollSSHIdentity(ctx, &models.SSHIdentity{
		TenantID:    approval.TenantID,
		PrincipalID: userID,
		Fingerprint: approval.Fingerprint,
		Data:        approval.Data,
		Source:      models.SSHIdentitySourceApproval,
		ExpiresAt:   sshIdentityExpiry(expiresIn),
	}); err != nil {
		return err
	}

	// The gateway's initial resolve missed (the key was not an identity yet), so
	// nothing stamped last-used for this connection. Stamp it here so the
	// connection that created the identity counts, not just the next one.
	// Non-fatal, as in ResolveSSHIdentity.
	if err := s.store.SSHIdentityTouchLastUsed(ctx, approval.TenantID, approval.Fingerprint); err != nil {
		log.WithError(err).WithField("fingerprint", approval.Fingerprint).
			Warn("failed to stamp ssh identity last-used on approval; connection proceeds")
	}

	return nil
}
