package session

import (
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"net"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer/dialertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	gossh "golang.org/x/crypto/ssh"
)

const forwardWait = 5 * time.Second

func verifyNoLeakedGoroutines(t *testing.T) {
	t.Helper()

	running := goleak.IgnoreCurrent()

	t.Cleanup(func() { goleak.VerifyNone(t, running) })
}

func laptopOverLoopback(t *testing.T) (*gossh.ServerConn, <-chan gossh.NewChannel) {
	t.Helper()

	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(key)
	require.NoError(t, err)

	serverConfig := &gossh.ServerConfig{NoClientAuth: true} //nolint:exhaustruct // a test server needs only a host key and no client auth
	serverConfig.AddHostKey(signer)

	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0") //nolint:exhaustruct // the zero ListenConfig is the default one
	require.NoError(t, err)

	t.Cleanup(func() { _ = listener.Close() })

	type handshake struct {
		conn *gossh.ServerConn
		err  error
	}

	accepted := make(chan handshake, 1)

	go func() {
		serverSide, err := listener.Accept()
		if err != nil {
			accepted <- handshake{err: err} //nolint:exhaustruct // a failed accept has no connection

			return
		}

		conn, chans, requests, err := gossh.NewServerConn(serverSide, serverConfig)
		if err == nil {
			go gossh.DiscardRequests(requests)
			go func() {
				for newChannel := range chans {
					_ = newChannel.Reject(gossh.Prohibited, "the laptop's server end takes no channels")
				}
			}()
		}

		accepted <- handshake{conn: conn, err: err}
	}()

	laptopSide, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", listener.Addr().String()) //nolint:exhaustruct // the zero Dialer is the default one
	require.NoError(t, err)

	laptopConfig := &gossh.ClientConfig{User: "user", HostKeyCallback: gossh.InsecureIgnoreHostKey()} //nolint:exhaustruct,gosec // a loopback client to a NoClientAuth server needs only a user, and the server's host key is throwaway
	laptop, chans, requests, err := gossh.NewClientConn(laptopSide, listener.Addr().String(), laptopConfig)
	require.NoError(t, err)

	go gossh.DiscardRequests(requests)

	server := <-accepted
	require.NoError(t, server.err)

	t.Cleanup(func() {
		_ = laptop.Close()
		_ = server.conn.Close()
	})

	return server.conn, chans
}

func copyBothWays(device, client gossh.Channel) {
	done := make(chan struct{}, 2)

	go func() {
		_, _ = io.Copy(client, device)
		_ = client.CloseWrite()
		done <- struct{}{}
	}()

	go func() {
		_, _ = io.Copy(device, client)
		_ = device.CloseWrite()
		done <- struct{}{}
	}()

	<-done
	<-done
}

func nextForward(t *testing.T, laptop <-chan gossh.NewChannel) gossh.NewChannel {
	t.Helper()

	select {
	case newChannel, ok := <-laptop:
		require.True(t, ok, "the laptop's connection closed before a forward reached it")
		require.Equal(t, AuthRequestOpenSSHChannel, newChannel.ChannelType())

		return newChannel
	case <-time.After(forwardWait):
		t.Fatal("no agent channel reached the laptop")

		return nil
	}
}

func acceptForward(t *testing.T, laptop <-chan gossh.NewChannel) gossh.Channel {
	t.Helper()

	channel, requests, err := nextForward(t, laptop).Accept()
	require.NoError(t, err)

	go gossh.DiscardRequests(requests)

	return channel
}

func openFromDevice(t *testing.T, agent *dialertest.Agent) func() gossh.Channel {
	t.Helper()

	type opened struct {
		channel gossh.Channel
		err     error
	}

	result := make(chan opened, 1)

	go func() {
		channel, err := agent.OpenChannel(AuthRequestOpenSSHChannel)
		result <- opened{channel: channel, err: err}
	}()

	return func() gossh.Channel {
		t.Helper()

		select {
		case got := <-result:
			require.NoError(t, got.err)

			return got.channel
		case <-time.After(forwardWait):
			t.Fatal("the device's agent channel was never answered")

			return nil
		}
	}
}

func readWithin(t *testing.T, r io.Reader, size int) string {
	t.Helper()

	type result struct {
		data []byte
		err  error
	}

	read := make(chan result, 1)

	go func() {
		data := make([]byte, size)
		n, err := io.ReadFull(r, data)
		read <- result{data: data[:n], err: err}
	}()

	select {
	case got := <-read:
		require.NoError(t, got.err)

		return string(got.data)
	case <-time.After(forwardWait):
		t.Fatal("nothing arrived through the forward")

		return ""
	}
}

