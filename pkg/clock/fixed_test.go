package clock_test

import (
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/stretchr/testify/assert"
)

func TestFixed(t *testing.T) {
	t.Parallel()

	start := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	t.Run("returns the instant it holds", func(t *testing.T) {
		t.Parallel()

		c := clock.NewFixed(start)

		assert.Equal(t, start, c.Now())
		assert.Equal(t, start, c.Now())
	})

	t.Run("Set replaces the instant", func(t *testing.T) {
		t.Parallel()

		c := clock.NewFixed(start)
		c.Set(start.Add(time.Hour))

		assert.Equal(t, start.Add(time.Hour), c.Now())
	})

	t.Run("Advance moves the instant forward", func(t *testing.T) {
		t.Parallel()

		c := clock.NewFixed(start)
		c.Advance(time.Minute)
		c.Advance(time.Minute)

		assert.Equal(t, start.Add(2*time.Minute), c.Now())
	})

	t.Run("Advance is safe for concurrent use", func(t *testing.T) {
		t.Parallel()

		c := clock.NewFixed(start)

		var wg sync.WaitGroup
		for range 100 {
			wg.Go(func() {
				c.Advance(time.Second)
				_ = c.Now()
			})
		}
		wg.Wait()

		assert.Equal(t, start.Add(100*time.Second), c.Now())
	})

	t.Run("Set is safe for concurrent use", func(t *testing.T) {
		t.Parallel()

		c := clock.NewFixed(start)

		set := make([]time.Time, 100)
		var wg sync.WaitGroup
		for i := range set {
			set[i] = start.Add(time.Duration(i) * time.Second)
			wg.Go(func() {
				c.Set(set[i])
				_ = c.Now()
			})
		}
		wg.Wait()

		assert.Contains(t, set, c.Now())
	})
}

func TestNew(t *testing.T) {
	t.Parallel()

	before := time.Now() //nolint:forbidigo // bounds the real clock under test
	now := clock.New().Now()
	after := time.Now() //nolint:forbidigo // bounds the real clock under test

	assert.False(t, now.Before(before))
	assert.False(t, now.After(after))
}
