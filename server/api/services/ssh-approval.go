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

const statusWaitTimeout = 20 * time.Second

const statusWaitInterval = 200 * time.Millisecond

func (s *service) GetSSHApprovalStatus(ctx context.Context, req *requests.SSHApprovalStatus) (*models.SSHApprovalStatus, error) {
	code := pairingcode.Normalize(req.Code)

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
		if approval.Kind == models.SSHApprovalReauth {
			return NewErrForbidden(ErrForbidden, nil)
		}

		if !member.Role.HasPermission(authorizer.SSHIdentityAdd) {
			return NewErrRoleForbidden()
		}
	}

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

func (s *service) applySSHApproval(ctx context.Context, userID string, approval *models.SSHApproval, expiresIn *int) error {
	if approval.Kind != models.SSHApprovalIdentity {
		return nil
	}

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

	if err := s.store.SSHIdentityTouchLastUsed(ctx, approval.TenantID, approval.Fingerprint); err != nil {
		log.WithError(err).WithField("fingerprint", approval.Fingerprint).
			Warn("failed to stamp ssh identity last-used on approval; connection proceeds")
	}

	return nil
}
