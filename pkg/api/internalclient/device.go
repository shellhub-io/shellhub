package internalclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/worker"
)

type deviceAPI interface {
	// ListDevices returns a list of devices.
	ListDevices() ([]models.Device, error)

	// GetDevice retrieves device information for the specified UID.
	GetDevice(uid string) (*models.Device, error)

	GetDeviceByPublicURLAddress(address string) (*models.Device, error)

	// DevicesOffline updates a device's status to offline.
	DevicesOffline(uid string) error

	// DevicesHeartbeat enqueues a task to send a heartbeat for the device.
	DevicesHeartbeat(tenant, uid string) error

	// Lookup performs a lookup operation based on the provided parameters.
	Lookup(lookup map[string]string) (string, []error)

	// DeviceLookup performs a lookup operation based on the provided parameters.
	DeviceLookup(lookup map[string]string) (*models.Device, []error)
}

func (c *client) DevicesOffline(uid string) error {
	_, err := c.http.
		R().
		Post(fmt.Sprintf("/internal/devices/%s/offline", uid))
	if err != nil {
		return err
	}

	return nil
}

func (c *client) DevicesHeartbeat(tenant, uid string) error {
	payload := fmt.Sprintf("%s:%s=%d", tenant, uid, clock.Now().Unix())

	return c.worker.SubmitToBatch(context.TODO(), worker.TaskPattern("api:heartbeat"), []byte(payload))
}

func (c *client) Lookup(lookup map[string]string) (string, []error) {
	var device struct {
		UID string `json:"uid"`
	}

	resp, _ := c.http.
		R().
		SetQueryParams(lookup).
		SetResult(&device).
		Get("/internal/lookup")

	if resp.StatusCode() != http.StatusOK {
		return "", []error{errors.New("lookup failed")}
	}

	return device.UID, nil
}

func (c *client) DeviceLookup(lookup map[string]string) (*models.Device, []error) {
	device := new(models.Device)

	resp, err := c.http.
		R().
		SetQueryParams(lookup).
		SetResult(&device).
		Get("/internal/lookup")
	if err != nil {
		return nil, []error{err}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, []error{errors.New("fail to get the device from the API")}
	}

	return device, nil
}

func (c *client) ListDevices() ([]models.Device, error) {
	list := []models.Device{}

	_, err := c.http.
		R().
		SetResult(list).
		Get("/api/devices")

	return list, err
}

func (c *client) GetDevice(uid string) (*models.Device, error) {
	device := new(models.Device)

	resp, err := c.http.
		R().
		SetResult(&device).
		Get(fmt.Sprintf("/api/devices/%s", uid))
	if err != nil {
		return nil, ErrConnectionFailed
	}

	switch resp.StatusCode() {
	case 400:
		return nil, ErrNotFound
	case 200:
		return device, nil
	default:
		return nil, ErrUnknown
	}
}

func (c *client) GetDeviceByPublicURLAddress(address string) (*models.Device, error) {
	httpClient := resty.New()

	var device *models.Device
	resp, err := httpClient.
		R().
		SetResult(&device).
		Get(fmt.Sprintf("/internal/devices/public/%s", address))
	if err != nil {
		return nil, ErrConnectionFailed
	}

	switch resp.StatusCode() {
	case 404:
		return nil, ErrNotFound
	case 200:
		return device, nil
	default:
		return nil, ErrUnknown
	}
}
