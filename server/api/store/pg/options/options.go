package options

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/uptrace/bun"
)

// Store is what an option prepares at startup: the database handle and the clock the store reads.
type Store struct {
	DB    *bun.DB
	Clock clock.Clock
}

// Option prepares the store at startup. Returning an error aborts startup.
type Option func(ctx context.Context, store *Store) error

// WithClock makes the store read the time from c instead of the wall clock.
func WithClock(c clock.Clock) Option {
	return func(_ context.Context, store *Store) error {
		store.Clock = c

		return nil
	}
}
