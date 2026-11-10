package server

import (
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/shellonopen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type openedChannel struct {
	gossh.NewChannel

	extra    []byte
	requests chan *gossh.Request
}

func (c *openedChannel) ExtraData() []byte { return c.extra }

func (c *openedChannel) Accept() (gossh.Channel, <-chan *gossh.Request, error) { //nolint:ireturn // gossh.Channel is the interface x/crypto hands out
	return nil, c.requests, nil
}

func acceptedRequests(t *testing.T, extra []byte, sent ...string) []*gossh.Request {
	t.Helper()

	requests := make(chan *gossh.Request, len(sent))
	for _, name := range sent {
		requests <- &gossh.Request{Type: name}
	}

	close(requests)

	var got []*gossh.Request

	handler := shellOnOpenHandler(func(_ *gliderssh.Server, _ *gossh.ServerConn, newChan gossh.NewChannel, _ gliderssh.Context) {
		_, reqs, err := newChan.Accept()
		require.NoError(t, err)

		for req := range reqs {
			got = append(got, req)
		}
	})

	handler(nil, nil, &openedChannel{extra: extra, requests: requests}, nil)

	return got
}

func requestTypes(requests []*gossh.Request) []string {
	types := make([]string, 0, len(requests))
	for _, req := range requests {
		types = append(types, req.Type)
	}

	return types
}

type shellOnOpenData struct {
	Marker   string
	Requests []byte `ssh:"rest"`
}

type openedRequest struct {
	Type    string
	Payload []byte
}

func openData(marker string, requests ...openedRequest) []byte {
	var list []byte
	for _, req := range requests {
		list = append(list, gossh.Marshal(openedRequest{Type: req.Type, Payload: req.Payload})...)
	}

	return gossh.Marshal(shellOnOpenData{Marker: marker, Requests: list})
}

func TestShellOnOpenHandler(t *testing.T) {
	pty := []byte("the pty request")
	env := []byte("the env request")

	t.Run("serves the requests the open data carries ahead of the client's", func(t *testing.T) {
		data := openData(shellonopen.Marker, openedRequest{Type: "env", Payload: env}, openedRequest{Type: "pty-req", Payload: pty}, openedRequest{Type: "shell"})

		got := acceptedRequests(t, data, "window-change")

		assert.Equal(t, []string{"env", "pty-req", "shell", "window-change"}, requestTypes(got))
		assert.Equal(t, env, got[0].Payload)
		assert.Equal(t, pty, got[1].Payload)

		for _, req := range got[:3] {
			assert.False(t, req.WantReply)
		}
	})

	t.Run("rejects open data that does not end in a shell", func(t *testing.T) {
		assert.False(t, rejected(t, openData(shellonopen.Marker, openedRequest{Type: "pty-req", Payload: pty}, openedRequest{Type: "shell"})))
		assert.True(t, rejected(t, openData(shellonopen.Marker, openedRequest{Type: "pty-req", Payload: pty})))
		assert.True(t, rejected(t, openData(shellonopen.Marker)))
	})

	t.Run("rejects open data that runs anything but a session setup before the shell", func(t *testing.T) {
		assert.True(t, rejected(t, openData(shellonopen.Marker, openedRequest{Type: "exec", Payload: []byte("id")}, openedRequest{Type: "shell"})))
		assert.True(t, rejected(t, openData(shellonopen.Marker, openedRequest{Type: "shell"}, openedRequest{Type: "shell"})))
	})

	t.Run("rejects open data whose requests do not parse", func(t *testing.T) {
		assert.True(t, rejected(t, gossh.Marshal(shellOnOpenData{Marker: shellonopen.Marker, Requests: []byte{0, 0, 0, 9, 's'}})))
	})

	t.Run("serves a channel opened without the marker as it comes", func(t *testing.T) {
		got := acceptedRequests(t, openData("another", openedRequest{Type: "shell"}), "pty-req", "shell")

		assert.Equal(t, []string{"pty-req", "shell"}, requestTypes(got))
	})

	t.Run("serves a channel opened without data as it comes", func(t *testing.T) {
		got := acceptedRequests(t, nil, "exec")

		assert.Equal(t, []string{"exec"}, requestTypes(got))
	})
}

type rejectingChannel struct {
	openedChannel

	rejected bool
}

func (c *rejectingChannel) Reject(gossh.RejectionReason, string) error {
	c.rejected = true

	return nil
}

func rejected(t *testing.T, extra []byte) bool {
	t.Helper()

	channel := &rejectingChannel{openedChannel: openedChannel{extra: extra, requests: make(chan *gossh.Request)}}
	close(channel.requests)

	shellOnOpenHandler(func(_ *gliderssh.Server, _ *gossh.ServerConn, newChan gossh.NewChannel, _ gliderssh.Context) {
		_, reqs, err := newChan.Accept()
		require.NoError(t, err)

		for req := range reqs {
			assert.NotNil(t, req)
		}
	})(nil, nil, channel, nil)

	return channel.rejected
}
