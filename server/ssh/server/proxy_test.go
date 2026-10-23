package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

const forgedClient = "198.51.100.7:40000"

func nonLoopbackIPv4(t *testing.T) net.IP {
	t.Helper()

	addrs, err := net.InterfaceAddrs()
	require.NoError(t, err)

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP
		}
	}

	t.Skip("no non-loopback IPv4 address to dial the listener from")

	return nil
}

func newTestServer(t *testing.T) *Server {
	t.Helper()

	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	block, err := gossh.MarshalPrivateKey(key, "")
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "ssh.key")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(block), 0o600))

	srv, err := NewServer(nil, nil, nil, session.NewRegistry(), &Options{HostKeyFile: path}) //nolint:exhaustruct // the listener needs only a host key
	require.NoError(t, err)
	t.Cleanup(func() { _ = srv.Close() })

	return srv
}

func listenOn(t *testing.T, ip net.IP) net.Listener {
	t.Helper()

	lis, err := new(net.ListenConfig).Listen(t.Context(), "tcp", net.JoinHostPort(ip.String(), "0"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = lis.Close() })

	return lis
}

func dialWithForgedHeader(t *testing.T, addr string, ip net.IP) net.Conn {
	t.Helper()

	conn, err := new(net.Dialer).DialContext(t.Context(), "tcp", addr)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	require.NoError(t, conn.SetDeadline(clock.New().Now().Add(5*time.Second)))

	forged, err := net.ResolveTCPAddr("tcp", forgedClient)
	require.NoError(t, err)

	_, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	_, err = fmt.Fprintf(conn, "PROXY TCP4 %s %s %d %s\r\n", forged.IP, ip, forged.Port, port)
	require.NoError(t, err)

	return conn
}

func TestProxyHeaderFromUntrustedPeerIsRefusedAndLogged(t *testing.T) {
	ip := nonLoopbackIPv4(t)

	hook := test.NewGlobal()
	t.Cleanup(hook.Reset)

	srv := newTestServer(t)
	lis := listenOn(t, ip)
	go func() { _ = srv.sshd.Serve(newProxyListener(lis, loopbackProxyPolicy)) }()

	conn := dialWithForgedHeader(t, lis.Addr().String(), ip)

	_, _, _, err := gossh.NewClientConn(conn, lis.Addr().String(), &gossh.ClientConfig{ //nolint:exhaustruct // the handshake fails before any other field is read
		User:            "user",
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec // the server closes the connection before it presents a host key
	})
	require.Error(t, err)

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		var warned bool
		for _, entry := range hook.AllEntries() {
			peer, _ := entry.Data["peer"].(string)
			warned = warned || (entry.Level == logrus.WarnLevel && strings.HasPrefix(peer, ip.String()+":"))
		}

		assert.True(tt, warned, "no warning names the peer %s", ip)
	}, 5*time.Second, 50*time.Millisecond)
}
