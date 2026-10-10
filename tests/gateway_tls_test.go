package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const acmeDomain = "shellhub.example.com"

func upGatewayStack(t *testing.T, ctx context.Context, cfg environment.Config) (*environment.Stack, string) {
	t.Helper()

	cfg.Edition = environment.EditionCommunity
	cfg.Name = uuid.Generate()
	cfg.Run = run

	stack, err := environment.Up(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stack.Down(context.WithoutCancel(ctx))) })

	return stack, cfg.Name
}

func requireAttachedDownRemoves(t *testing.T, ctx context.Context, stack *environment.Stack, name, dir string) {
	t.Helper()

	require.DirExists(t, dir)

	attached, err := environment.Attach(ctx, name, stack.Files(), stack.Envs())
	require.NoError(t, err)
	require.NoError(t, attached.Down(ctx))

	assert.NoDirExists(t, dir)
}

func servedCertificate(ctx context.Context, address string, config *tls.Config) (*x509.Certificate, error) {
	dialer := &tls.Dialer{NetDialer: &net.Dialer{Timeout: 10 * time.Second}, Config: config}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}

	defer func() { _ = conn.Close() }()

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return nil, fmt.Errorf("dialing %s returned a %T, not a TLS connection", address, conn)
	}

	return tlsConn.ConnectionState().PeerCertificates[0], nil
}

// TestGatewaySuppliedCertificate runs the gateway with automatic HTTPS on and a certificate it is
// given, which is how TLS runs on a name no public CA signs.
func TestGatewaySuppliedCertificate(t *testing.T) {
	ctx := context.Background()

	certificate, err := environment.SelfSignedCertificate("localhost")
	require.NoError(t, err)

	block, _ := pem.Decode(certificate.CertificatePEM)
	require.NotNil(t, block)

	supplied, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	roots := x509.NewCertPool()
	roots.AddCert(supplied)

	compose, name := upGatewayStack(t, ctx, environment.Config{SuppliedCertificate: &certificate})

	t.Run("the gateway serves the supplied certificate", func(t *testing.T) {
		served, err := servedCertificate(t.Context(), compose.HTTPSAddress(), &tls.Config{ServerName: "localhost", RootCAs: roots, MinVersion: tls.VersionTLS12})
		require.NoError(t, err)
		assert.Equal(t, supplied.Raw, served.Raw)

		client := &http.Client{
			Transport:     &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}},
			CheckRedirect: withoutRedirects.CheckRedirect,
		}

		info := fetch(t, client, "https://"+compose.HTTPSAddress()+"/info", "", nil)
		assert.Equal(t, http.StatusOK, info.status, info.body)
	})

	t.Run("plain HTTP to the domain is redirected to HTTPS", func(t *testing.T) {
		plain := fetch(t, withoutRedirects, compose.BaseURL()+"/devices?page=2", "", nil)

		assert.Equal(t, http.StatusPermanentRedirect, plain.status)
		assert.Equal(t, "https://localhost/devices?page=2", plain.header.Get("Location"))
	})

	t.Run("plain HTTP for a name the gateway does not serve is answered 404", func(t *testing.T) {
		unknown := fetch(t, withoutRedirects, compose.BaseURL()+"/info", "unknown.example", nil)

		assert.Equal(t, http.StatusNotFound, unknown.status)
		assert.Empty(t, unknown.header.Get("Location"))
	})

	t.Run("tearing the stack down from another process removes the certificate it wrote", func(t *testing.T) {
		requireAttachedDownRemoves(t, ctx, compose, name, compose.Envs()["SHELLHUB_TEST_TLS_DIR"])
	})
}

// TestGatewayACME runs the gateway with automatic HTTPS on and its ACME CA pointed at a Pebble
// server in the stack, in place of Let's Encrypt's production directory.
func TestGatewayACME(t *testing.T) {
	ctx := context.Background()

	compose, name := upGatewayStack(t, ctx, environment.Config{ACME: true, Envs: map[string]string{"SHELLHUB_DOMAIN": acmeDomain}})

	t.Run("the gateway serves a certificate the overriding CA issued", func(t *testing.T) {
		var served *x509.Certificate

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			var err error

			served, err = servedCertificate(t.Context(), compose.HTTPSAddress(), &tls.Config{
				ServerName:         acmeDomain,
				InsecureSkipVerify: true, //nolint:gosec // Pebble's root is generated when it starts; the issuer is what the case reads
				MinVersion:         tls.VersionTLS12,
			})
			assert.NoError(tt, err, "the gateway holds no certificate for %s yet", acmeDomain)
		}, 2*time.Minute, 2*time.Second)

		assert.True(t, strings.HasPrefix(served.Issuer.CommonName, "Pebble Intermediate CA"), "issued by %q", served.Issuer.CommonName)
		assert.NoError(t, served.VerifyHostname(acmeDomain))
	})

	t.Run("tearing the stack down from another process removes the Pebble files it wrote", func(t *testing.T) {
		requireAttachedDownRemoves(t, ctx, compose, name, compose.Envs()["SHELLHUB_TEST_ACME_DIR"])
	})
}
