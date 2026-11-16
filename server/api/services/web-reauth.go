package services

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// WebReauthService re-checks a logged-in user's identity before a sensitive action, so that a
// stolen session alone is not enough to perform one.
type WebReauthService interface {
	// WebReauthVerify validates the step-up factor for the logged-in user and, on
	// success, stamps the re-auth freshness marker. Community verifies the
	// account password through [VerifySignedInPassword], after the key, so it
	// returns that function's errors besides [VerifyWebReauth]'s; the enterprise
	// overlay adds TOTP for MFA users. WebReauthVerify returns the confirmation
	// code the person carries back to the terminal, empty when the request
	// released no parked login.
	WebReauthVerify(ctx context.Context, req *requests.WebReauthVerify) (string, error)
}

func (s *service) WebReauthVerify(ctx context.Context, req *requests.WebReauthVerify) (string, error) {
	return VerifyWebReauth(ctx, s.store, req, s.clock.Now(), func(user *models.User) error {
		return VerifySignedInPassword(ctx, s.cache, user, req.Password)
	})
}

// VerifyWebReauth runs a step-up whose factor verify checks: it resolves the user, confirms the key
// is one of the user's identities, and only then calls verify, so a request naming a key the user
// does not own learns nothing about the factor. When verify accepts, it stamps the step-up as
// [StampWebReauth] does. Exported so the enterprise overlay, which checks TOTP for MFA users,
// keeps the same order.
//
// It returns [ErrUserNotFound] when the user cannot be resolved, [ErrSSHIdentityNotFound] when the
// key is not one of the user's identities, verify's error as is, and otherwise what
// [StampWebReauth] returns.
func VerifyWebReauth(ctx context.Context, st store.Store, req *requests.WebReauthVerify, now time.Time, verify func(*models.User) error) (string, error) {
	user, err := st.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil {
		return "", NewErrUserNotFound(req.UserID, err)
	}

	if err := requireOwnSSHIdentity(ctx, st, req); err != nil {
		return "", err
	}

	if err := verify(user); err != nil {
		return "", err
	}

	return stampWebReauth(ctx, st, req, now)
}

// StampWebReauth records a successful step-up on the identity that presented the
// key, and releases the login waiting on it. Exported so the enterprise overlay's
// SAML step-up, which proves the factor through the identity provider, gets both
// identically instead of duplicating them.
//
// Owning the identity is the authorization: the caller proved a factor and the
// key resolves to their own account, which is stricter than the membership the
// approval routes check. So no separate permission gate is needed here — and a
// login can only be released by the person whose key it is.
//
// It returns [ErrSSHIdentityNotFound] when the key is not one of the caller's
// identities, and [ErrSSHApprovalCodeNotFound] when the approval code is not a
// pending re-auth approval for that key in the caller's tenant. Neither is a
// forbidden error, so a 403 from the step-up only ever means a failed factor. now is the instant
// the approval's expiry is checked against.
func StampWebReauth(ctx context.Context, st store.Store, req *requests.WebReauthVerify, now time.Time) (string, error) {
	if err := requireOwnSSHIdentity(ctx, st, req); err != nil {
		return "", err
	}

	return stampWebReauth(ctx, st, req, now)
}

func requireOwnSSHIdentity(ctx context.Context, st store.Store, req *requests.WebReauthVerify) error {
	sc, err := scope.NewBounded(req.TenantID)
	if err != nil {
		return NewErrSSHIdentityNotFound(req.Fingerprint, err)
	}

	identity, err := st.SSHIdentityResolve(ctx, sc, store.SSHIdentityFingerprintResolver, req.Fingerprint)
	if err != nil || identity.PrincipalID != req.UserID {
		return NewErrSSHIdentityNotFound(req.Fingerprint, err)
	}

	return nil
}

func stampWebReauth(ctx context.Context, st store.Store, req *requests.WebReauthVerify, now time.Time) (string, error) {
	var confirmationCode string

	err := st.WithTransaction(ctx, func(ctx context.Context) error {
		if err := st.SSHIdentityTouchReauth(ctx, req.TenantID, req.Fingerprint); err != nil {
			return err
		}

		if req.ApprovalCode == "" {
			return nil
		}

		code, err := releaseSSHApproval(ctx, st, req, now)
		if err != nil {
			return err
		}

		confirmationCode = code

		return nil
	})
	if err != nil {
		return "", err
	}

	return confirmationCode, nil
}

func releaseSSHApproval(ctx context.Context, st store.Store, req *requests.WebReauthVerify, now time.Time) (string, error) {
	approval, err := st.SSHApprovalGet(ctx, pairingcode.Normalize(req.ApprovalCode), now)
	if err != nil {
		return "", NewErrSSHApprovalCodeNotFound(req.ApprovalCode, err)
	}

	mismatched := approval.Kind != models.SSHApprovalReauth ||
		approval.TenantID != req.TenantID ||
		approval.Fingerprint != req.Fingerprint
	if mismatched {
		return "", NewErrSSHApprovalCodeNotFound(req.ApprovalCode, nil)
	}

	confirmationCode := pairingcode.New(pairingcode.DeviceCodeLength)

	claimed, err := st.SSHApprovalDecide(ctx, approval.Code, models.SSHApprovalDecision{
		State:             models.SSHApprovalConfirmed,
		DecidedBy:         req.UserID,
		ConfirmationCode:  confirmationCode,
		IdentityExpiresIn: nil,
	}, now)
	if err != nil {
		return "", err
	}

	if !claimed {
		return "", NewErrSSHApprovalCodeNotFound(req.ApprovalCode, nil)
	}

	return confirmationCode, nil
}
