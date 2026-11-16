package services

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

const signedInPasswordAttemptSource = "signed-in-password"

const passwordGuessInFlightTTL = 10 * time.Second

func passwordGuessInFlightKey(userID string) string {
	return "password-guess-in-flight={" + userID + "}"
}

// VerifySignedInPassword checks password against user's for a caller already signed in as user,
// before an action that asks for the password again. Wrong passwords count toward a lockout keyed
// on the user alone, separate from the sign-in one so a stolen session cannot keep the owner from
// signing in; the third starts it. A password sent while the lockout holds, or while another one
// for the user is being checked, is refused unchecked. A right password clears the count. An empty
// password is no guess: it is refused without taking the guard or counting toward the lockout, so a
// form that sends none cannot lock its owner out.
//
// It returns [ErrUserPasswordNotMatch] for a wrong or empty password, [ErrFactorLockedOut] carrying
// [ErrDataLockout] when the password is refused unchecked or the wrong guess starts the lockout,
// and [ErrFactorCheckUnavailable] when the cache cannot admit the guess. A cache failure while
// reading, counting or clearing attempts is logged and the check goes on. Every cache call runs
// under ctx, except the release of the guess, which [cache.AdmitOne] runs past its cancellation.
func VerifySignedInPassword(ctx context.Context, c cache.Cache, user *models.User, password string) error {
	if password == "" {
		return NewErrUserPasswordNotMatch(nil)
	}

	release, admitted, err := cache.AdmitOne(ctx, c, passwordGuessInFlightKey(user.ID), passwordGuessInFlightTTL)
	if err != nil {
		return NewErrFactorCheckUnavailable(err)
	}

	if !admitted {
		return NewErrFactorLockedOut(0)
	}

	defer release()

	logger := log.WithField("user_id", user.ID)

	lockout, _, err := c.HasAccountLockout(ctx, signedInPasswordAttemptSource, user.ID)
	if err != nil {
		logger.WithError(err).Warn("unable to check the password lockout")
	}

	if lockout > 0 {
		return NewErrFactorLockedOut(lockout)
	}

	if !user.Password.Compare(password) {
		lockout, _, err := c.StoreLoginAttempt(ctx, signedInPasswordAttemptSource, user.ID)
		if err != nil {
			logger.WithError(err).Warn("unable to store a wrong password")
		}

		if lockout > 0 {
			return NewErrFactorLockedOut(lockout)
		}

		return NewErrUserPasswordNotMatch(nil)
	}

	if err := c.ResetLoginAttempts(ctx, signedInPasswordAttemptSource, user.ID); err != nil {
		logger.WithError(err).Warn("unable to clear the password attempts")
	}

	return nil
}
