package web

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/server/ssh/pkg/webhandoff"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func testHostKey(t *testing.T) gossh.PublicKey {
	t.Helper()

	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	key, err := gossh.NewPublicKey(pub)
	require.NoError(t, err)

	return key
}

func TestDialDeviceStopsWhenItsContextEnds(t *testing.T) {
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	accepted := make(chan struct{})

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		if _, err := conn.Read(make([]byte, 1)); err == nil {
			close(accepted)
		}

		<-t.Context().Done()
		_ = conn.Close()
	}()

	ctx, cancel := context.WithCancel(t.Context())

	creds := &Credentials{Username: "user", Password: "password", Device: "device"} //nolint:exhaustruct // a password login names nothing else
	hostKey := testHostKey(t)
	done := make(chan error, 1)

	go func() {
		_, err := dialDevice(ctx, nil, webhandoff.NewStore(), nil, creds, Info{IP: "127.0.0.1"}, hostKey, listener.Addr().String(), log.NewEntry(log.StandardLogger()))
		done <- err
	}()

	<-accepted
	cancel()

	select {
	case err := <-done:
		require.ErrorIs(t, err, ErrAuthentication)
	case <-time.After(5 * time.Second):
		t.Fatal("dialDevice kept waiting on a gateway that never answers after its context ended")
	}
}

func TestDialDeviceReportsAGatewayItCannotReach(t *testing.T) {
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	address := listener.Addr().String()
	require.NoError(t, listener.Close())

	creds := &Credentials{Username: "user", Password: "password", Device: "device"} //nolint:exhaustruct // a password login names nothing else

	_, err = dialDevice(t.Context(), nil, webhandoff.NewStore(), nil, creds, Info{IP: "127.0.0.1"}, testHostKey(t), address, log.NewEntry(log.StandardLogger()))
	require.ErrorIs(t, err, ErrConnect)
}