func TestForwardAgentRelaysAnAgentUseAfterRepeatedRequests(t *testing.T) {
	verifyNoLeakedGoroutines(t)

	agent := dialertest.NewAgent(t)
	sess, _ := newConnectedSession(t, agent)
	client, laptop := laptopOverLoopback(t)

	sess.ForwardAgent(client, copyBothWays)
	sess.ForwardAgent(client, copyBothWays)

	opened := openFromDevice(t, agent)
	forwarded := acceptForward(t, laptop)
	device := opened()

	_, err := device.Write([]byte("sign"))
	require.NoError(t, err)
	assert.Equal(t, "sign", readWithin(t, forwarded, len("sign")))

	_, err = forwarded.Write([]byte("signature"))
	require.NoError(t, err)
	assert.Equal(t, "signature", readWithin(t, device, len("signature")))

	require.NoError(t, device.Close())
	require.NoError(t, sess.agent.Close())
}

func TestForwardAgentRelaysAgentUsesAtTheSameTime(t *testing.T) {
	verifyNoLeakedGoroutines(t)

	agent := dialertest.NewAgent(t)
	sess, _ := newConnectedSession(t, agent)
	client, laptop := laptopOverLoopback(t)

	sess.ForwardAgent(client, copyBothWays)

	firstOpened := openFromDevice(t, agent)
	firstForwarded := acceptForward(t, laptop)
	first := firstOpened()

	secondOpened := openFromDevice(t, agent)
	secondForwarded := acceptForward(t, laptop)
	second := secondOpened()

	_, err := second.Write([]byte("two"))
	require.NoError(t, err)
	assert.Equal(t, "two", readWithin(t, secondForwarded, len("two")))

	_, err = first.Write([]byte("one"))
	require.NoError(t, err)
	assert.Equal(t, "one", readWithin(t, firstForwarded, len("one")))

	require.NoError(t, sess.agent.Close())
}

func requireEnded(t *testing.T, channel gossh.Channel, end string) {
	t.Helper()

	ended := make(chan error, 1)

	go func() {
		_, err := io.ReadAll(channel)
		ended <- err
	}()

	select {
	case err := <-ended:
		require.NoError(t, err)
	case <-time.After(forwardWait):
		t.Fatalf("the %s end of a finished agent use is still open", end)
	}
}

func TestForwardAgentClosesBothEndsWhenAnAgentUseEnds(t *testing.T) {
	verifyNoLeakedGoroutines(t)

	agent := dialertest.NewAgent(t)
	sess, _ := newConnectedSession(t, agent)
	client, laptop := laptopOverLoopback(t)

	sess.ForwardAgent(client, func(gossh.Channel, gossh.Channel) {})

	opened := openFromDevice(t, agent)
	forwarded := acceptForward(t, laptop)
	device := opened()

	requireEnded(t, device, "device")
	requireEnded(t, forwarded, "laptop")

	require.NoError(t, sess.agent.Close())
}

func TestForwardAgentKeepsRelayingAfterTheClientRefusesAnAgentUse(t *testing.T) {
	verifyNoLeakedGoroutines(t)

	agent := dialertest.NewAgent(t)
	sess, _ := newConnectedSession(t, agent)
	client, laptop := laptopOverLoopback(t)

	sess.ForwardAgent(client, copyBothWays)

	refused := make(chan error, 1)

	go func() {
		_, err := agent.OpenChannel(AuthRequestOpenSSHChannel)
		refused <- err
	}()

	require.NoError(t, nextForward(t, laptop).Reject(gossh.Prohibited, "no agent here"))

	select {
	case err := <-refused:
		var rejection *gossh.OpenChannelError
		require.ErrorAs(t, err, &rejection, "the device must learn its agent use was refused")
	case <-time.After(forwardWait):
		t.Fatal("the device's refused agent channel was never answered")
	}

	opened := openFromDevice(t, agent)
	forwarded := acceptForward(t, laptop)
	device := opened()

	_, err := device.Write([]byte("sign"))
	require.NoError(t, err)
	assert.Equal(t, "sign", readWithin(t, forwarded, len("sign")))

	require.NoError(t, sess.agent.Close())
}
