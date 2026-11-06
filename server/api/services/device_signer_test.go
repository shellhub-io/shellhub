package services

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

type mintedSigner struct {
	service *APIService
	clock   *clock.Fixed
	key     *rsa.PrivateKey
	request requests.PublicKeyAuth
}

func mintSignerFor(t *testing.T, device models.UID) mintedSigner {
	t.Helper()

	at := clock.NewFixed(time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC))
	signers := newExpiringMemoryCache(at)
	s := NewService(storemock.NewMockStore(t), privateKey, publicKey, signers, WithClock(at))

	key, err := s.MintDeviceSigner(t.Context(), device)
	require.NoError(t, err)

	public, err := ssh.NewPublicKey(&key.PublicKey)
	require.NoError(t, err)

	return mintedSigner{
		service: s,
		clock:   at,
		key:     key,
		request: requests.PublicKeyAuth{Fingerprint: ssh.FingerprintLegacyMD5(public), Data: `{"Username":"root","Namespace":"device-a"}`},
	}
}

func TestDeviceSigner(t *testing.T) {
	t.Run("signs for the device the key was minted for", func(t *testing.T) {
		minted := mintSignerFor(t, "device-a")

		res, err := minted.service.AuthPublicKey(t.Context(), "device-a", minted.request)
		require.NoError(t, err)

		signature, err := base64.StdEncoding.DecodeString(res.Signature)
		require.NoError(t, err)

		digest := sha256.Sum256([]byte(minted.request.Data))
		require.NoError(t, rsa.VerifyPKCS1v15(&minted.key.PublicKey, crypto.SHA256, digest[:], signature))
	})

	t.Run("refuses any other device", func(t *testing.T) {
		minted := mintSignerFor(t, "device-a")

		_, err := minted.service.AuthPublicKey(t.Context(), "device-b", minted.request)
		require.ErrorIs(t, err, ErrPublicKeyNotFound)
	})

	t.Run("refuses a caller that is no device", func(t *testing.T) {
		minted := mintSignerFor(t, "device-a")

		_, err := minted.service.AuthPublicKey(t.Context(), "", minted.request)
		require.ErrorIs(t, err, ErrPublicKeyNotFound)
	})

	t.Run("refuses a key the server did not mint", func(t *testing.T) {
		minted := mintSignerFor(t, "device-a")

		other := requests.PublicKeyAuth{Fingerprint: "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff", Data: minted.request.Data}
		_, err := minted.service.AuthPublicKey(t.Context(), "device-a", other)
		require.ErrorIs(t, err, ErrPublicKeyNotFound)
	})

	t.Run("signs until the minute is up", func(t *testing.T) {
		minted := mintSignerFor(t, "device-a")
		minted.clock.Advance(time.Minute - time.Nanosecond)

		_, err := minted.service.AuthPublicKey(t.Context(), "device-a", minted.request)
		require.NoError(t, err)
	})

	t.Run("refuses the key once the minute is up", func(t *testing.T) {
		minted := mintSignerFor(t, "device-a")
		minted.clock.Advance(time.Minute)

		_, err := minted.service.AuthPublicKey(t.Context(), "device-a", minted.request)
		require.ErrorIs(t, err, ErrPublicKeyNotFound)
	})
}
