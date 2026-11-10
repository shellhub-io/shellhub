package session

import (
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type closeCountingChannel struct {
	closes atomic.Int32
}

func (*closeCountingChannel) Read([]byte) (int, error)    { return 0, io.EOF }
func (*closeCountingChannel) Write(p []byte) (int, error) { return len(p), nil }
func (*closeCountingChannel) CloseWrite() error           { return nil }
func (*closeCountingChannel) Stderr() io.ReadWriter       { return nil }

func (c *closeCountingChannel) Close() error {
	c.closes.Add(1)

	return nil
}

func (*closeCountingChannel) SendRequest(string, bool, []byte) (bool, error) { return true, nil }

func TestLateChannel(t *testing.T) {
	t.Run("keeps taking the device's requests after the handler closed it", func(t *testing.T) {
		late := newLateChannel()
		requests := make(chan *gossh.Request)

		late.resolve(&closeCountingChannel{}, requests, nil, nil)
		require.NoError(t, late.Close())

		for range 3 {
			select {
			case requests <- &gossh.Request{Type: "exit-status"}:
			case <-time.After(5 * time.Second):
				t.Fatal("a request from the device blocked on a channel nobody reads")
			}
		}

		close(requests)
	})

	t.Run("closes a channel that resolves after it was closed", func(t *testing.T) {
		late := newLateChannel()
		channel := &closeCountingChannel{}

		require.NoError(t, late.Close())

		requests := make(chan *gossh.Request)
		close(requests)
		late.resolve(channel, requests, nil, nil)

		assert.Equal(t, int32(1), channel.closes.Load())
	})
}

func TestLateChannelReplace(t *testing.T) {
	t.Run("hands on the new channel's requests and refuses the old one's", func(t *testing.T) {
		late := newLateChannel()
		old := make(chan *gossh.Request)
		replacement := make(chan *gossh.Request)
		previous := &closeCountingChannel{}

		late.resolve(previous, old, nil, nil)
		require.True(t, late.replace(&closeCountingChannel{}, replacement, []byte("pty")))

		assert.Equal(t, int32(1), previous.closes.Load(), "the replaced channel must be closed")
		assert.Equal(t, []byte("pty"), late.startedShell())

		select {
		case old <- &gossh.Request{Type: "exit-status"}:
		case <-time.After(5 * time.Second):
			t.Fatal("a request from the replaced channel blocked")
		}

		close(old)

		go func() { replacement <- &gossh.Request{Type: "eow@openssh.com"} }()

		select {
		case req, ok := <-late.requests:
			require.True(t, ok, "the requests must stay open while the new channel serves")
			assert.Equal(t, "eow@openssh.com", req.Type)
		case <-time.After(5 * time.Second):
			t.Fatal("the new channel's request never arrived")
		}

		close(replacement)

		select {
		case _, ok := <-late.requests:
			assert.False(t, ok, "the requests end with the channel that replaced the old one")
		case <-time.After(5 * time.Second):
			t.Fatal("the requests did not end with the new channel")
		}
	})

	t.Run("closes the replacement of a channel already closed", func(t *testing.T) {
		late := newLateChannel()
		late.resolve(&closeCountingChannel{}, make(chan *gossh.Request), nil, nil)
		require.NoError(t, late.Close())

		replacement := &closeCountingChannel{}
		assert.False(t, late.replace(replacement, make(chan *gossh.Request), nil))
		assert.Equal(t, int32(1), replacement.closes.Load())
	})

	t.Run("closes the replacement of a channel whose device side ended", func(t *testing.T) {
		late := newLateChannel()
		requests := make(chan *gossh.Request)
		late.resolve(&closeCountingChannel{}, requests, nil, nil)
		close(requests)

		_, ok := <-late.requests
		require.False(t, ok)

		replacement := &closeCountingChannel{}
		assert.False(t, late.replace(replacement, make(chan *gossh.Request), nil))
		assert.Equal(t, int32(1), replacement.closes.Load())
	})
}
