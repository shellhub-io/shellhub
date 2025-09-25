package internalclient

import (
	"context"
	"errors"
	"fmt"
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
	Lookup(ctx context.Context, lookup map[string]string) (string, []error)

	// DeviceLookup performs a lookup operation based on the provided parameters.
	DeviceLookup(ctx context.Context, tenantID, name string) (*models.Device, error)

	// LookupWebEndpoints retrieves a web endpoint by its address.
	LookupWebEndpoints(ctx context.Context, address string) (*WebEndpoint, error)
}

func (c *client) DevicesOffline(ctx context.Context, uid string) error {
	_, err := c.http.
		R().
		SetContext(ctx).
		Post(fmt.Sprintf("/internal/devices/%s/offline", uid))
	if err != nil {
		return err
	}

	return nil
}

func (c *client) DevicesHeartbeat(ctx context.Context, uid string) error {
	return c.worker.SubmitToBatch(ctx, worker.TaskPattern("api:heartbeat"), []byte(uid))
}

func (c *client) Lookup(ctx context.Context, lookup map[string]string) (string, []error) {
	var device struct {
		UID string `json:"uid"`
	}

	resp, _ := c.http.
		R().
		SetContext(ctx).
		SetQueryParams(lookup).
		SetResult(&device).
		Get("/internal/lookup")

	if resp.StatusCode() != http.StatusOK {
		return "", []error{errors.New("lookup failed")}
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
		Get("/internal/device/lookup")
	if err != nil {
		return nil, ErrConnectionFailed
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return device, nil
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusForbidden:
		return nil, ErrForbidden
	default:
		return nil, ErrUnknown
	}
}

func (c *client) ListDevices(ctx context.Context) ([]models.Device, error) {
	list := []models.Device{}

	_, err := c.http.
		R().
		SetContext(ctx).
		SetResult(list).
		Get("/api/devices")

	return list, err
}

func (c *client) GetDevice(ctx context.Context, uid string) (*models.Device, error) {
	device := new(models.Device)
	resp, err := c.http.
		R().
		SetContext(ctx).
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
	var tunnel *WebEndpoint
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetResult(&tunnel).
		Get(fmt.Sprintf("http://cloud:8080/internal/web-endpoints/%s", address))
	if err != nil {
		return nil, ErrConnectionFailed
	}

	switch resp.StatusCode() {
	case 404:
		return nil, ErrNotFound
	case 403:
		return nil, ErrForbidden
	case 200:
		return tunnel, nil
	default:
		return nil, ErrUnknown
	}
}
