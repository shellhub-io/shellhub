package services

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
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

type recordingGenerator struct {
	mu        sync.Mutex
	pool      []*rsa.PrivateKey
	generated int
}

func newRecordingGenerator(t *testing.T, size int) *recordingGenerator {
	t.Helper()

	pool := make([]*rsa.PrivateKey, size)
	for i := range pool {
		key, err := rsa.GenerateKey(rand.Reader, deviceSignerBits)
		require.NoError(t, err)

		pool[i] = key
	}

	return &recordingGenerator{pool: pool}
}

func (g *recordingGenerator) generate() (*rsa.PrivateKey, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.generated == len(g.pool) {
		return nil, errors.New("the test ran out of keys")
	}

	key := g.pool[g.generated]
	g.generated++

	return key, nil
}

func (g *recordingGenerator) generatedSoFar() []*rsa.PrivateKey {
	g.mu.Lock()
	defer g.mu.Unlock()

	return slices.Clone(g.pool[:g.generated])
}

func TestDeviceSignerReserve(t *testing.T) {
	generator := newRecordingGenerator(t, 16)

	synctest.Test(t, func(t *testing.T) {
		at := clock.NewFixed(time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC))
		s := NewService(storemock.NewMockStore(t), privateKey, publicKey, newExpiringMemoryCache(at), WithClock(at), withDeviceSignerReserve(2, 2, generator.generate))

		seen := map[string]bool{}

		for range 3 {
			synctest.Wait()
			ahead := generator.generatedSoFar()

			key, err := s.MintDeviceSigner(t.Context(), "device-a")
			require.NoError(t, err)

			require.True(t, slices.ContainsFunc(ahead, func(reserved *rsa.PrivateKey) bool { return reserved.Equal(key) }), "the key must have been generated before the mint asked for it")

			public, err := ssh.NewPublicKey(&key.PublicKey)
			require.NoError(t, err)

			fingerprint := ssh.FingerprintLegacyMD5(public)
			require.False(t, seen[fingerprint], "the reserve handed the same key out twice")
			seen[fingerprint] = true

			_, err = s.AuthPublicKey(t.Context(), "device-a", requests.PublicKeyAuth{Fingerprint: fingerprint, Data: "data"})
			require.NoError(t, err)
		}
	})
}

func TestDeviceSignerReserveStopsRefillingWhileItsGeneratorFails(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const spinning = 100

		var calls atomic.Int32

		stuck := make(chan struct{})
		failing := func() (*rsa.PrivateKey, error) {
			if calls.Add(1) > spinning {
				<-stuck
			}

			return nil, errors.New("no entropy")
		}

		at := clock.NewFixed(time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC))
		s := NewService(storemock.NewMockStore(t), privateKey, publicKey, newExpiringMemoryCache(at), WithClock(at), withDeviceSignerReserve(2, 2, failing))

		synctest.Wait()
		require.LessOrEqual(t, calls.Load(), int32(2), "a failing generator must not be retried until a key is asked for")

		_, err := s.MintDeviceSigner(t.Context(), "device-a")
		require.Error(t, err)

		synctest.Wait()
		require.LessOrEqual(t, calls.Load(), int32(5), "a failing generator must not be retried until a key is asked for")

		close(stuck)
	})
}

func TestDeviceSignerReserveRefillsEveryTakenKeyInParallel(t *testing.T) {
	const size = 4

	generator := newRecordingGenerator(t, 2*size)

	synctest.Test(t, func(t *testing.T) {
		var (
			filled   atomic.Bool
			inFlight atomic.Int32
		)

		release := make(chan struct{})
		open := sync.OnceFunc(func() { close(release) })
		defer open()

		gated := func() (*rsa.PrivateKey, error) {
			if filled.Load() {
				inFlight.Add(1)
				defer inFlight.Add(-1)

				<-release
			}

			return generator.generate()
		}

		at := clock.NewFixed(time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC))
		s := NewService(storemock.NewMockStore(t), privateKey, publicKey, newExpiringMemoryCache(at), WithClock(at), withDeviceSignerReserve(size, size, gated))

		synctest.Wait()
		filled.Store(true)

		for taken := 1; taken <= size; taken++ {
			_, err := s.MintDeviceSigner(t.Context(), "device-a")
			require.NoError(t, err)

			synctest.Wait()
			require.Equal(t, int32(taken), inFlight.Load(), "each key taken must start generating its replacement")
		}

		open()
		synctest.Wait()

		require.Len(t, generator.generatedSoFar(), 2*size, "the reserve must refill every key it handed out")
	})
}
