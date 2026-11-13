package services

import (
	"context"
	"errors"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

// SSHApprovalTTL is how long an approval can be confirmed for. A caller holding
// a login open while someone decides must outlast it, or it cuts off a person
// the console would still accept.
const SSHApprovalTTL = 90 * time.Second

// SSHApprovalService mediates connections that need a person to approve them: the gateway
// parks the request behind a prompt, and the console confirms it with a code the person types back.
type SSHApprovalService interface {
	// CreateSSHApproval stores a pending approval and returns a short-lived code
	// the SSH gateway shows on its approval prompt. The code itself is the
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
	// approve permission. On an identity approval it binds no key: the key becomes
	// an identity only when RedeemSSHApproval receives the code.
	//
	// It returns the confirmation code to show the approver, which they carry to
	// the terminal the login is waiting at. It returns ErrSSHApprovalCodeNotFound
	// for a malformed, unknown, expired or already decided code;
	// ErrNamespaceNotFound when the target namespace is gone;
	// ErrNamespaceMemberNotFound for a user outside it; ErrRoleForbidden for a
	// role that cannot approve logins or add SSH keys; ErrForbidden for a reauth
	// approval, which only the step-up route releases; ErrSSHIdentityDuplicated
	// when the key is already another account's identity in the namespace; and
	// the store's error otherwise.
	ConfirmSSHApproval(ctx context.Context, userID string, req *requests.SSHApprovalConfirm) (string, error)

	// RedeemSSHApproval binds the key of a confirmed identity approval to the
	// account that confirmed it, once the gateway has the confirmation code typed
	// on the connection that asked. It consumes the approval, so a code binds
	// once. The gateway calls it only for a login the access policy grants the
	// approver.
	//
	// It returns ErrSSHApprovalCodeNotFound when no confirmed, unexpired
	// identity approval matches the code and confirmation code;
	// ErrNamespaceNotFound when the namespace is gone; ErrNamespaceMemberNotFound
	// when the approver has left it and ErrRoleForbidden when their role can no
	// longer add SSH keys, since the approver's authority is read again here;
	// ErrSSHIdentityDuplicated when another account bound the key after the
	// approval was confirmed; and the store's error otherwise. On any error the
	// approval stays unconsumed and nothing is bound.
	RedeemSSHApproval(ctx context.Context, req *requests.SSHApprovalRedeem) error

	// RejectSSHApproval rejects a pending approval. Same authorization as
	// confirm.
	RejectSSHApproval(ctx context.Context, userID string, req *requests.SSHApprovalReject) error
}

func (s *service) CreateSSHApproval(ctx context.Context, req *requests.SSHApprovalCreate) (*models.SSHApprovalCreated, error) {
	code := pairingcode.New(pairingcode.DeviceCodeLength)

	now := s.clock.Now()
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
		ExpiresAt:    now.Add(SSHApprovalTTL),
	}); err != nil {
		return nil, err
	}

	return &models.SSHApprovalCreated{
		Code:      code,
		ExpiresIn: int(SSHApprovalTTL.Seconds()),
	}, nil
}

func (s *service) GetSSHApprovalStatus(ctx context.Context, req *requests.SSHApprovalStatus) (*models.SSHApprovalStatus, error) {
	code := pairingcode.Normalize(req.Code)

	approval, err := s.store.SSHApprovalGet(ctx, code, s.clock.Now())
	if err != nil {
		return nil, NewErrSSHApprovalCodeNotFound(code, err)
	}

	return &models.SSHApprovalStatus{
		State:            approval.State,
		UserID:           approval.DecidedBy,
		ConfirmationCode: approval.ConfirmationCode,
	}, nil
}

