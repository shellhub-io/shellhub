package services

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"sync/atomic"
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
	keys     chan *rsa.PrivateKey
	filling  atomic.Int32
	workers  int32
	generate func() (*rsa.PrivateKey, error)
}

func generateDeviceSigner() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, deviceSignerBits)
}

func newSignerReserve(size, workers int, generate func() (*rsa.PrivateKey, error)) *signerReserve {
	return &signerReserve{keys: make(chan *rsa.PrivateKey, size), workers: int32(min(workers, size)), generate: generate} //nolint:gosec // workers is capped by size, a channel capacity that fits int32 by far
}

func (r *signerReserve) take() (*rsa.PrivateKey, error) {
	if r == nil {
		return generateDeviceSigner()
	}

	defer r.refill()

	select {
	case key := <-r.keys:
		return key, nil
	default:
		return r.generate()
	}
}

func (r *signerReserve) refill() {
	for range r.workers {
		filling := r.filling.Load()
		if filling >= r.workers || int(filling) >= cap(r.keys)-len(r.keys) {
			return
		}

		if r.filling.CompareAndSwap(filling, filling+1) {
			go r.fill()
		}
	}
}

func (r *signerReserve) fill() {
	for len(r.keys) < cap(r.keys) {
		key, err := r.generate()
		if err != nil {
			r.filling.Add(-1)

			return
		}

		select {
		case r.keys <- key:
		default:
		}
	}

	r.filling.Add(-1)
	r.refill()
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
