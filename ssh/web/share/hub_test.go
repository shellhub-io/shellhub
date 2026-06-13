package share

import (
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func drain(t *testing.T, s *subscriber) []message {
	t.Helper()

	var msgs []message
	for {
		select {
		case m, ok := <-s.out:
			if !ok {
				return msgs
			}
			msgs = append(msgs, m)
		default:
			return msgs
		}
	}
}

func TestHubBroadcastsToAllSubscribers(t *testing.T) {
	hub := newHub()

	a := hub.Subscribe()
	b := hub.Subscribe()

	hub.Output([]byte("hello"))

	for _, sub := range []*subscriber{a, b} {
		msgs := drain(t, sub)
		require.Len(t, msgs, 1)
		assert.Equal(t, websocket.BinaryMessage, msgs[0].typ)
		assert.Equal(t, []byte("hello"), msgs[0].data)
	}
}

func TestHubLateJoinerReceivesRingAndResize(t *testing.T) {
	hub := newHub()

	hub.Resize(Dimensions{Cols: 120, Rows: 40})
	hub.Output([]byte("scrollback"))

	late := hub.Subscribe()
	msgs := drain(t, late)

	require.Len(t, msgs, 2)
	assert.Equal(t, websocket.TextMessage, msgs[0].typ)
	assert.Equal(t, websocket.BinaryMessage, msgs[1].typ)
	assert.Equal(t, []byte("scrollback"), msgs[1].data)
}

func TestHubRingResetsOnScreenClear(t *testing.T) {
	hub := newHub()

	hub.Output([]byte("stale content from before the clear"))
	hub.Output([]byte("\x1b[2Jcurrent screen"))

	late := hub.Subscribe()
	msgs := drain(t, late)

	// The snapshot must start at the clear, dropping the stale pre-clear content.
	require.Len(t, msgs, 1)
	assert.Equal(t, websocket.BinaryMessage, msgs[0].typ)
	assert.Equal(t, []byte("\x1b[2Jcurrent screen"), msgs[0].data)
}

func TestHubDropsSlowConsumer(t *testing.T) {
	hub := newHub()

	slow := hub.Subscribe()

	// Overflow the bounded buffer; the slow consumer must be dropped, never blocking the producer.
	for i := 0; i < subscriberBuffer+10; i++ {
		hub.Output([]byte("x"))
	}

	hub.mu.Lock()
	_, present := hub.subscribers[slow]
	hub.mu.Unlock()

	assert.False(t, present, "slow consumer should have been dropped")

	_, open := <-slow.out
	// Channel is closed for a dropped consumer; eventually a receive returns !ok.
	for open {
		_, open = <-slow.out
	}
}

func TestHubCloseClosesSubscribers(t *testing.T) {
	hub := newHub()

	sub := hub.Subscribe()
	hub.Close()

	// Done channel is closed.
	select {
	case <-hub.Done():
	default:
		t.Fatal("expected Done to be closed")
	}

	// Drain any buffered frames, then the channel must be closed.
	open := true
	for open {
		_, open = <-sub.out
	}
}
