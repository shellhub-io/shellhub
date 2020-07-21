// +build !internal_api

package client

import (
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/models"
)

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
	AuthDevice(req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error)
	DeleteDevice(uid string) error
}

func (c *client) GetInfo() (*models.Info, error) {
	var info *models.Info
	_, _, errs := c.http.Get(buildURL(c, "/info")).EndStruct(&info)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return info, nil
}

func (c *client) AuthDevice(req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error) {
	var res *models.DeviceAuthResponse
	_, _, errs := c.http.Post(buildURL(c, "/api/devices/auth")).Send(req).EndStruct(&res)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return res, nil
}

func (c *client) Endpoints() (*models.Endpoints, error) {
	var endpoints *models.Endpoints
	_, _, errs := c.http.Get(buildURL(c, "/endpoints")).EndStruct(&endpoints)

	if len(errs) > 0 {
		return nil, errs[0]
	}

	return endpoints, nil
}

func (c *client) DeleteDevice(uid string) error {
	fmt.Println("delete device")
	_, _, errs := c.http.Delete(buildURL(c, fmt.Sprintf("/internal/devices/%s", uid))).End()
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
