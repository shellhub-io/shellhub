package internalclient

import (
	"context"
	"errors"
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// sshkeyAPI defines methods for interacting with SSH key-related functionality.
type sshkeyAPI interface {
	// GetPublicKey retrieves the public key identified by the provided fingerprint and tenant.
	GetPublicKey(ctx context.Context, fingerprint, tenant string) (*models.PublicKey, error)

	// CreatePrivateKey creates a new private key.
	CreatePrivateKey(ctx context.Context) (*models.PrivateKey, error)

	// EvaluateKey evaluates whether a given public key identified by fingerprint is valid for a device and username combination.
	EvaluateKey(ctx context.Context, fingerprint string, dev *models.Device, username string) (bool, error)
}

var (
	// ErrSSHKeyRequestFailed indicates that the SSH key request failed.
	ErrSSHKeyRequestFailed = errors.New("sshkey request failed")
	// ErrGetPublicKeyFailed indicates that the operation to get the public key failed.
	ErrGetPublicKeyFailed = errors.New("get public key failed")
)

func (c *client) GetPublicKey(ctx context.Context, fingerprint, tenant string) (*models.PublicKey, error) {
	pubKey := new(models.PublicKey)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParams(map[string]string{
			"fingerprint": fingerprint,
			"tenant":      tenant,
		}).
		SetResult(&pubKey).
		Get(c.Config.APIBaseURL + "/internal/sshkeys/public-keys/{fingerprint}/{tenant}")
	if err != nil {
		return nil, errors.Join(ErrSSHKeyRequestFailed, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, ErrGetPublicKeyFailed
	}

	return pubKey, nil
}

var ErrEvaluateKeyFailed = errors.New("evaluate key failed")

func (c *client) EvaluateKey(ctx context.Context, fingerprint string, dev *models.Device, username string) (bool, error) {
	var evaluate *bool

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParams(map[string]string{
			"fingerprint": fingerprint,
			"username":    username,
		}).
		SetBody(dev).
		SetResult(&evaluate).
		Post(c.Config.APIBaseURL + "/internal/sshkeys/public-keys/evaluate/{fingerprint}/{username}")
	if err != nil {
		return false, errors.Join(ErrSSHKeyRequestFailed, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return false, ErrEvaluateKeyFailed
	}

	return true, nil
}

var ErrCreatePrivateKeyFailed = errors.New("create private key failed")

func (c *client) CreatePrivateKey(ctx context.Context) (*models.PrivateKey, error) {
	privKey := new(models.PrivateKey)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetResult(&privKey).
		Post(c.Config.APIBaseURL + "/internal/sshkeys/private-keys")
	if err != nil {
		return nil, errors.Join(ErrSSHKeyRequestFailed, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, ErrCreatePrivateKeyFailed
	}

	return privKey, nil
}
