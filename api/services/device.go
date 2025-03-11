package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const StatusAccepted = "accepted"

type DeviceService interface {
	ListDevices(ctx context.Context, req *requests.DeviceList) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	GetDeviceByPublicURLAddress(ctx context.Context, address string) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID, tenant string) error
	RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	OfflineDevice(ctx context.Context, uid models.UID) error
	UpdateDeviceStatus(ctx context.Context, tenant string, uid models.UID, status models.DeviceStatus) error
	UpdateDevice(ctx context.Context, tenant string, uid models.UID, name *string, publicURL *bool) error
}

func (s *service) ListDevices(ctx context.Context, req *requests.DeviceList) ([]models.Device, int, error) {
	return nil, 0, nil
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	return nil, nil
}

func (s *service) GetDeviceByPublicURLAddress(ctx context.Context, address string) (*models.Device, error) {
	return nil, nil
}

// DeleteDevice deletes a device from a namespace.
//
// It receives a context, used to "control" the request flow and, the device UID from models.Device and the tenant ID
// from models.Namespace.
//
// It can return an error if the device is not found, NewErrDeviceNotFound(uid, err), if the namespace is not found,
// NewErrNamespaceNotFound(tenant, err), if the usage cannot be reported, ErrReport or if the store function that
// delete the device fails.
func (s *service) DeleteDevice(ctx context.Context, uid models.UID, tenant string) error {
	return nil
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error {
	return nil
}

func (s *service) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	device, err := s.store.DeviceGet(ctx, "name", name, namespace)
	if err != nil || device == nil {
		return nil, NewErrDeviceLookupNotFound(namespace, name, err)
	}

	return device, nil
}

func (s *service) OfflineDevice(ctx context.Context, uid models.UID) error {
	return nil
}

// UpdateDeviceStatus updates the device status.
func (s *service) UpdateDeviceStatus(ctx context.Context, tenant string, uid models.UID, status models.DeviceStatus) error {
	return nil
}

func (s *service) UpdateDevice(ctx context.Context, tenant string, uid models.UID, name *string, publicURL *bool) error {
	return nil
}
