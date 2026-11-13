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

var emailCheckHooks []EmailCheckHookFn

// OnEmailCheck registers a hook that may refuse an address before it is written. It must be called
// during package init, before the server starts handling requests. With no hook registered every
// address is accepted. It panics when fn is nil.
func OnEmailCheck(fn EmailCheckHookFn) {
	if fn == nil {
		panic("services: OnEmailCheck called with nil hook")
	}

	emailCheckHooks = append(emailCheckHooks, fn)
}

func checkEmail(ctx context.Context, email string) string {
	for _, fn := range emailCheckHooks {
		if reason := fn(ctx, email); reason != "" {
			return reason
		}
	}

	return ""
}

// EmailChangeRequestedHookFn is called after a user asked, with their current password, to change
// their email to email, and the rest of the change is saved. user is the account as saved, still
// holding its old email. The hook owns proving that whoever asked controls email, and applies it
// through [UserService.ConfirmEmailChange] once they do. It runs on the request's context, and an
// error it returns fails the request.
type EmailChangeRequestedHookFn func(ctx context.Context, user *models.User, email string) error

var emailChangeRequestedHooks []EmailChangeRequestedHookFn

// OnEmailChangeRequested registers a hook that confirms a changed email before it takes effect.
// Registering one declares that the instance can mail the new address, so from then on
// [UserService.UpdateUser] never writes a new email itself. It must be called during package init,
// before the server starts handling requests, and panics when fn is nil.
func OnEmailChangeRequested(fn EmailChangeRequestedHookFn) {
	if fn == nil {
		panic("services: OnEmailChangeRequested called with nil hook")
	}

	emailChangeRequestedHooks = append(emailChangeRequestedHooks, fn)
}

func emailChangeConfirmationWired() bool {
	return len(emailChangeRequestedHooks) > 0
}

func fireEmailChangeRequested(ctx context.Context, user *models.User, email string) error {
	for _, fn := range emailChangeRequestedHooks {
		if err := fn(ctx, user, email); err != nil {
			return fmt.Errorf("email change requested hook failed: %w", err)
		}
	}

	return nil
}

// EmailsChange is a change to a user's email or recovery email, as the request that made it saw it.
type EmailsChange struct {
	// Before is the account before the request.
	Before *models.User
	// After is the account as the request leaves it, except that After.Email is the requested
	// address when EmailHeld.
	After *models.User
	// EmailHeld reports that After.Email waits for its owner's confirmation and is not set yet.
	EmailHeld bool
}

// UserEmailsChangedHookFn is called once a request that changes a user's email or recovery email
// is saved, a held email included, so the owner hears of the change when it is asked for. It runs
// on the request's context and has no way to fail the request, which is already saved: a hook that
// sends mail logs its own failure.
type UserEmailsChangedHookFn func(ctx context.Context, change EmailsChange)

var userEmailsChangedHooks []UserEmailsChangedHookFn

// OnUserEmailsChanged registers a hook that learns of every request that changes a user's email or
// recovery email, when it is saved. It must be called during package init, before the server starts handling
// requests, and panics when fn is nil.
func OnUserEmailsChanged(fn UserEmailsChangedHookFn) {
	if fn == nil {
		panic("services: OnUserEmailsChanged called with nil hook")
	}

	userEmailsChangedHooks = append(userEmailsChangedHooks, fn)
}

func fireUserEmailsChanged(ctx context.Context, change EmailsChange) {
	if change.Before.Email == change.After.Email && change.Before.RecoveryEmail == change.After.RecoveryEmail {
		return
	}

	for _, fn := range userEmailsChangedHooks {
		fn(ctx, change)
	}
}
