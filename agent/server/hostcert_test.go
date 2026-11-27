package server

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func deviceSigner(t *testing.T) gossh.Signer {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(key)
	require.NoError(t, err)

	return signer
}

func TestCertifyHostKey(t *testing.T) {
	device := deviceSigner(t)

	signer, err := certifyHostKey(device)
	require.NoError(t, err)

	certificate, ok := signer.PublicKey().(*gossh.Certificate)
	require.True(t, ok, "the host key must be a certificate")

	assert.Equal(t, uint32(gossh.HostCert), certificate.CertType)
	assert.Equal(t, gossh.KeyAlgoED25519, certificate.Key.Type())
	assert.True(t, bytes.Equal(device.PublicKey().Marshal(), certificate.SignatureKey.Marshal()), "the device key must sign the certificate")
	require.NoError(t, (&gossh.CertChecker{}).CheckCert("", certificate))
}

func hostKeyPresented(t *testing.T, s *Server, conn *GatewayLoginConn, device gossh.Signer) gossh.PublicKey { //nolint:ireturn // the test reads back what x/crypto presented
	t.Helper()

	ctx := newSSHContext("")
	s.connCallback(ctx, conn)

	config := s.serverConfig(ctx)
	config.AddHostKey(device)
	config.NoClientAuth = true

	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0") //nolint:exhaustruct // the zero ListenConfig is the default one
	require.NoError(t, err)

	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		accepted, err := listener.Accept()
		if err != nil {
			return
		}

		defer accepted.Close() //nolint:errcheck // the device side of a test connection

		_, _, _, _ = gossh.NewServerConn(accepted, config)
	}()

	dialled, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", listener.Addr().String()) //nolint:exhaustruct // the zero Dialer is the default one
	require.NoError(t, err)

	t.Cleanup(func() { _ = dialled.Close() })

	var presented gossh.PublicKey

	client, _, _, err := gossh.NewClientConn(dialled, "device", &gossh.ClientConfig{ //nolint:exhaustruct // a test client needs only the user and host key policy
		User: "root",
		HostKeyCallback: func(_ string, _ net.Addr, key gossh.PublicKey) error {
			presented = key

			return nil
		},
	})
	require.NoError(t, err)

	_ = client.Close()

	return presented
}

func TestServerConfigHostKey(t *testing.T) {
	device := deviceSigner(t)

	certified, err := certifyHostKey(device)
	require.NoError(t, err)

	s := &Server{hostCertificate: certified}

	t.Run("presents the certified key to a gateway that accepts it", func(t *testing.T) {
		presented := hostKeyPresented(t, s, &GatewayLoginConn{GatewayStream: GatewayStream{HostCertificate: true}}, device)

		certificate, ok := presented.(*gossh.Certificate)
		require.True(t, ok)
		assert.True(t, bytes.Equal(device.PublicKey().Marshal(), certificate.SignatureKey.Marshal()))
	})

	t.Run("presents the device key to a gateway that did not ask", func(t *testing.T) {
		presented := hostKeyPresented(t, s, &GatewayLoginConn{GatewayStream: GatewayStream{Secret: "secret"}}, device)

		assert.True(t, bytes.Equal(device.PublicKey().Marshal(), presented.Marshal()))
	})

	t.Run("presents the device key when no key could be certified", func(t *testing.T) {
		presented := hostKeyPresented(t, &Server{}, &GatewayLoginConn{GatewayStream: GatewayStream{HostCertificate: true}}, device)

		assert.True(t, bytes.Equal(device.PublicKey().Marshal(), presented.Marshal()))
	})
}
