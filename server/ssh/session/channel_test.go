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

type fakeNewChannel struct {
	channel gossh.Channel
}

func (f *fakeNewChannel) Accept() (gossh.Channel, <-chan *gossh.Request, error) {
	requests := make(chan *gossh.Request)
	close(requests)

	if f.channel != nil {
		return f.channel, requests, nil
	}

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

	sess := newTestSession(servicemocks.NewMockService(t), nil)

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

	assert.Len(t, sess.client.channels, seats)

	for seat := range seats {
		assert.NotNil(t, sess.client.channels[seat], "seat %d lost its channel", seat)
	}
}

func TestNewClientChannelRejectsSeatTwice(t *testing.T) {
	sess := newTestSession(servicemocks.NewMockService(t), nil)

	_, err := sess.NewClientChannel(&fakeNewChannel{}, 0)
	require.NoError(t, err)

	_, err = sess.NewClientChannel(&fakeNewChannel{}, 0)
	assert.ErrorIs(t, err, ErrSeatAlreadySet)
}

func countSeats(sess *Session) int {
	count := 0

	sess.seats.items.Range(func(_, _ any) bool {
		count++

		return true
	})

	return count
}

func occupySeat(t *testing.T, sess *Session) (int, *recordingChannel, *recordingChannel) {
	t.Helper()

	seat, err := sess.NewSeat()
	require.NoError(t, err)

	agent := new(recordingChannel)
	sess.agent.channels[seat] = &AgentChannel{Channel: agent, Requests: nil}

	client := new(recordingChannel)
	_, err = sess.NewClientChannel(&fakeNewChannel{channel: client}, seat)
	require.NoError(t, err)

	return seat, agent, client
}

func assertSeatKept(t *testing.T, sess *Session, seat int, agent, client *recordingChannel) {
	t.Helper()

	assert.False(t, agent.closed)
	assert.False(t, client.closed)
	assert.Contains(t, sess.agent.channels, seat)
	assert.Contains(t, sess.client.channels, seat)

	_, ok := sess.Seat(seat)
	assert.True(t, ok)
	assert.Equal(t, 1, sess.client.open)
}

func TestReleaseSeatForgetsTheSeat(t *testing.T) {
	const seats = 100

	sess := newTestSession(servicemocks.NewMockService(t), nil)

	for range seats {
		seat, agent, client := occupySeat(t, sess)

		sess.ReleaseSeat(seat)

		assert.True(t, agent.closed, "seat %d kept its agent channel open", seat)
		assert.True(t, client.closed, "seat %d kept its client channel open", seat)

		_, ok := sess.Seat(seat)
		assert.False(t, ok, "seat %d is still known", seat)
		require.ErrorIs(t, sess.CloseAgentWrite(seat), ErrSeatNotFound)
	}

	assert.Empty(t, sess.agent.channels)
	assert.Empty(t, sess.client.channels)
	assert.Zero(t, countSeats(sess))
	assert.Zero(t, sess.client.open)
}

func TestReleaseSeatLeavesOtherSeatsAlone(t *testing.T) {
	sess := newTestSession(servicemocks.NewMockService(t), nil)

	released, _, _ := occupySeat(t, sess)
	kept, agent, client := occupySeat(t, sess)

	sess.ReleaseSeat(released)

	assertSeatKept(t, sess, kept, agent, client)
}

func TestReleaseSeatIgnoresAReleasedOrUnknownSeat(t *testing.T) {
	sess := newTestSession(servicemocks.NewMockService(t), nil)

	released, _, _ := occupySeat(t, sess)
	kept, agent, client := occupySeat(t, sess)

	sess.ReleaseSeat(released)
	sess.ReleaseSeat(released)
	sess.ReleaseSeat(404)

	assertSeatKept(t, sess, kept, agent, client)
}

func TestReleaseSeatConcurrent(t *testing.T) {
	const seats = 32

	sess := newTestSession(servicemocks.NewMockService(t), nil)

	ids := make([]int, 0, seats)

	for range seats {
		seat, _, _ := occupySeat(t, sess)
		ids = append(ids, seat)
	}

	wg := new(sync.WaitGroup)

	for _, seat := range ids {
		wg.Go(func() {
			sess.ReleaseSeat(seat)
		})
	}

	wg.Wait()

	assert.Empty(t, sess.agent.channels)
	assert.Empty(t, sess.client.channels)
	assert.Zero(t, countSeats(sess))
}

// TestSeatsConcurrentAccess covers the seat fields themselves: they are written
// from the client-request goroutine and read from the pipe goroutine, so the
// setters mutate in place and Get has to hand back a copy.
func TestSeatsConcurrentAccess(t *testing.T) {
	const seats = 32

	sess := newTestSession(servicemocks.NewMockService(t), nil)

	ids := make([]int, 0, seats)

	for range seats {
		id, err := sess.seats.NewSeat()
		require.NoError(t, err)

		ids = append(ids, id)
	}

	wg := new(sync.WaitGroup)

	for _, id := range ids {
		wg.Add(3)

		go func() {
			defer wg.Done()

			sess.seats.SetPty(id, true)
		}()

		go func() {
			defer wg.Done()

			sess.seats.SetType(id, "exec")
		}()

		go func() {
			defer wg.Done()

			sess.seats.Get(id)
		}()
	}

	wg.Wait()

	for _, id := range ids {
		seat, ok := sess.seats.Get(id)
		assert.True(t, ok)
		assert.True(t, seat.HasPty)
		assert.Equal(t, "exec", seat.Type)
	}
}

// TestSeatsGetReturnsCopy pins the reason Get does not hand out the stored
// pointer: a caller holding it would observe later mutations without any
// synchronization.
func TestSeatsGetReturnsCopy(t *testing.T) {
	sess := newTestSession(servicemocks.NewMockService(t), nil)

	id, err := sess.seats.NewSeat()
	require.NoError(t, err)

	seat, ok := sess.seats.Get(id)
	require.True(t, ok)
	require.False(t, seat.HasPty)

	sess.seats.SetPty(id, true)

	assert.False(t, seat.HasPty)

	seat, ok = sess.seats.Get(id)
	require.True(t, ok)
	assert.True(t, seat.HasPty)
}
