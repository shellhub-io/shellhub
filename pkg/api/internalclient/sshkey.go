package internalclient

import (
	"context"

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
		Get(c.config.APIBaseURL + "/internal/sshkeys/public-keys/{fingerprint}/{tenant}")
	if HasError(resp, err) {
		return nil, NewError(resp, err)
	}

	return pubKey, nil
}

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
		Post(c.config.APIBaseURL + "/internal/sshkeys/public-keys/evaluate/{fingerprint}/{username}")
	if HasError(resp, err) {
		return false, NewError(resp, err)
	}

	return *evaluate, nil
}

func (c *client) CreatePrivateKey(ctx context.Context) (*models.PrivateKey, error) {
	privKey := new(models.PrivateKey)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetResult(&privKey).
		Post(c.config.APIBaseURL + "/internal/sshkeys/private-keys")
	if HasError(resp, err) {
		return nil, NewError(resp, err)
	}

	return privKey, nil
}
