// +build internal_api

package client

import (
	"errors"
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/models"
	"go.uber.org/multierr"
)

const (
	apiHost   = "api"
	apiPort   = 8080
	apiScheme = "http"
)

type Client interface {
	commonAPI
	internalAPI
}

type internalAPI interface {
	LookupDevice()
	GetPublicKey(fingerprint, tenant string) (*models.PublicKey, error)
	CreatePrivateKey() (*models.PrivateKey, error)
	EvaluateKey(fingerprint string, dev *models.Device) (bool, error)
}

func (c *client) LookupDevice() {
}

func (c *client) GetPublicKey(fingerprint, tenant string) (*models.PublicKey, error) {
	var pubKey *models.PublicKey
	resp, _, errs := c.http.Get(buildURL(c, fmt.Sprintf("/internal/sshkeys/public-keys/%s/%s", fingerprint, tenant))).EndStruct(&pubKey)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	if resp.StatusCode == 404 {
		return nil, errors.New(NotFoundErr)
	}

	return pubKey, nil
}

func (c *client) EvaluateKey(fingerprint string, dev *models.Device) (bool, error) {
	var evaluate *bool

	resp, _, errs := c.http.Post(buildURL(c, fmt.Sprintf("/internal/sshkeys/public-keys/evaluate/%s", fingerprint))).Send(dev).EndStruct(&evaluate)
	if len(errs) > 0 {
		var err error
		for _, e := range errs {
			err = multierr.Append(err, e)
		}
		return false, err

	}
	if resp.StatusCode == 200 {
		return *evaluate, nil
	}

	return false, nil
}

func (c *client) CreatePrivateKey() (*models.PrivateKey, error) {
	var privKey *models.PrivateKey
	_, _, errs := c.http.Post(buildURL(c, "/internal/sshkeys/private-keys")).EndStruct(&privKey)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return privKey, nil
}
