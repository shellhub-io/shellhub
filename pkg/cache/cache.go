package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error

	// HasAccountLockout reports whether the source is currently blocked from attempting to
	// log in to a user with the specified userID. It returns the absolute Unix timestamp
	// in seconds representing the end of the lockout, or 0 if no lockout was found; the
	// attempt number and an error if any.
	HasAccountLockout(ctx context.Context, source, userID string) (lockout int64, attempt int, err error)

	// StoreLoginAttempt stores a login attempt from source to the user with the specified userID.
	// If the attempt number equals or exceeds 3, it sets a lockout for future login attempts.
	//
	// The lockout duration is calculated based on the number of attempts made, increasing exponentially
	// by a factor of 4 after the third attempt. Attempts must last for half of the double lockout duration.
	//
	// This means that a user who was locked out for 4 minutes must have the attempts stored for 10
	// minutes (or 6 minutes after the timeout). Any wrong attempt within this time will increase the
	// lockout once again. After this, the attempts will be reset, and new wrong attempts will start the
	// attempt counter from 0.
	//
	// The following equations are used to calculate both lockout and attempt duration, with 'x' representing
	// the lockout duration and 'y' the attempt duration:
	//
	//	F(x) = min(4^(a - 3), M)
	//	F(y) = min(x * 2.5, M)
	//
	// Where:
	//
	//	x is the lockout duration in minutes.
	//	y is the attempt duration in minutes.
	//	a is the attempt number.
	//	M is the maximum duration value, specified by the "SHELLHUB_MAXIMUM_ACCOUNT_LOCKOUT" environment variable.
	//
	// Examples for M = 32768 (15 days) and a = n:
	//
	//	n    = 3 | 4  | 5  | 8    | 11
	//	_________________________________
	//	F(x) = 1 | 4  | 16 | 1024 | 32768
	//	F(y) = 3 | 10 | 40 | 2560 | 32768
	//
	// It returns the absolute Unix timestamp in seconds representing the end of the lockout, or 0 if no
	// lockout was found; the attempt number and an error if any.
	StoreLoginAttempt(ctx context.Context, source, userID string) (lockout int64, attempt int, err error)

	// ResetLoginAttempts resets the login attempts and associated lockout from the source to
	// the user with the specified userID.
	ResetLoginAttempts(ctx context.Context, source, userID string) error
}
