package internalclient

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/worker"
)

type deviceAPI interface {
	// ListDevices returns a list of devices.
	ListDevices(ctx context.Context) ([]models.Device, error)

	// GetDevice retrieves device information for the specified UID.
	GetDevice(ctx context.Context, uid string) (*models.Device, error)

	// DevicesOffline updates a device's status to offline.
	DevicesOffline(ctx context.Context, uid string) error

	// DevicesHeartbeat enqueues a task to send a heartbeat for the device.
	DevicesHeartbeat(ctx context.Context, uid string) error

	// Lookup performs a lookup operation based on the provided parameters.
	Lookup(ctx context.Context, lookup map[string]string) (string, error)

	// DeviceLookup performs a lookup operation based on the provided parameters.
	DeviceLookup(ctx context.Context, tenantID, name string) (*models.Device, error)

	// LookupWebEndpoints retrieves a web endpoint by its address.
	LookupWebEndpoints(ctx context.Context, address string) (*WebEndpoint, error)
}

// ErrDeviceRequestFailed indicates that the device request failed.
var ErrDeviceRequestFailed = errors.New("device request failed")

func (c *client) DevicesOffline(ctx context.Context, uid string) error {
	_, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		Post(c.Config.APIBaseURL + "/internal/devices/{uid}/offline")
	if err != nil {
		return errors.Join(ErrDeviceRequestFailed, err)
	}

	return nil
}

func (c *client) DevicesHeartbeat(ctx context.Context, uid string) error {
	return c.worker.SubmitToBatch(ctx, worker.TaskPattern("api:heartbeat"), []byte(uid))
}

// ErrLookupFailed indicates that the lookup operation failed.
var ErrLookupFailed = errors.New("lookup failed")

func (c *client) Lookup(ctx context.Context, lookup map[string]string) (string, error) {
	var device struct {
		UID string `json:"uid"`
	}

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetQueryParams(lookup).
		SetResult(&device).
		Get(c.Config.APIBaseURL + "/internal/lookup")
	if err != nil {
		return "", errors.Join(ErrDeviceRequestFailed, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return "", ErrLookupFailed
	}

	return device.UID, nil
}

// ErrDeviceLookupFailed indicates that the device lookup operation failed.
var ErrDeviceLookupFailed = errors.New("device lookup failed")

func (c *client) DeviceLookup(ctx context.Context, tenantID, name string) (*models.Device, error) {
	device := new(models.Device)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetQueryParam("tenant_id", tenantID).
		SetQueryParam("name", name).
		SetResult(&device).
		Get(c.Config.APIBaseURL + "/internal/device/lookup")
	if err != nil {
		return nil, errors.Join(ErrDeviceRequestFailed, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, ErrDeviceLookupFailed
	}

	return device, nil
}

// ErrListDevicesFailed indicates that the operation to list devices failed.
var ErrListDevicesFailed = errors.New("list devices failed")

func (c *client) ListDevices(ctx context.Context) ([]models.Device, error) {
	list := []models.Device{}

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetResult(list).
		Get(c.Config.APIBaseURL + "/api/devices")
	if err != nil {
		return nil, errors.Join(ErrDeviceRequestFailed, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, ErrListDevicesFailed
	}

	return list, nil
}

// ErrGetDeviceFailed indicates that the operation to get a device failed.
var ErrGetDeviceFailed = errors.New("get device failed")

func (c *client) GetDevice(ctx context.Context, uid string) (*models.Device, error) {
	device := new(models.Device)
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetResult(&device).
		Get(c.Config.APIBaseURL + "/api/devices/{uid}")
	if err != nil {
		return nil, errors.Join(ErrDeviceRequestFailed, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, ErrGetDeviceFailed
	}

	return device, nil
}

type WebEndpoint struct {
	Address    string         `json:"address"`
	Namespace  string         `json:"namespace"`
	DeviceUID  string         `json:"device_uid"`
	Device     *models.Device `json:"device"`
	Host       string         `json:"host"`
	Port       int            `json:"port"`
	TimeToLive int            `json:"ttl"`
	ExpiresIn  time.Time      `json:"expires_in"`
	CreatedAt  time.Time      `json:"time" bson:"time"`
}

var (
	// ErrWebEndpointRequestFailed indicates that the web endpoint request failed.
	ErrWebEndpointRequestFailed = errors.New("web endpoint request failed")
	// ErrWebEndpointForbidden indicates that access to the web endpoint is forbidden.
	ErrWebEndpointForbidden = errors.New("web endpoint access forbidden")
)

func (c *client) LookupWebEndpoints(ctx context.Context, address string) (*WebEndpoint, error) {
	var endpoint *WebEndpoint
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("address", address).
		SetResult(&endpoint).
		Get(c.Config.EnterpriseBaseURL + "/internal/web-endpoints/{address}")
	if err != nil {
		return nil, errors.Join(ErrWebEndpointRequestFailed, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, ErrWebEndpointForbidden
	}

	return endpoint, nil
}
