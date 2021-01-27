// +build internal_api

package client

import (
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/models"
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
}

func (c *client) LookupDevice() {
}

func (c *client) GetPublicKey(fingerprint, tenant string) (*models.PublicKey, error) {
	var pubKey *models.PublicKey
	_, _, errs := c.http.Get(buildURL(c, fmt.Sprintf("/internal/sshkeys/public_keys/%s/%s", fingerprint, tenant))).EndStruct(&pubKey)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return pubKey, nil
}

func (c *client) CreatePrivateKey() (*models.PrivateKey, error) {
	var privKey *models.PrivateKey
	_, _, errs := c.http.Post(buildURL(c, "/internal/sshkeys/private_keys")).EndStruct(&privKey)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return privKey, nil
}
