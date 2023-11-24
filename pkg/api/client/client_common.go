package client

import (
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/models"
)

func (c *client) ListDevices() ([]models.Device, error) {
	devices := make([]models.Device, 0)

	response, err := c.http.R().
		SetResult(&devices).
		Get("/api/devices")
	if err != nil {
		return nil, err
	}

	if err := ErrorFromResponse(response); err != nil {
		return nil, err
	}

	return devices, nil
}

func (c *client) GetDevice(uid string) (*models.Device, error) {
	var device *models.Device

	response, err := c.http.R().
		SetResult(&device).
		Get(fmt.Sprintf("/api/devices/%s", uid))
	if err != nil {
		return nil, ErrConnectionFailed
	}

	if err := ErrorFromResponse(response); err != nil {
		return nil, err
	}

	return device, nil
}
