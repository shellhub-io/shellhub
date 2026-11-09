package session

import (
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type openedOnDevice struct {
	data   []byte
	closed chan struct{}
}

func deviceRecordingOpens(t *testing.T) (*gossh.Client, <-chan openedOnDevice) {
	t.Helper()

	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(key)
	require.NoError(t, err)

	config := &gossh.ServerConfig{NoClientAuth: true} //nolint:exhaustruct // a test device authenticates nobody
	config.AddHostKey(signer)

	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0") //nolint:exhaustruct // the zero ListenConfig is the default one
	require.NoError(t, err)

	t.Cleanup(func() { _ = listener.Close() })

	opens := make(chan openedOnDevice, 8)

	go func() {
		deviceSide, err := listener.Accept()
		if err != nil {
			return
		}

		t.Cleanup(func() { _ = deviceSide.Close() })

		_, chans, reqs, err := gossh.NewServerConn(deviceSide, config)
		if err != nil {
			return
		}

		go gossh.DiscardRequests(reqs)

		for newChannel := range chans {
			channel, requests, err := newChannel.Accept()
			if err != nil {
				return
			}

			opened := openedOnDevice{data: newChannel.ExtraData(), closed: make(chan struct{})}
			opens <- opened

			go func() {
				gossh.DiscardRequests(requests)
				close(opened.closed)
				_ = channel.Close()
			}()
		}
	}()

	gatewaySide, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", listener.Addr().String()) //nolint:exhaustruct // the zero Dialer is the default one
	require.NoError(t, err)

	t.Cleanup(func() { _ = gatewaySide.Close() })

	conn, chans, reqs, err := gossh.NewClientConn(gatewaySide, "device", &gossh.ClientConfig{ //nolint:exhaustruct // a test client needs only the user and host key policy
		User:            "user",
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec // an in-memory test device
	})
	require.NoError(t, err)

	return gossh.NewClient(conn, chans, reqs), opens
}

func sessionOnDevice(t *testing.T, startsShell bool) (*Session, <-chan openedOnDevice) {
	t.Helper()

	client, opens := deviceRecordingOpens(t)

	return &Session{agent: &Agent{client: client, channels: make(map[int]*AgentChannel), startsShell: startsShell}}, opens //nolint:exhaustruct // only the device side matters here
}

func nextOpen(t *testing.T, opens <-chan openedOnDevice) openedOnDevice {
	t.Helper()

	select {
	case opened := <-opens:
		return opened
	case <-time.After(5 * time.Second):
		t.Fatal("the device saw no channel open")

		return openedOnDevice{}
	}
}

func awaitClosed(t *testing.T, opened openedOnDevice) {
	t.Helper()

	select {
	case <-opened.closed:
	case <-time.After(5 * time.Second):
		t.Fatal("the replaced channel was not closed on the device")
	}
}

func TestSessionStartShell(t *testing.T) {
	pty := &gossh.Request{Type: ptyRequestType, Payload: []byte("the pty")}
	env := &gossh.Request{Type: "env", Payload: []byte("LANG")}

	t.Run("opens the shell with the setup on a device that starts shells on open", func(t *testing.T) {
		sess, opens := sessionOnDevice(t, true)

		channel, err := sess.NewAgentChannel(firstSeat)
		require.NoError(t, err)

		plain := nextOpen(t, opens)
		assert.Empty(t, plain.data)

		assert.True(t, sess.CanStartShell(firstSeat))
		require.True(t, sess.StartShell(firstSeat, []*gossh.Request{env, pty}))
		assert.False(t, sess.CanStartShell(firstSeat), "a channel whose shell started cannot start another")

		assert.Equal(t, []byte(
			"\x00\x00\x00\x19shell-on-open@shellhub.io"+
				"\x00\x00\x00\x03env\x00\x00\x00\x04LANG"+
				"\x00\x00\x00\x07pty-req\x00\x00\x00\x07the pty"+
				"\x00\x00\x00\x05shell\x00\x00\x00\x00",
		), nextOpen(t, opens).data)
		awaitClosed(t, plain)

		started, ok := channel.StartedShell()
		assert.True(t, ok)
		assert.Equal(t, pty.Payload, started)
	})

	t.Run("keeps the channel on a device that does not start shells on open", func(t *testing.T) {
		sess, opens := sessionOnDevice(t, false)

		channel, err := sess.NewAgentChannel(firstSeat)
		require.NoError(t, err)

		nextOpen(t, opens)

		assert.False(t, sess.CanStartShell(firstSeat))
		assert.False(t, sess.StartShell(firstSeat, []*gossh.Request{pty}))

		_, ok := channel.StartedShell()
		assert.False(t, ok)
	})

	t.Run("keeps the channel for a setup that is not a session setup with a pty", func(t *testing.T) {
		sess, opens := sessionOnDevice(t, true)

		_, err := sess.NewAgentChannel(firstSeat)
		require.NoError(t, err)

		nextOpen(t, opens)

		assert.False(t, sess.StartShell(firstSeat, []*gossh.Request{env}))
		assert.False(t, sess.StartShell(firstSeat, []*gossh.Request{{Type: "x11-req"}, pty}))

		select {
		case opened := <-opens:
			t.Fatalf("the device saw another channel open: %q", opened.data)
		case <-time.After(50 * time.Millisecond):
		}
	})

	t.Run("reopens a plain channel in place of the started shell", func(t *testing.T) {
		sess, opens := sessionOnDevice(t, true)

		channel, err := sess.NewAgentChannel(firstSeat)
		require.NoError(t, err)

		nextOpen(t, opens)

		require.True(t, sess.StartShell(firstSeat, []*gossh.Request{pty}))

		shell := nextOpen(t, opens)

		require.NoError(t, sess.ReopenAgentChannel(firstSeat))

		assert.Empty(t, nextOpen(t, opens).data)
		awaitClosed(t, shell)

		assert.True(t, sess.CanStartShell(firstSeat), "a reopened channel runs nothing yet")

		_, ok := channel.StartedShell()
		assert.False(t, ok)
	})

	t.Run("does not reopen a seat without a session channel", func(t *testing.T) {
		sess, _ := sessionOnDevice(t, true)

		require.ErrorIs(t, sess.ReopenAgentChannel(firstSeat), ErrSeatNotFound)
		assert.False(t, sess.CanStartShell(firstSeat))
		assert.False(t, sess.StartShell(firstSeat, []*gossh.Request{pty}))
	})
}
