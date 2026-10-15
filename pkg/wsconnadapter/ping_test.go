package wsconnadapter_test

import (
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter/wsconnadaptertest"
	"github.com/stretchr/testify/assert"
)

func TestPingIntervalPacesThePing(t *testing.T) {
	cases := []struct {
		description string
		options     []wsconnadapter.Option
		interval    time.Duration
	}{
		{
			description: "without an interval it pings every 30 seconds",
			interval:    30 * time.Second,
		},
		{
			description: "it pings at the interval it is given",
			options:     []wsconnadapter.Option{wsconnadapter.WithPingInterval(2 * time.Second)},
			interval:    2 * time.Second,
		},
		{
			description: "a zero interval keeps the default",
			options:     []wsconnadapter.Option{wsconnadapter.WithPingInterval(0)},
			interval:    30 * time.Second,
		},
		{
			description: "a negative interval keeps the default",
			options:     []wsconnadapter.Option{wsconnadapter.WithPingInterval(-time.Second)},
			interval:    30 * time.Second,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				adapter, peer := wsconnadaptertest.NewPair(t, tc.options...)

				var pings atomic.Int32

				peer.SetPingHandler(func(string) error {
					pings.Add(1)

					return nil
				})

				go drainPeer(peer)

				adapter.Ping()

				time.Sleep(tc.interval - time.Millisecond)
				synctest.Wait()
				assert.Zero(t, pings.Load(), "pinged before the interval")

				time.Sleep(time.Millisecond)
				synctest.Wait()
				assert.EqualValues(t, 1, pings.Load(), "did not ping at the interval")
			})
		})
	}
}

func TestPongTimeoutFollowsThePingInterval(t *testing.T) {
	cases := []struct {
		description string
		options     []wsconnadapter.Option
		timeout     time.Duration
	}{
		{
			description: "without an interval a silent peer is dropped after 35 seconds",
			timeout:     35 * time.Second,
		},
		{
			description: "a short interval drops a silent peer five seconds after it",
			options:     []wsconnadapter.Option{wsconnadapter.WithPingInterval(2 * time.Second)},
			timeout:     7 * time.Second,
		},
		{
			description: "a long interval waits five seconds past it before dropping a silent peer",
			options:     []wsconnadapter.Option{wsconnadapter.WithPingInterval(time.Minute)},
			timeout:     65 * time.Second,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				adapter, peer := wsconnadaptertest.NewPair(t, tc.options...)

				peer.SetPingHandler(func(string) error { return nil })

				go drainPeer(peer)
				closed := drainAdapter(adapter)

				adapter.Ping()

				time.Sleep(tc.timeout - time.Millisecond)
				synctest.Wait()
				assert.False(t, isClosed(closed), "dropped before the pong timeout")

				time.Sleep(time.Millisecond)
				synctest.Wait()
				assert.True(t, isClosed(closed), "still open after the pong timeout")
			})
		})
	}

	t.Run("a peer that answers a long interval is kept", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			adapter, peer := wsconnadaptertest.NewPair(t, wsconnadapter.WithPingInterval(time.Minute))

			go drainPeer(peer)
			closed := drainAdapter(adapter)

			adapter.Ping()

			time.Sleep(3*time.Minute + time.Second)
			synctest.Wait()
			assert.False(t, isClosed(closed), "dropped a peer that answered every ping")
		})
	})
}

func drainPeer(peer *websocket.Conn) {
	for {
		if _, _, err := peer.NextReader(); err != nil {
			return
		}
	}
}

func drainAdapter(adapter *wsconnadapter.Adapter) <-chan struct{} {
	closed := make(chan struct{})

	go func() {
		defer close(closed)

		buffer := make([]byte, 1024)

		for {
			if _, err := adapter.Read(buffer); err != nil {
				return
			}
		}
	}()

	return closed
}

func isClosed(closed <-chan struct{}) bool {
	select {
	case <-closed:
		return true
	default:
		return false
	}
}
