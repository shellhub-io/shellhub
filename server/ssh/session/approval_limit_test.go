package session

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApprovalLimiter(t *testing.T) {
	t.Run("grants up to the per-source limit", func(t *testing.T) {
		limiter := newApprovalLimiter(2, 10)

		_, ok := limiter.acquire("10.0.0.1")
		require.True(t, ok)

		_, ok = limiter.acquire("10.0.0.1")
		require.True(t, ok)

		_, ok = limiter.acquire("10.0.0.1")
		assert.False(t, ok, "a third wait from the same address should be refused")
	})

	t.Run("one address does not starve another", func(t *testing.T) {
		limiter := newApprovalLimiter(1, 10)

		_, ok := limiter.acquire("10.0.0.1")
		require.True(t, ok)

		_, ok = limiter.acquire("10.0.0.2")
		assert.True(t, ok)
	})

	t.Run("releasing frees the slot", func(t *testing.T) {
		limiter := newApprovalLimiter(1, 10)

		release, ok := limiter.acquire("10.0.0.1")
		require.True(t, ok)

		_, ok = limiter.acquire("10.0.0.1")
		require.False(t, ok)

		release()

		_, ok = limiter.acquire("10.0.0.1")
		assert.True(t, ok)
	})

	t.Run("releasing twice does not hand out a slot that was never taken", func(t *testing.T) {
		limiter := newApprovalLimiter(1, 10)

		release, ok := limiter.acquire("10.0.0.1")
		require.True(t, ok)

		release()
		release()

		_, ok = limiter.acquire("10.0.0.1")
		require.True(t, ok)

		_, ok = limiter.acquire("10.0.0.1")
		assert.False(t, ok)
	})

	t.Run("the total bounds what many addresses can do", func(t *testing.T) {
		limiter := newApprovalLimiter(1, 2)

		_, ok := limiter.acquire("10.0.0.1")
		require.True(t, ok)

		_, ok = limiter.acquire("10.0.0.2")
		require.True(t, ok)

		_, ok = limiter.acquire("10.0.0.3")
		assert.False(t, ok)
	})

	// The web terminal bridge dials the gateway over loopback, so every browser
	// session in the instance shares that address.
	t.Run("loopback is exempt", func(t *testing.T) {
		limiter := newApprovalLimiter(1, 1)

		for range 8 {
			_, ok := limiter.acquire("127.0.0.1")
			require.True(t, ok)
		}

		_, ok := limiter.acquire("::1")
		assert.True(t, ok)
	})

	t.Run("stays consistent under contention", func(t *testing.T) {
		limiter := newApprovalLimiter(4, 64)

		wg := new(sync.WaitGroup)
		wg.Add(64)

		for range 64 {
			go func() {
				defer wg.Done()

				if release, ok := limiter.acquire("10.0.0.1"); ok {
					release()
				}
			}()
		}

		wg.Wait()

		limiter.mu.Lock()
		defer limiter.mu.Unlock()

		assert.Zero(t, limiter.current)
		assert.Empty(t, limiter.inFlight)
	})
}
