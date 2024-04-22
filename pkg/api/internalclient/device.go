package internalclient

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type deviceAPI interface {
	// ListDevices returns a list of devices.
	ListDevices() ([]models.Device, error)

	// GetDevice retrieves device information for the specified UID.
	GetDevice(uid string) (*models.Device, error)

	GetDeviceByPublicURLAddress(address string) (*models.Device, error)

	// Lookup performs a lookup operation based on the provided parameters.
	Lookup(lookup map[string]string) (string, []error)

	// DeviceLookup performs a lookup operation based on the provided parameters.
	DeviceLookup(lookup map[string]string) (*models.Device, []error)

	// UpdateDeviceConnectionStats updates the `connected_at` and `disconnected_at` attributes of a device with the specified
	// tenant and UID. If you want to avoid updating either attribute, pass a [time.Time]{} value.
	UpdateDeviceConnectionStats(tenant, uid string, connectedAt, disconnectedAt time.Time) (int, error)

	// NotifyConnectedDevicesIncrease sends a notification to increase the count of connected devices
	// for a specified tenant and target. The target can typically be a [github.com/shellhub-io/shellhub/pkg/models.DeviceStatus],
	// but if the device's status is unknown or not relevant, it also be the UID of the device. In such cases, the server will
	// use the device status stored in the database.
	//
	// This operation is asynchronous. An error is returned if the method is unable to queue the payload for sending.
	NotifyConnectedDevicesIncrease(tenant, target string) error

	// NotifyConnectedDevicesDecrease sends a notification to decrease the count of connected devices
	// for a specified tenant and target. The target can typically be a [github.com/shellhub-io/shellhub/pkg/models.DeviceStatus],
	// but if the device's status is unknown or not relevant, it also be the UID of the device. In such cases, the server will
	// use the device status stored in the database.
	//
	// This operation is asynchronous. An error is returned if the method is unable to queue the payload for sending.
	NotifyConnectedDevicesDecrease(tenant, target string) error
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

func (c *client) UpdateDeviceConnectionStats(tenant, uid string, connectedAt, disconnectedAt time.Time) (int, error) {
	r, err := c.http.R().
		SetHeader("X-Tenant-ID", tenant).
		SetHeader("Content-Type", "application/json").
		SetBody(&map[string]time.Time{
			"connected_at":    connectedAt,
			"disconnected_at": disconnectedAt,
		}).
		Patch(fmt.Sprintf("/internal/devices/%s/connection-stats", uid))
	if err != nil {
		return 0, err
	}

	return r.StatusCode(), nil
}

func (c *client) NotifyConnectedDevicesIncrease(tenant, target string) error {
	key := tenant + ":" + target
	_, err := c.asynq.Enqueue(asynq.NewTask("connected_devices:increase", []byte(key)), asynq.Queue("device"))

	return err
}

func (c *client) NotifyConnectedDevicesDecrease(tenant, target string) error {
	key := tenant + ":" + target
	_, err := c.asynq.Enqueue(asynq.NewTask("connected_devices:decrease", []byte(key)), asynq.Queue("device"))

	return err
}
