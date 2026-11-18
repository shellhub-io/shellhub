package services

import (
	"context"
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// EmailCheckHookFn is called before an address the user typed is written to their account, on
// open sign-up and on a change of email or recovery email. It returns the reason it refuses the
// address, which becomes the field's message in the 400, or "" to accept it. It runs on the
// request's context, so a hook that does I/O must bound it itself; it has no way to fail other
// than accepting or refusing.
type EmailCheckHookFn func(ctx context.Context, email string) string

func (h *Hooks) checkEmail(ctx context.Context, email string) string {
	for _, fn := range h.EmailCheck {
		if reason := fn(ctx, email); reason != "" {
			return reason
		}
	}

	return ""
}

// EmailField names one of a user's two addresses.
type EmailField string

const (
	// EmailFieldEmail is the address the user signs in with and receives mail at.
	EmailFieldEmail EmailField = "email"
	// EmailFieldRecoveryEmail is the address that receives one of the two MFA reset codes.
	EmailFieldRecoveryEmail EmailField = "recovery_email"
)

// EmailChangeRequestedHookFn is called after a user asked, with their current password, to change
// the address field names to email, and the rest of the change is saved. user is the account as
// saved, still holding its old addresses. The hook owns proving that whoever asked controls email,
// and applies it through [UserService.ConfirmEmailChange] once they do. It is called once for each
// address the request changes, on the request's context, and an error it returns fails the
// request.
type EmailChangeRequestedHookFn func(ctx context.Context, user *models.User, field EmailField, email string) error

func (h *Hooks) emailChangeConfirmationWired() bool {
	return len(h.EmailChangeRequested) > 0
}

func (h *Hooks) emailChangeRequested(ctx context.Context, user *models.User, field EmailField, email string) error {
	for _, fn := range h.EmailChangeRequested {
		if err := fn(ctx, user, field, email); err != nil {
			return fmt.Errorf("email change requested hook failed: %w", err)
		}
	}

	return nil
}

// EmailsChange is a change to a user's email or recovery email, as the request that made it saw it.
type EmailsChange struct {
	// Before is the account before the request.
	Before *models.User
	// After is the account as the request leaves it, except that After.Email and
	// After.RecoveryEmail are the requested addresses when EmailHeld and RecoveryEmailHeld.
	After *models.User
	// EmailHeld reports that After.Email waits for its owner's confirmation and is not set yet.
	EmailHeld bool
	// RecoveryEmailHeld reports that After.RecoveryEmail waits for its owner's confirmation and is
	// not set yet.
	RecoveryEmailHeld bool
}

// UserEmailsChangedHookFn is called once a request that changes a user's email or recovery email
// is saved, a held address included, and once [UserService.NotifyEmailsChanged] reports a change
// saved elsewhere, so the owner hears of the change when it is asked for. It runs on the request's
// context and has no way to fail the request, which is already saved: a hook that sends mail logs
// its own failure.
type UserEmailsChangedHookFn func(ctx context.Context, change EmailsChange)

func (h *Hooks) userEmailsChanged(ctx context.Context, change EmailsChange) {
	if change.Before.Email == change.After.Email && change.Before.RecoveryEmail == change.After.RecoveryEmail {
		return
	}

	for _, fn := range h.UserEmailsChanged {
		fn(ctx, change)
	}
}

// MFACodeCheckHookFn checks code against the second factor of user, who has MFA enabled, before
// [UserService.UpdateUser] changes their password, email or recovery email. It returns the reason
// it refuses the code, which becomes the mfa_code field's message in the 400, or "" to accept it,
// and an error only when it cannot check the code at all, which fails the request. It runs on the
// request's context.
type MFACodeCheckHookFn func(ctx context.Context, user *models.User, code string) (string, error)

func (h *Hooks) mfaCodeRequired(user *models.User) bool {
	return len(h.MFACodeCheck) > 0 && user.MFA.Enabled
}

func (h *Hooks) checkMFACode(ctx context.Context, user *models.User, code string) error {
	if !h.mfaCodeRequired(user) {
		return nil
	}

	if code == "" {
		return NewErrInvalidFields(ErrBadRequest, map[string]string{mfaCodeField: "is required while MFA is enabled"})
	}

	for _, fn := range h.MFACodeCheck {
		reason, err := fn(ctx, user, code)
		if err != nil {
			return NewErrUserUpdate(user, fmt.Errorf("mfa code check hook failed: %w", err))
		}

		if reason != "" {
			return NewErrInvalidFields(ErrBadRequest, map[string]string{mfaCodeField: reason})
		}
	}

	return nil
}

const mfaCodeField = "mfa_code"
