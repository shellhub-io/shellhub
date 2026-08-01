package session

import (
	"io"
	"sync"
	"testing"

	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type fakeChannel struct{}

func (*fakeChannel) Read([]byte) (int, error)    { return 0, io.EOF }
func (*fakeChannel) Write(p []byte) (int, error) { return len(p), nil }
func (*fakeChannel) Close() error                { return nil }
func (*fakeChannel) CloseWrite() error           { return nil }
func (*fakeChannel) Stderr() io.ReadWriter       { return nil }

func (*fakeChannel) SendRequest(_ string, _ bool, _ []byte) (bool, error) {
	return true, nil
}

type fakeNewChannel struct{}

func (*fakeNewChannel) Accept() (gossh.Channel, <-chan *gossh.Request, error) {
	requests := make(chan *gossh.Request)
	close(requests)

	return &fakeChannel{}, requests, nil
}

func (*fakeNewChannel) Reject(gossh.RejectionReason, string) error { return nil }

func (*fakeNewChannel) ChannelType() string { return "session" }
func (*fakeNewChannel) ExtraData() []byte   { return nil }

// TestNewClientChannelConcurrent covers the crash a multiplexed connection could
// trigger: channel handlers run one goroutine per channel open, so unguarded
// writes to the seat map are a runtime fatal error that recover cannot contain.
// Without the mutex this fails under -race, and often dies outright.
func TestNewClientChannelConcurrent(t *testing.T) {
	const seats = 64

	sess := newTestSession(servicemocks.NewMockService(t))

	wg := new(sync.WaitGroup)
	wg.Add(seats)

	for seat := range seats {
		go func() {
			defer wg.Done()

			channel, err := sess.NewClientChannel(&fakeNewChannel{}, seat)
			assert.NoError(t, err)
			assert.NotNil(t, channel)
		}()
	}

	wg.Wait()

	assert.Len(t, sess.Client.Channels, seats)

	for seat := range seats {
		assert.NotNil(t, sess.Client.Channels[seat], "seat %d lost its channel", seat)
	}
}

func TestNewClientChannelRejectsSeatTwice(t *testing.T) {
	sess := newTestSession(servicemocks.NewMockService(t))

	_, err := sess.NewClientChannel(&fakeNewChannel{}, 0)
	require.NoError(t, err)

	_, err = sess.NewClientChannel(&fakeNewChannel{}, 0)
	assert.ErrorIs(t, err, ErrSeatAlreadySet)
}

// TestDropAgentChannelConcurrent exercises the removal path against the same map
// under contention, as the channel handler uses it to roll back a seat whose
// client side failed to open.
func TestDropAgentChannelConcurrent(t *testing.T) {
	const seats = 32

	sess := newTestSession(servicemocks.NewMockService(t))

	for seat := range seats {
		sess.Agent.Channels[seat] = &AgentChannel{Channel: &fakeChannel{}, Requests: nil}
	}

	wg := new(sync.WaitGroup)
	wg.Add(seats)

	for seat := range seats {
		go func() {
			defer wg.Done()

			sess.DropAgentChannel(seat)
		}()
	}

	wg.Wait()

	assert.Empty(t, sess.Agent.Channels)
}

// TestSeatsConcurrentAccess covers the seat fields themselves: they are written
// from the client-request goroutine and read from the pipe goroutine, so the
// setters mutate in place and Get has to hand back a copy.
func TestSeatsConcurrentAccess(t *testing.T) {
	const seats = 32

	sess := newTestSession(servicemocks.NewMockService(t))

	ids := make([]int, 0, seats)

	for range seats {
		id, err := sess.Seats.NewSeat()
		require.NoError(t, err)

		ids = append(ids, id)
	}

	wg := new(sync.WaitGroup)

	for _, id := range ids {
		wg.Add(3)

		go func() {
			defer wg.Done()

			sess.Seats.SetPty(id, true)
		}()

		go func() {
			defer wg.Done()

			sess.Seats.SetType(id, "exec")
		}()

		go func() {
			defer wg.Done()

			sess.Seats.Get(id)
		}()
	}

	wg.Wait()

	for _, id := range ids {
		seat, ok := sess.Seats.Get(id)
		assert.True(t, ok)
		assert.True(t, seat.HasPty)
		assert.Equal(t, "exec", seat.Type)
	}
}

// TestSeatsGetReturnsCopy pins the reason Get does not hand out the stored
// pointer: a caller holding it would observe later mutations without any
// synchronization.
func TestSeatsGetReturnsCopy(t *testing.T) {
	sess := newTestSession(servicemocks.NewMockService(t))

	id, err := sess.Seats.NewSeat()
	require.NoError(t, err)

	seat, ok := sess.Seats.Get(id)
	require.True(t, ok)
	require.False(t, seat.HasPty)

	sess.Seats.SetPty(id, true)

	assert.False(t, seat.HasPty)

	seat, ok = sess.Seats.Get(id)
	require.True(t, ok)
	assert.True(t, seat.HasPty)
}
