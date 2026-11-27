package services

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"runtime"
	"sync"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"golang.org/x/crypto/ssh"
)

const (
	deviceSignerTTL  = time.Minute
	deviceSignerBits = 2048
)

type deviceSigner struct {
	Key    []byte
	Device models.UID
}

func deviceSignerKey(fingerprint string) string {
	return "device-signer/" + fingerprint
}

type signerReserve struct {
	keys      chan *rsa.PrivateKey
	refilling sync.Mutex
}

func newSignerReserve(size int) *signerReserve {
	return &signerReserve{keys: make(chan *rsa.PrivateKey, size)}
}

func (r *signerReserve) take() (*rsa.PrivateKey, error) {
	if r == nil {
		return rsa.GenerateKey(rand.Reader, deviceSignerBits)
	}

	defer r.refill()

	select {
	case key := <-r.keys:
		return key, nil
	default:
		return rsa.GenerateKey(rand.Reader, deviceSignerBits)
	}
}

func (r *signerReserve) refill() {
	if !r.refilling.TryLock() {
		return
	}

	workers := min(max(runtime.GOMAXPROCS(0)/2, 1), cap(r.keys)-len(r.keys))

	var generating sync.WaitGroup

	for range workers {
		generating.Go(func() {
			for len(r.keys) < cap(r.keys) {
				key, err := rsa.GenerateKey(rand.Reader, deviceSignerBits)
				if err != nil {
					return
				}

				select {
				case r.keys <- key:
				default:
					return
				}
			}
		})
	}

	go func() {
		generating.Wait()
		r.refilling.Unlock()
	}()
}

func (s *service) MintDeviceSigner(ctx context.Context, device models.UID) (*rsa.PrivateKey, error) {
	key, err := s.signers.take()
	if err != nil {
		return nil, err
	}

	public, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}

	signer := deviceSigner{Key: x509.MarshalPKCS1PrivateKey(key), Device: device}
	if err := s.cache.Set(ctx, deviceSignerKey(ssh.FingerprintLegacyMD5(public)), signer, deviceSignerTTL); err != nil {
		return nil, err
	}

	return key, nil
}

func (s *service) AuthPublicKey(ctx context.Context, device models.UID, req requests.PublicKeyAuth) (*models.PublicKeyAuthResponse, error) {
	signer, err := cache.Get[deviceSigner](ctx, s.cache, deviceSignerKey(req.Fingerprint))
	if err != nil {
		return nil, NewErrPublicKeyNotFound(req.Fingerprint, err)
	}

	if device == "" || signer.Device != device {
		return nil, NewErrPublicKeyNotFound(req.Fingerprint, nil)
	}

	key, err := x509.ParsePKCS1PrivateKey(signer.Key)
	if err != nil {
		return nil, err
	}

	digest := sha256.Sum256([]byte(req.Data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return nil, err
	}

	return &models.PublicKeyAuthResponse{Signature: base64.StdEncoding.EncodeToString(signature)}, nil
}
