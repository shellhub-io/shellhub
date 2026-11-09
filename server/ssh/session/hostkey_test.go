package session

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func certificate(t *testing.T, authority gossh.Signer, edit func(*gossh.Certificate)) *gossh.Certificate {
	t.Helper()

	cert := &gossh.Certificate{ //nolint:exhaustruct // a certificate needs only its key, type and validity here
		Key:         ed25519Signer(t).PublicKey(),
		CertType:    gossh.HostCert,
		ValidBefore: gossh.CertTimeInfinity,
	}

	if edit != nil {
		edit(cert)
	}

	require.NoError(t, cert.SignCert(rand.Reader, authority))

	return cert
}

func TestDeviceHostKey(t *testing.T) {
	device := ed25519Signer(t)
	check := deviceHostKey(device.PublicKey())

	t.Run("accepts the device key", func(t *testing.T) {
		assert.NoError(t, check("device", nil, device.PublicKey()))
	})

	t.Run("refuses another key", func(t *testing.T) {
		assert.ErrorIs(t, check("device", nil, ed25519Signer(t).PublicKey()), ErrHostKey)
	})

	t.Run("accepts a host certificate the device key signed", func(t *testing.T) {
		assert.NoError(t, check("device", nil, certificate(t, device, nil)))
	})

	t.Run("refuses a host certificate another key signed", func(t *testing.T) {
		err := check("device", nil, certificate(t, ed25519Signer(t), nil))

		require.ErrorIs(t, err, ErrHostKey)
		assert.ErrorIs(t, err, errCertifiedByAnotherDevice)
	})

	t.Run("refuses a user certificate the device key signed", func(t *testing.T) {
		err := check("device", nil, certificate(t, device, func(c *gossh.Certificate) { c.CertType = gossh.UserCert }))

		require.ErrorIs(t, err, ErrHostKey)
		assert.ErrorIs(t, err, errNotAHostCertificate)
	})

	t.Run("refuses an expired certificate", func(t *testing.T) {
		expired := certificate(t, device, func(c *gossh.Certificate) { c.ValidBefore = 1 })

		assert.ErrorIs(t, check("device", nil, expired), ErrHostKey)
	})

	t.Run("refuses a certificate whose key was swapped after signing", func(t *testing.T) {
		forged := certificate(t, device, nil)
		forged.Key = ed25519Signer(t).PublicKey()

		assert.ErrorIs(t, check("device", nil, forged), ErrHostKey)
	})
}
