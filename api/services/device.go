package services

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
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
	opts := []store.QueryOption{
		s.store.Options().InNamespace(req.TenantID),
		s.store.Options().Filter(req.Filters),
		s.store.Options().Paginate(req.Paginator),
		s.store.Options().Order(req.Sorter),
	}

	if req.DeviceStatus != "" {
		opts = append(opts, s.store.Options().WithStatus(string(req.DeviceStatus)))
	}

	return s.store.DeviceList(ctx, opts...)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	device, err := s.store.DeviceGet(ctx, store.DeviceIdentUID, string(uid))
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
	device, err := s.store.DeviceGet(ctx, store.DeviceIdentUID, string(uid))
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	// ns, err := s.store.NamespaceGet(ctx, store.NamespaceIdentTenantID, tenant)
	// if err != nil {
	// 	return NewErrNamespaceNotFound(tenant, err)
	// }
	//
	// // If the namespace has a limit of devices, we change the device's slot status to removed.
	// // This way, we can keep track of the number of devices that were removed from the namespace and void the device
	// // switching.
	// if envs.IsCloud() && envs.HasBilling() && !ns.Billing.IsActive() {
	// 	if err := s.store.DeviceRemovedInsert(ctx, tenant, device); err != nil {
	// 		return NewErrDeviceRemovedInsert(err)
	// 	}
	// }

	return s.store.DeviceDelete(ctx, device)
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
	device, err := s.store.DeviceGet(ctx, store.DeviceIdentUID, string(uid))
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	now := clock.Now()
	device.DisconnectedAt = &now

	return s.store.DeviceSave(ctx, device)
}

// UpdateDeviceStatus updates the device status.
func (s *service) UpdateDeviceStatus(ctx context.Context, tenant string, uid models.UID, status models.DeviceStatus) error {
	device, err := s.store.DeviceGet(ctx, store.DeviceIdentUID, string(uid))
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	if device.Status == models.DeviceStatusAccepted {
		return NewErrDeviceStatusAccepted(nil)
	}

	device.Status = status

	return s.store.DeviceSave(ctx, device)

	// NOTICE: when the device is intended to be rejected or in pending status, we don't check for duplications as it
	// is not going to be considered for connections.
	// if status == models.DeviceStatusPending || status == models.DeviceStatusRejected {
	// 	device.Status = status
	//
	// 	return s.store.DeviceSave(ctx, device)
	// }
	//
	// namespace, err := s.store.NamespaceGet(ctx, store.NamespaceIdentTenantID, tenant)
	// if err != nil {
	// 	return NewErrNamespaceNotFound(tenant, err)
	// }
	//
	// // NOTICE: when the intended status is not accepted, we return an error because these status are not allowed
	// // to be set by the user.
	// if status != models.DeviceStatusAccepted {
	// 	return NewErrDeviceStatusInvalid(string(status), nil)
	// }
	//
	// // NOTICE: when there is an already accepted device with the same MAC address, we need to update the device UID
	// // transfer the sessions and delete the old device.
	// sameMacDev, err := s.store.DeviceGetByMac(ctx, device.Identity.MAC, device.TenantID, models.DeviceStatusAccepted)
	// if err != nil && err != store.ErrNoDocuments {
	// 	return NewErrDeviceNotFound(models.UID(device.UID), err)
	// }
	//
	// // TODO: move this logic to store's transactions.
	// if sameMacDev != nil && sameMacDev.UID != device.UID {
	// 	if sameName, err := s.store.DeviceGetByName(ctx, device.Name, device.TenantID, models.DeviceStatusAccepted); sameName != nil && sameName.Identity.MAC != device.Identity.MAC {
	// 		return NewErrDeviceDuplicated(device.Name, err)
	// 	}
	//
	// 	if err := s.store.SessionUpdateDeviceUID(ctx, models.UID(sameMacDev.UID), models.UID(device.UID)); err != nil && err != store.ErrNoDocuments {
	// 		return err
	// 	}
	//
	// 	if err := s.store.DeviceRename(ctx, models.UID(device.UID), sameMacDev.Name); err != nil {
	// 		return err
	// 	}
	//
	// 	if err := s.store.DeviceDelete(ctx, models.UID(sameMacDev.UID)); err != nil {
	// 		return err
	// 	}
	//
	// 	return s.store.DeviceUpdateStatus(ctx, uid, status)
	// }
	//
	// if sameName, err := s.store.DeviceGetByName(ctx, device.Name, device.TenantID, models.DeviceStatusAccepted); sameName != nil {
	// 	return NewErrDeviceDuplicated(device.Name, err)
	// }
	//
	// if status != models.DeviceStatusAccepted {
	// 	return s.store.DeviceUpdateStatus(ctx, uid, status)
	// }
	//
	// switch {
	// case envs.IsCommunity(), envs.IsEnterprise():
	// 	if namespace.HasMaxDevices() && namespace.HasMaxDevicesReached() {
	// 		return NewErrDeviceMaxDevicesReached(namespace.MaxDevices)
	// 	}
	// case envs.IsCloud():
	// 	if namespace.Billing.IsActive() {
	// 		if err := s.BillingReport(s.client, namespace.TenantID, ReportDeviceAccept); err != nil {
	// 			return NewErrBillingReportNamespaceDelete(err)
	// 		}
	// 	} else {
	// 		// TODO: this strategy that stores the removed devices in the database can be simplified.
	// 		removed, err := s.store.DeviceRemovedGet(ctx, tenant, uid)
	// 		if err != nil && err != store.ErrNoDocuments {
	// 			return NewErrDeviceRemovedGet(err)
	// 		}
	//
	// 		if removed != nil {
	// 			if err := s.store.DeviceRemovedDelete(ctx, tenant, uid); err != nil {
	// 				return NewErrDeviceRemovedDelete(err)
	// 			}
	// 		} else {
	// 			count, err := s.store.DeviceRemovedCount(ctx, tenant)
	// 			if err != nil {
	// 				return NewErrDeviceRemovedCount(err)
	// 			}
	//
	// 			if namespace.HasMaxDevices() && namespace.HasLimitDevicesReached(count) {
	// 				return NewErrDeviceRemovedFull(namespace.MaxDevices, nil)
	// 			}
	// 		}
	//
	// 		ok, err := s.BillingEvaluate(s.client, namespace.TenantID)
	// 		if err != nil {
	// 			return NewErrBillingEvaluate(err)
	// 		}
	//
	// 		if !ok {
	// 			return ErrDeviceLimit
	// 		}
	// 	}
	// }
	//
	// return s.store.DeviceUpdateStatus(ctx, uid, status)
}

func (s *service) UpdateDevice(ctx context.Context, req *requests.DeviceUpdate) error {
	device, err := s.store.DeviceGet(ctx, store.DeviceIdentUID, string(req.UID))
	if err != nil {
		return NewErrDeviceNotFound(models.UID(req.UID), err)
	}

	conflictsTarget := &models.DeviceConflicts{Name: req.Name}
	conflictsTarget.Distinct(device)
	if _, has, err := s.store.DeviceConflicts(ctx, conflictsTarget); err != nil || has {
		return NewErrDeviceDuplicated(req.Name, err)
	}

	if req.Name != "" {
		device.Name = strings.ToLower(req.Name)
	}

	if err := s.store.DeviceSave(ctx, device); err != nil {
		return err
	}

	return nil
}
