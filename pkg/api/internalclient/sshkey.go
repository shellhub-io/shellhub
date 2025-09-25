package internalclient

import (
	"context"
	"fmt"

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

func (c *client) GetPublicKey(ctx context.Context, fingerprint, tenant string) (*models.PublicKey, error) {
	pubKey := new(models.PublicKey)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetResult(&pubKey).
		Get(fmt.Sprintf("/internal/sshkeys/public-keys/%s/%s", fingerprint, tenant))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrNotFound
	}

	return pubKey, nil
}

func (c *client) EvaluateKey(ctx context.Context, fingerprint string, dev *models.Device, username string) (bool, error) {
	var evaluate *bool

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetBody(dev).
		SetResult(&evaluate).
		Post(fmt.Sprintf("/internal/sshkeys/public-keys/evaluate/%s/%s", fingerprint, username))
	if err != nil {
		return false, err
	}

	if resp.StatusCode() == http.StatusOK {
		return *evaluate, nil
	}

	return false, nil
}

func (c *client) CreatePrivateKey(ctx context.Context) (*models.PrivateKey, error) {
	privKey := new(models.PrivateKey)

	_, err := c.http.
		R().
		SetContext(ctx).
		SetResult(&privKey).
		Post("/internal/sshkeys/private-keys")
	if err != nil {
		return nil, err
	}

	return privKey, nil
}
