package dialer

import (
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter/wsconnadaptertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerBindBeatsAtThePingInterval(t *testing.T) {
	cases := []struct {
		description string
		configured  time.Duration
		interval    time.Duration
	}{
		{
			description: "without an interval it pings every 35 seconds",
			interval:    35 * time.Second,
		},
		{
			description: "it pings at the interval it is given",
			configured:  2 * time.Second,
			interval:    2 * time.Second,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			isolateFromYamuxTimerPool(t)

			synctest.Test(t, func(t *testing.T) {
				server, peer := wsconnadaptertest.NewPair(t)

				agent, err := yamux.Server(wsconnadapter.New(peer), nil)
				require.NoError(t, err)
				t.Cleanup(func() { _ = agent.Close() })

				var beats atomic.Int32

				m := NewManager(clock.New())
				m.PingInterval = tc.configured
				m.DialerKeepAliveCallback = func(string) { beats.Add(1) }

				require.NoError(t, m.Bind("tenant", "uid", server))

				synctest.Wait()
				require.EqualValues(t, 1, beats.Load(), "binding is the first beat")

				time.Sleep(tc.interval - time.Millisecond)
				synctest.Wait()
				assert.EqualValues(t, 1, beats.Load(), "beat before the interval")

				time.Sleep(time.Millisecond)
				synctest.Wait()
				assert.EqualValues(t, 2, beats.Load(), "no beat at the interval")
			})
		})
	}
}

func TestManagerBindClosesATunnelWhosePingGoesUnanswered(t *testing.T) {
	isolateFromYamuxTimerPool(t)

	synctest.Test(t, func(t *testing.T) {
		server, peer := wsconnadaptertest.NewPair(t)

		agentSawClose := make(chan struct{})

		go func() {
			defer close(agentSawClose)

			for {
				if _, _, err := peer.ReadMessage(); err != nil {
					return
				}
			}
		}()

		var offline atomic.Int32

		interval := time.Second

		m := NewManager(clock.New())
		m.PingInterval = interval
		m.DialerDoneCallback = func(string) { offline.Add(1) }

		require.NoError(t, m.Bind("tenant", "uid", server))

		time.Sleep(interval + connectionWriteTimeout + time.Millisecond)
		synctest.Wait()

		select {
		case <-agentSawClose:
		default:
			t.Fatal("the tunnel was left open after its ping went unanswered")
		}

		_, _, err := m.Dial(context.Background(), NewKey("tenant", "uid"))
		require.ErrorIs(t, err, ErrNoConnection)
		assert.EqualValues(t, 1, offline.Load(), "the device was not reported offline exactly once")
	})
}

func isolateFromYamuxTimerPool(t *testing.T) {
	t.Helper()

	dropYamuxPooledTimers()
	t.Cleanup(dropYamuxPooledTimers)
}

func dropYamuxPooledTimers() {
	runtime.GC()
	runtime.GC()
}
