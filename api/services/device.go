package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const StatusAccepted = "accepted"

type DeviceService interface {
	ListDevices(ctx context.Context, req *requests.DeviceList) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID, tenant string) error
	RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	OfflineDevice(ctx context.Context, uid models.UID) error
	UpdateDeviceStatus(ctx context.Context, tenant string, uid models.UID, status models.DeviceStatus) error

	UpdateDevice(ctx context.Context, req *requests.DeviceUpdate) error
}

func (s *service) ListDevices(ctx context.Context, req *requests.DeviceList) ([]models.Device, int, error) {
	// if req.DeviceStatus == models.DeviceStatusRemoved {
	// 	// TODO: unique DeviceList
	// 	removed, count, err := s.store.DeviceRemovedList(ctx, req.TenantID, req.Paginator, req.Filters, req.Sorter)
	// 	if err != nil {
	// 		return nil, 0, err
	// 	}
	//
	// 	devices := make([]models.Device, 0, len(removed))
	// 	for _, device := range removed {
	// 		devices = append(devices, *device.Device)
	// 	}
	//
	// 	return devices, count, nil
	// }
	//
	// if req.TenantID != "" {
	// 	ns, err := s.store.NamespaceGet(ctx, req.TenantID, s.store.Options().CountAcceptedDevices())
	// 	if err != nil {
	// 		return nil, 0, NewErrNamespaceNotFound(req.TenantID, err)
	// 	}
	//
	// 	if ns.HasMaxDevices() {
	// 		switch {
	// 		case envs.IsCloud():
	// 			removed, err := s.store.DeviceRemovedCount(ctx, ns.TenantID)
	// 			if err != nil {
	// 				return nil, 0, NewErrDeviceRemovedCount(err)
	// 			}
	//
	// 			if ns.HasLimitDevicesReached(removed) {
	// 				return s.store.DeviceList(ctx, req.DeviceStatus, req.Paginator, req.Filters, req.Sorter, store.DeviceAcceptableFromRemoved)
	// 			}
	// 		case envs.IsEnterprise():
	// 			fallthrough
	// 		case envs.IsCommunity():
	// 			if ns.HasMaxDevicesReached() {
	// 				return s.store.DeviceList(ctx, req.DeviceStatus, req.Paginator, req.Filters, req.Sorter, store.DeviceAcceptableAsFalse)
	// 			}
	// 		}
	// 	}
	// }

	return s.store.DeviceList(
		ctx,
		s.store.Options().InNamespace(req.TenantID),
		s.store.Options().Filter(req.Filters),
		s.store.Options().Paginate(req.Paginator),
		s.store.Options().Order(req.Sorter),
	)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	device, err := s.store.DeviceGet(ctx, store.DeviceIdentID, string(uid))
	if err != nil {
		return nil, NewErrDeviceNotFound(uid, err)
	}

	return device, nil
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
	ns, err := s.store.NamespaceGet(ctx, store.NamespaceIdentID, tenant)
	if err != nil {
		return NewErrNamespaceNotFound(tenant, err)
	}

	device, err := s.store.DeviceGet(ctx, store.DeviceIdentID, string(uid))
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	// If the namespace has a limit of devices, we change the device's slot status to removed.
	// This way, we can keep track of the number of devices that were removed from the namespace and void the device
	// switching.
	if envs.IsCloud() && envs.HasBilling() && !ns.Billing.IsActive() {
		if err := s.store.DeviceRemovedInsert(ctx, tenant, device); err != nil {
			return NewErrDeviceRemovedInsert(err)
		}
	}

	return s.store.DeviceDelete(ctx, uid)
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error {
	return nil
}

// LookupDevice looks for a device in a namespace.
//
// It receives a context, used to "control" the request flow and, the namespace name from a models.Namespace and a
// device name from models.Device.
func (s *service) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	device, err := s.store.DeviceGet(ctx, store.DeviceIdentName, name)
	if err != nil {
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

func (s *service) UpdateDevice(ctx context.Context, req *requests.DeviceUpdate) error {
	return nil
}
