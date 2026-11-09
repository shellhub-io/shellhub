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

		late.resolve(&closeCountingChannel{}, requests, nil)
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
		late.resolve(channel, requests, nil)

		assert.Equal(t, int32(1), channel.closes.Load())
	})
}
