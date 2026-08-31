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
type UserRegisteredHookFn func(ctx context.Context, user *models.User, forwardedHost, forwardedProto string, validUntil time.Time) error

var userRegisteredHooks []UserRegisteredHookFn

// OnUserRegistered registers a hook that fires after a not-confirmed user is created.
// It must be called during package init, before the server starts handling requests.
func OnUserRegistered(fn UserRegisteredHookFn) {
	if fn == nil {
		panic("services: OnUserRegistered called with nil hook")
	}

	userRegisteredHooks = append(userRegisteredHooks, fn)
}

func fireUserRegistered(ctx context.Context, user *models.User, forwardedHost, forwardedProto string, validUntil time.Time) error {
	for _, fn := range userRegisteredHooks {
		if err := fn(ctx, user, forwardedHost, forwardedProto, validUntil); err != nil {
			return fmt.Errorf("user registered hook failed: %w", err)
		}
	}

	return nil
}

var openSignupEnabled bool

// EnableOpenSignup turns on open self-registration. It must be called during package init.
func EnableOpenSignup() {
	openSignupEnabled = true
}

func openSignupAllowed() bool {
	return openSignupEnabled
}
