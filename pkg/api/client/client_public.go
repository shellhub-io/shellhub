// +build !internal_api

package client

import "github.com/shellhub-io/shellhub/pkg/models"

const (
	apiHost = "ssh.shellhub.io"
	apiPort = 80
)

type Client interface {
	commonAPI
	publicAPI
}

type publicAPI interface {
	GetInfo() (*models.Info, error)
	Endpoints() (*models.Endpoints, error)
}

func (c *client) GetInfo() (*models.Info, error) {
	var info *models.Info
	_, _, errs := c.http.Get(buildURL(c, "/info")).EndStruct(&info)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return info, nil
}

func (c *client) Endpoints() (*models.Endpoints, error) {
	var endpoints *models.Endpoints
	_, _, errs := c.http.Get(buildURL(c, "/endpoints")).EndStruct(&endpoints)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return endpoints, nil
}
