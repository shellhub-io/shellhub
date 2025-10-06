package internalclient

import (
	"context"
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

func (c *client) DevicesOffline(ctx context.Context, uid string) error {
	res, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		Post(c.config.APIBaseURL + "/internal/devices/{uid}/offline")

	return NewError(res, err)
}

func (c *client) DevicesHeartbeat(ctx context.Context, uid string) error {
	return c.worker.SubmitToBatch(ctx, worker.TaskPattern("api:heartbeat"), []byte(uid))
}

func (c *client) Lookup(ctx context.Context, lookup map[string]string) (string, error) {
	var device struct {
		UID string `json:"uid"`
	}

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetQueryParams(lookup).
		SetResult(&device).
		Get(c.config.APIBaseURL + "/internal/lookup")
	if HasError(resp, err) {
		return "", NewError(resp, err)
	}

	return device.UID, nil
}

func (c *client) DeviceLookup(ctx context.Context, tenantID, name string) (*models.Device, error) {
	device := new(models.Device)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetQueryParam("tenant_id", tenantID).
		SetQueryParam("name", name).
		SetResult(&device).
		Get(c.config.APIBaseURL + "/internal/device/lookup")
	if HasError(resp, err) {
		return nil, NewError(resp, err)
	}

	return device, nil
}

func (c *client) ListDevices(ctx context.Context) ([]models.Device, error) {
	list := []models.Device{}

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetResult(&list).
		Get(c.config.APIBaseURL + "/api/devices")
	if HasError(resp, err) {
		return nil, NewError(resp, err)
	}

	return list, nil
}

func (c *client) GetDevice(ctx context.Context, uid string) (*models.Device, error) {
	device := new(models.Device)
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetResult(&device).
		Get(c.config.APIBaseURL + "/api/devices/{uid}")
	if HasError(resp, err) {
		return nil, NewError(resp, err)
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

func (c *client) LookupWebEndpoints(ctx context.Context, address string) (*WebEndpoint, error) {
	var endpoint *WebEndpoint
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("address", address).
		SetResult(&endpoint).
		Get(c.config.EnterpriseBaseURL + "/internal/web-endpoints/{address}")
	if HasError(resp, err) {
		return nil, NewError(resp, err)
	}

	return endpoint, nil
}
