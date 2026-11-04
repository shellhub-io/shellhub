package agentd

import (
	"context"
	"net"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	client_mocks "github.com/shellhub-io/shellhub/pkg/api/client/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestReconnectBackoffDoublesItsWindowUpToTwoMinutes(t *testing.T) {
	var windows []time.Duration

	backoff := newReconnectBackoff()
	backoff.random = func(n int64) int64 {
		windows = append(windows, time.Duration(n))

		return n - 1
	}

	for range 7 {
		delay := backoff.next()
		assert.Equal(t, windows[len(windows)-1]-1, delay, "the delay is drawn from the window")
	}

	assert.Equal(t, []time.Duration{
		10 * time.Second,
		20 * time.Second,
		40 * time.Second,
		80 * time.Second,
		2 * time.Minute,
		2 * time.Minute,
		2 * time.Minute,
	}, windows)

	backoff.reset()
	backoff.next()

	assert.Equal(t, 10*time.Second, windows[len(windows)-1], "a reset starts again from the first window")
}

func TestReconnectBackoffDrawsEveryDelayInsideItsWindow(t *testing.T) {
	const draws = 10000

	backoff := newReconnectBackoff()

	var sum time.Duration

	for range draws {
		delay := backoff.next()
		backoff.reset()

		require.GreaterOrEqual(t, delay, time.Duration(0))
		require.Less(t, delay, reconnectBackoffBase)

		sum += delay
	}

	assert.InDelta(t, float64(reconnectBackoffBase/2), float64(sum/draws), float64(reconnectBackoffBase)/50,
		"full jitter spreads the delays evenly over the window")
}

func TestServeTunnelSpreadsTheReconnectsAfterATunnelIsLost(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cli := client_mocks.NewMockClient(t)
		cli.On("AuthDevice", mock.Anything, mock.Anything, mock.Anything).
			Return(&models.DeviceAuthResponse{UID: "uid", Token: "token", Name: "dev", Namespace: "ns"}, nil)

		ag := removalTestAgent(t, cli, "00000000-0000-4000-0000-000000000000")
		require.NoError(t, ag.Authorize(context.Background()))

		var mu sync.Mutex
		var dials []time.Time

		started := time.Now() //nolint:forbidigo // synctest fakes time.Now, which is the clock the reconnect timer reads

		dial := func(context.Context) (net.Listener, error) {
			mu.Lock()
			defer mu.Unlock()

			dials = append(dials, time.Now()) //nolint:forbidigo // synctest fakes time.Now, which is the clock the reconnect timer reads

			return newIdleListener(), nil
		}
		lose := func(context.Context, net.Listener) error { return nil }

		ctx, cancel := context.WithCancel(context.Background())

		served := make(chan error, 1)
		go func() { served <- ag.serveTunnel(ctx, dial, lose) }()

		time.Sleep(5 * time.Minute)
		synctest.Wait()
		cancel()
		require.NoError(t, <-served)

		mu.Lock()
		defer mu.Unlock()

		require.GreaterOrEqual(t, len(dials), 3)
		assert.Equal(t, started, dials[0], "the first tunnel is dialed at once")

		windows := []time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second, 80 * time.Second}
		spread := false

		for i := 1; i < len(dials); i++ {
			window := 2 * time.Minute
			if i <= len(windows) {
				window = windows[i-1]
			}

			gap := dials[i].Sub(dials[i-1])

			assert.Less(t, gap, window, "reconnect %d waited past its window", i)
			spread = spread || gap > 0
		}

		assert.True(t, spread, "a lost tunnel was redialed at once every time")
	})
}

func TestServeTunnelStartsAgainFromTheFirstWindowAfterAStableTunnel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cli := client_mocks.NewMockClient(t)
		cli.On("AuthDevice", mock.Anything, mock.Anything, mock.Anything).
			Return(&models.DeviceAuthResponse{UID: "uid", Token: "token", Name: "dev", Namespace: "ns"}, nil)

		ag := removalTestAgent(t, cli, "00000000-0000-4000-0000-000000000000")
		require.NoError(t, ag.Authorize(context.Background()))

		const (
			shortTunnels  = 4
			stableTunnels = 5
		)

		var mu sync.Mutex
		dials := 0

		dialed := func() int {
			mu.Lock()
			defer mu.Unlock()

			return dials
		}

		dial := func(context.Context) (net.Listener, error) {
			mu.Lock()
			defer mu.Unlock()

			dials++

			return newIdleListener(), nil
		}

		lost := make(chan struct{}, 1)

		serve := func(ctx context.Context, _ net.Listener) error {
			switch served := dialed(); {
			case served <= shortTunnels:
			case served <= shortTunnels+stableTunnels:
				select {
				case <-time.After(3 * time.Minute):
				case <-ctx.Done():
				}

				lost <- struct{}{}
			default:
				<-ctx.Done()
			}

			return nil
		}

		ctx, cancel := context.WithCancel(context.Background())

		served := make(chan error, 1)
		go func() { served <- ag.serveTunnel(ctx, dial, serve) }()

		for stable := 1; stable <= stableTunnels; stable++ {
			<-lost
			time.Sleep(10 * time.Second)
			synctest.Wait()

			assert.Equal(t, shortTunnels+stable+1, dialed(),
				"after stable tunnel %d, held for three minutes, the reconnect waited past the first window", stable)
		}

		cancel()
		require.NoError(t, <-served)
	})
}
