package services

import (
	"context"
	"fmt"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// UserRegisteredHookFn is called after a not-confirmed user is created during registration. Cloud
// uses it to send the email-verification link. It runs outside any transaction and its failure is
// non-fatal (registration succeeds; the user can request a resend).
type UserRegisteredHookFn func(ctx context.Context, user *models.User, forwardedHost string, validUntil time.Time) error

var userRegisteredHooks []UserRegisteredHookFn

// OnUserRegistered registers a hook that fires after a not-confirmed user is created.
// It must be called during package init, before the server starts handling requests.
func OnUserRegistered(fn UserRegisteredHookFn) {
	if fn == nil {
		panic("services: OnUserRegistered called with nil hook")
	}

	userRegisteredHooks = append(userRegisteredHooks, fn)
}

// fireUserRegistered dispatches all registered post-registration hooks sequentially.
func fireUserRegistered(ctx context.Context, user *models.User, forwardedHost string, validUntil time.Time) error {
	for _, fn := range userRegisteredHooks {
		if err := fn(ctx, user, forwardedHost, validUntil); err != nil {
			return fmt.Errorf("user registered hook failed: %w", err)
		}
	}

	return nil
}

// openSignupEnabled reports whether open self-registration is allowed. Cloud turns this on at init;
// community and enterprise are invite-only (an uninvited RegisterUser is refused).
var openSignupEnabled bool

// EnableOpenSignup turns on open self-registration. It must be called during package init.
func EnableOpenSignup() {
	openSignupEnabled = true
}

// openSignupAllowed reports whether open self-registration is on.
func openSignupAllowed() bool {
	return openSignupEnabled
}