func (s *service) GetSSHApproval(ctx context.Context, userID, code string) (*models.SSHApprovalRequest, error) {
	code = pairingcode.Normalize(code)

	now := s.clock.Now()

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

func (s *service) ConfirmSSHApproval(ctx context.Context, userID string, req *requests.SSHApprovalConfirm) (string, error) {
	confirmationCode := pairingcode.New(pairingcode.DeviceCodeLength)

	if err := s.decideSSHApproval(ctx, req.Code, models.SSHApprovalDecision{
		State:             models.SSHApprovalConfirmed,
		DecidedBy:         userID,
		ConfirmationCode:  confirmationCode,
		IdentityExpiresIn: req.ExpiresIn,
	}); err != nil {
		return "", err
	}

	return confirmationCode, nil
}

func (s *service) RejectSSHApproval(ctx context.Context, userID string, req *requests.SSHApprovalReject) error {
	return s.decideSSHApproval(ctx, req.Code, models.SSHApprovalDecision{
		State:             models.SSHApprovalRejected,
		DecidedBy:         userID,
		ConfirmationCode:  "",
		IdentityExpiresIn: nil,
	})
}

func (s *service) RedeemSSHApproval(ctx context.Context, req *requests.SSHApprovalRedeem) error {
	code := pairingcode.Normalize(req.Code)
	confirmationCode := pairingcode.Normalize(req.ConfirmationCode)

	err := s.redeemSSHApproval(ctx, code, confirmationCode)
	if errors.Is(err, ErrSSHIdentityDuplicated) {
		err = s.redeemSSHApproval(ctx, code, confirmationCode)
	}

	return err
}

func (s *service) redeemSSHApproval(ctx context.Context, code, confirmationCode string) error {
	return s.store.WithTransaction(ctx, func(ctx context.Context) error {
		approval, err := s.store.SSHApprovalRedeem(ctx, code, confirmationCode, s.clock.Now())
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrSSHApprovalCodeNotFound(code, err)
		}

		if err != nil {
			return err
		}

		if approval.Kind != models.SSHApprovalIdentity {
			return NewErrSSHApprovalCodeNotFound(code, nil)
		}

		if err := s.requireSSHApprovalBinder(ctx, approval); err != nil {
			return err
		}

		return s.bindSSHApprovalKey(ctx, approval)
	})
}

func (s *service) decideSSHApproval(ctx context.Context, code string, decision models.SSHApprovalDecision) error {
	code = pairingcode.Normalize(code)
	if !pairingcode.IsValid(code, pairingcode.DeviceCodeLength) {
		return NewErrSSHApprovalCodeNotFound(code, nil)
	}

	now := s.clock.Now()

	approval, err := s.store.SSHApprovalGet(ctx, code, now)
	if err != nil {
		return NewErrSSHApprovalCodeNotFound(code, err)
	}

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, approval.TenantID)
	if err != nil {
		return NewErrNamespaceNotFound(approval.TenantID, err)
	}

	member, ok := namespace.FindMember(decision.DecidedBy)
	if !ok {
		return NewErrNamespaceMemberNotFound(decision.DecidedBy, nil)
	}

	if !member.Role.HasPermission(authorizer.SessionApprove) {
		return NewErrRoleForbidden()
	}

	if decision.State == models.SSHApprovalConfirmed {
		if approval.Kind == models.SSHApprovalReauth {
			return NewErrForbidden(ErrForbidden, nil)
		}

		if !member.Role.HasPermission(authorizer.SSHIdentityAdd) {
			return NewErrRoleForbidden()
		}

		if err := s.requireSSHApprovalKeyFree(ctx, approval, decision.DecidedBy); err != nil {
			return err
		}
	}

	claimed, err := s.store.SSHApprovalDecide(ctx, code, decision, now)
	if err != nil {
		return err
	}

	if !claimed {
		return NewErrSSHApprovalCodeNotFound(code, nil)
	}

	return nil
}

func (s *service) requireSSHApprovalBinder(ctx context.Context, approval *models.SSHApproval) error {
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, approval.TenantID)
	if err != nil {
		return NewErrNamespaceNotFound(approval.TenantID, err)
	}

	member, ok := namespace.FindMember(approval.DecidedBy)
	if !ok {
		return NewErrNamespaceMemberNotFound(approval.DecidedBy, nil)
	}

	if !member.Role.HasPermission(authorizer.SSHIdentityAdd) {
		return NewErrRoleForbidden()
	}

	return nil
}

func (s *service) requireSSHApprovalKeyFree(ctx context.Context, approval *models.SSHApproval, userID string) error {
	existing, err := s.resolveEnrolledSSHIdentity(ctx, approval.TenantID, approval.Fingerprint)
	if err != nil {
		return err
	}

	if existing != nil && existing.PrincipalID != userID {
		return NewErrSSHIdentityDuplicated(approval.Fingerprint, nil)
	}

	return nil
}

func (s *service) bindSSHApprovalKey(ctx context.Context, approval *models.SSHApproval) error {
	if _, err := s.reenrollSSHIdentity(ctx, &models.SSHIdentity{
		TenantID:    approval.TenantID,
		PrincipalID: approval.DecidedBy,
		Fingerprint: approval.Fingerprint,
		Data:        approval.Data,
		Source:      models.SSHIdentitySourceApproval,
		ExpiresAt:   sshIdentityExpiry(s.clock.Now(), approval.IdentityExpiresIn),
	}); err != nil {
		return err
	}

	if err := s.store.SSHIdentityTouchLastUsed(ctx, approval.TenantID, approval.Fingerprint); err != nil {
		log.WithError(err).WithField("fingerprint", approval.Fingerprint).
			Warn("failed to stamp ssh identity last-used on approval; connection proceeds")
	}

	return nil
}
