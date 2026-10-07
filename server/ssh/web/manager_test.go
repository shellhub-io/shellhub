package web

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerSave(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description string
		id          string
		waitFor     time.Duration
		data        *Credentials
	}{
		{
			description: "insert credential on manager and delete after 1 second",
			id:          "foo",
			waitFor:     1 * time.Second,
			data:        nil,
		},
		{
			description: "insert credential on manager and delete after 2 seconds",
			id:          "bar",
			waitFor:     2 * time.Second,
			data:        nil,
		},
	}

	for _, ts := range tests {
		test := ts

		t.Run(test.description, func(t *testing.T) {
			t.Parallel()

			manager := newManager(test.waitFor)
			manager.save(test.id, nil)

			time.Sleep(2 * test.waitFor)

			_, ok := manager.get(test.id)
			assert.False(t, ok)
		})
	}
}

func TestManagerGetRedeemsTheTokenOnce(t *testing.T) {
	t.Parallel()

	manager := newManager(time.Minute)
	manager.save("token", &Credentials{Device: "device"})

	creds, ok := manager.get("token")
	require.True(t, ok)
	assert.Equal(t, "device", creds.Device)

	_, ok = manager.get("token")
	assert.False(t, ok)
}

func TestManagerConcurrentGetsYieldOneWinner(t *testing.T) {
	t.Parallel()

	const redeemers = 50

	manager := newManager(time.Minute)
	manager.save("token", &Credentials{Device: "device"})

	var (
		wg      sync.WaitGroup
		winners atomic.Int32
	)

	start := make(chan struct{})
	for range redeemers {
		wg.Go(func() {
			<-start

			if _, ok := manager.get("token"); ok {
				winners.Add(1)
			}
		})
	}

	close(start)
	wg.Wait()

	assert.Equal(t, int32(1), winners.Load())
}
