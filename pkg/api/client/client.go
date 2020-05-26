package client

import (
	"errors"
	"fmt"
	"path"

	"github.com/parnurzeal/gorequest"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	DeviceUIDHeader = "X-Device-UID"

	ConnectionFailedErr = "Connection failed"
	NotFoundErr         = "Not found"
	UnknownErr          = "Unknown error"
)

func NewClient(opts ...Opt) Client {
	c := &client{
		host: apiHost,
		port: apiPort,
		http: gorequest.New(),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil
		}
	}

	return c
}

type commonAPI interface {
	ListDevices() ([]models.Device, error)
	GetDevice(uid string) (*models.Device, error)
}

type client struct {
	host string
	port int
	http *gorequest.SuperAgent
}

func (c *client) ListDevices() ([]models.Device, error) {
	list := []models.Device{}
	_, _, errs := c.http.Get(buildURL(c, "/api/devices")).EndStruct(&list)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return list, nil
}

func (c *client) GetDevice(uid string) (*models.Device, error) {
	var device *models.Device
	resp, _, errs := c.http.Get(buildURL(c, fmt.Sprintf("/api/devices/%s", uid))).EndStruct(&device)
	fmt.Println(buildURL(c, fmt.Sprintf("/api/devices/%s", uid)))
	if len(errs) > 0 {
		return nil, errors.New(ConnectionFailedErr)
	}

	if resp.StatusCode == 400 {
		return nil, errors.New(NotFoundErr)
	} else if resp.StatusCode == 200 {
		return device, nil
	}

	return nil, errors.New(UnknownErr)
}

func buildURL(c *client, uri string) string {
	return fmt.Sprintf("http://%s", path.Join(fmt.Sprintf("%s:%d", c.host, c.port), uri))
}
