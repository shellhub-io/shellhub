package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// WebReauthService re-checks a logged-in user's identity before a sensitive action, so that a
// stolen session alone is not enough to perform one.
type WebReauthService interface {
	// WebReauthVerify validates the step-up factor for the logged-in user and, on
	// success, stamps the re-auth freshness marker. Community verifies the
	// account password; the enterprise overlay adds TOTP for MFA users.
	WebReauthVerify(ctx context.Context, req *requests.WebReauthVerify) error
}

func (s *service) WebReauthVerify(ctx context.Context, req *requests.WebReauthVerify) error {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil {
		return NewErrUserNotFound(req.UserID, err)
	}

	if !user.Password.Compare(req.Password) {
		return NewErrUserPasswordNotMatch(nil)
	}

	return StampWebReauth(ctx, s.store, req)
}

// StampWebReauth records a successful step-up on the identity that presented the
// key, and releases the login waiting on it. Exported so the enterprise overlay,
// which swaps in TOTP factor validation, gets both identically instead of
// duplicating them.
//
// Owning the identity is the authorization: the caller proved a factor and the
// key resolves to their own account, which is stricter than the membership the
// approval routes check. So no separate permission gate is needed here — and a
// login can only be released by the person whose key it is.
func StampWebReauth(ctx context.Context, st store.Store, req *requests.WebReauthVerify) error {
	sc, err := scope.NewBounded(req.TenantID)
	if err != nil {
		return NewErrForbidden(ErrForbidden, err)
	}

	identity, err := st.SSHIdentityResolve(ctx, sc, store.SSHIdentityFingerprintResolver, req.Fingerprint)
	if err != nil || identity.PrincipalID != req.UserID {
		return NewErrForbidden(ErrForbidden, nil)
	}

	return st.WithTransaction(ctx, func(ctx context.Context) error {
		if err := st.SSHIdentityTouchReauth(ctx, req.TenantID, req.Fingerprint); err != nil {
			return err
		}

		if req.ApprovalCode == "" {
			return nil
		}

		return releaseSSHApproval(ctx, st, req)
	})
}

func releaseSSHApproval(ctx context.Context, st store.Store, req *requests.WebReauthVerify) error {
	now := clock.Now()

	approval, err := st.SSHApprovalGet(ctx, pairingcode.Normalize(req.ApprovalCode), now)
	if err != nil {
		return NewErrSSHApprovalCodeNotFound(req.ApprovalCode, err)
	}

	mismatched := approval.Kind != models.SSHApprovalReauth ||
		approval.TenantID != req.TenantID ||
		approval.Fingerprint != req.Fingerprint
	if mismatched {
		return NewErrForbidden(ErrForbidden, nil)
	}

	claimed, err := st.SSHApprovalDecide(ctx, approval.Code, models.SSHApprovalConfirmed, req.UserID, now)
	if err != nil {
		return err
	}

	if !claimed {
		return NewErrSSHApprovalCodeNotFound(req.ApprovalCode, nil)
	}

	return nil
}
