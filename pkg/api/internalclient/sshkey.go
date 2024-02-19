package internalclient

import (
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// sshkeyAPI defines methods for interacting with SSH key-related functionality.
type sshkeyAPI interface {
	// GetPublicKey retrieves the public key identified by the provided fingerprint and tenant.
	GetPublicKey(fingerprint, tenant string) (*models.PublicKey, error)

	// CreatePrivateKey creates a new private key.
	CreatePrivateKey() (*models.PrivateKey, error)

	// EvaluateKey evaluates whether a given public key identified by fingerprint is valid for a device and username combination.
	EvaluateKey(fingerprint string, dev *models.Device, username string) (bool, error)
}

func (c *client) GetPublicKey(fingerprint, tenant string) (*models.PublicKey, error) {
	pubKey := new(models.PublicKey)

	resp, err := c.http.
		R().
		SetResult(&pubKey).
		Get(fmt.Sprintf("/internal/sshkeys/public-keys/%s/%s", fingerprint, tenant))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == 404 {
		return nil, ErrNotFound
	}

	return pubKey, nil
}

func (c *client) EvaluateKey(fingerprint string, dev *models.Device, username string) (bool, error) {
	var evaluate *bool

	resp, err := c.http.
		R().
		SetBody(dev).
		SetResult(&evaluate).
		Post(fmt.Sprintf("/internal/sshkeys/public-keys/evaluate/%s/%s", fingerprint, username))
	if err != nil {
		return false, err
	}

	if resp.StatusCode() == 200 {
		return *evaluate, nil
	}

	return false, nil
}

func (c *client) CreatePrivateKey() (*models.PrivateKey, error) {
	privKey := new(models.PrivateKey)

	_, err := c.http.
		R().
		SetResult(&privKey).
		Post("/internal/sshkeys/private-keys")
	if err != nil {
		return nil, err
	}

	return privKey, nil
}
