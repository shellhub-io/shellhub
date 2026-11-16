package cache

import (
	"context"
	"errors"
	"time"
)

// ErrGetNotFound is returned when a key is absent. A miss is an error here rather than a zero
// value, so a caller cannot mistake "nothing cached" for "cached nothing".
var ErrGetNotFound = errors.New("failed to find the value on the cache")

// Get reads a key and decodes it into T, saving the caller a variable to pass by address. It
// returns ErrGetNotFound when the key is absent.
func Get[T any](ctx context.Context, cache Cache, key string) (*T, error) {
	var t *T

	if err := cache.Get(ctx, key, &t); err != nil {
		return nil, err
	}

	if t == nil {
		return nil, ErrGetNotFound
	}

	return t, nil
}

// ErrTTLTooShort is returned by [Cache.IncrBy] for a ttl under one millisecond, which Redis would
// take as an order to delete the key it just counted.
var ErrTTLTooShort = errors.New("cache: ttl must be at least one millisecond")

// Cache is the key/value store the services cache in. Implementations must be safe for concurrent
// use; NewNullCache is the no-op one, which is what an instance without Redis runs.
type Cache interface {
	Get(ctx context.Context, key string, value any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error

	// SetNX atomically sets key to value with the given ttl only if the key does
	// not already exist, and reports whether it was set. It is the building block
	// for single-use reservations, such as a one-time token that must be redeemed
	// exactly once under concurrent requests, which Get+Set cannot do race-free.
	SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error)

	// CompareAndDelete atomically deletes key only while it still holds value, and reports whether
	// it did. It releases a reservation taken with SetNX under a value unique to its owner, so an
	// owner whose reservation expired cannot delete the one another caller took after it.
	CompareAndDelete(ctx context.Context, key, value string) (bool, error)

	// IncrBy atomically adds delta to the integer at key, counting from 0 when the key is missing,
	// sets the key to expire after ttl, and returns the new value. A delta of 0 reads the value
	// through the same call that changes it, so a caller that must not show which it did makes the
	// same round trip either way. It returns ErrTTLTooShort, touching nothing, for a ttl under one
	// millisecond, and the cache's error, leaving the key and its expiry unchanged, when the key
	// holds something other than an integer or the sum would overflow int64. Keep the key out of
	// Get and Set, which store values encoded.
	IncrBy(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error)

	// HasAccountLockout reports whether the source is currently blocked from attempting to
	// log in to a user with the specified userID. It returns the absolute Unix timestamp
	// in seconds representing the end of the lockout while the cache's clock is before it, or 0
	// once that second is reached or no lockout was found; the attempt number and an error
	// if any.
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
	//	F(x) = min(B * 4^(a - 3), M)
	//	F(y) = max(min(x * 2.5, M), 2 * B)
	//
	// Where:
	//
	//	x is the lockout duration.
	//	y is the attempt duration.
	//	a is the attempt number.
	//	B is the first lockout's duration, specified by the "SHELLHUB_ACCOUNT_LOCKOUT_BASE" environment variable.
	//	M is the maximum duration value, specified by the "SHELLHUB_MAXIMUM_ACCOUNT_LOCKOUT" environment variable.
	//
	// Examples, in minutes, for B = 1 minute, M = 32768 (15 days) and a = n:
	//
	//	n    = 3   | 4  | 5  | 8    | 11
	//	__________________________________
	//	F(x) = 1   | 4  | 16 | 1024 | 32768
	//	F(y) = 2.5 | 10 | 40 | 2560 | 32768
	//
	// It returns the absolute Unix timestamp in seconds representing the end of the lockout, or 0 if no
	// lockout was found; the attempt number and an error if any.
	StoreLoginAttempt(ctx context.Context, source, userID string) (lockout int64, attempt int, err error)

	// ResetLoginAttempts resets the login attempts and associated lockout from the source to
	// the user with the specified userID.
	ResetLoginAttempts(ctx context.Context, source, userID string) error
}
