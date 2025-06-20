package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

const StatusAccepted = "accepted"

type DeviceService interface {
	ListDevices(ctx context.Context, req *requests.DeviceList) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)

	// ResolveDevice attempts to resolve a device by searching for either its UID or hostname. When both are provided,
	// UID takes precedence over hostname. The search is scoped to the namespace's tenant ID to limit results.
	//
	// It returns the resolved device and any error encountered.
	ResolveDevice(ctx context.Context, req *requests.ResolveDevice) (*models.Device, error)

	DeleteDevice(ctx context.Context, uid models.UID, tenant string) error
	RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	OfflineDevice(ctx context.Context, uid models.UID) error

	// UpdateDeviceStatus updates a device's status. Devices that are already accepted cannot change their status.
	//
	// If a device with the same MAC address is already accepted within the same namespace, it merges these devices
	// unless a third device with the same hostname already exists. The merge transfers all sessions from the old
	// device to the new one and renames the new device to preserve the old device's identity. The old device is deleted.
	//
	// Accepting a device depends on the instance's environment:
	//   - Community/Enterprise: Only checks the namespace's device limit
	//   - Cloud: Verifies billing status and considers removed devices in limit calculations. When billing is active,
	//     the device acceptance is reported for billing purposes
	UpdateDeviceStatus(ctx context.Context, req *requests.UpdateDeviceStatus) error

	UpdateDevice(ctx context.Context, req *requests.DeviceUpdate) error
}

func (s *service) ListDevices(ctx context.Context, req *requests.DeviceList) ([]models.Device, int, error) {
	if req.DeviceStatus == models.DeviceStatusRemoved {
		// TODO: unique DeviceList
		removed, count, err := s.store.DeviceRemovedList(ctx, req.TenantID, req.Paginator, req.Filters, req.Sorter)
		if err != nil {
			return nil, 0, err
		}

		devices := make([]models.Device, 0, len(removed))
		for _, device := range removed {
			devices = append(devices, *device.Device)
		}

		return devices, count, nil
	}

	if req.TenantID != "" {
		ns, err := s.store.NamespaceGet(ctx, req.TenantID)
		if err != nil {
			return nil, 0, NewErrNamespaceNotFound(req.TenantID, err)
		}

		if ns.HasMaxDevices() {
			switch {
			case envs.IsCloud():
				removed, err := s.store.DeviceRemovedCount(ctx, ns.TenantID)
				if err != nil {
					return nil, 0, NewErrDeviceRemovedCount(err)
				}

				if ns.HasLimitDevicesReached(removed) {
					return s.store.DeviceList(ctx, req.DeviceStatus, req.Paginator, req.Filters, req.Sorter, store.DeviceAcceptableFromRemoved)
				}
			case envs.IsEnterprise():
				fallthrough
			case envs.IsCommunity():
				if ns.HasMaxDevicesReached() {
					return s.store.DeviceList(ctx, req.DeviceStatus, req.Paginator, req.Filters, req.Sorter, store.DeviceAcceptableAsFalse)
				}
			}
		}
	}

	return s.store.DeviceList(ctx, req.DeviceStatus, req.Paginator, req.Filters, req.Sorter, store.DeviceAcceptableIfNotAccepted)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid))
	if err != nil {
		return nil, NewErrDeviceNotFound(uid, err)
	}

	return device, nil
}

func (s *service) ResolveDevice(ctx context.Context, req *requests.ResolveDevice) (*models.Device, error) {
	n, err := s.store.NamespaceGet(ctx, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	var device *models.Device
	switch {
	case req.UID != "":
		device, err = s.store.DeviceResolve(ctx, store.DeviceUIDResolver, req.UID, s.store.Options().InNamespace(n.TenantID))
	case req.Hostname != "":
		device, err = s.store.DeviceResolve(ctx, store.DeviceHostnameResolver, req.Hostname, s.store.Options().InNamespace(n.TenantID))
	}

	if err != nil {
		// TODO: refactor this error to accept a string instead of models.UID
		return nil, NewErrDeviceNotFound(models.UID(""), err)
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
	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid), s.store.Options().InNamespace(tenant))
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	ns, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil {
		return NewErrNamespaceNotFound(tenant, err)
	}

	// If the namespace has a limit of devices, we change the device's slot status to removed.
	// This way, we can keep track of the number of devices that were removed from the namespace and void the device
	// switching.
	if envs.IsCloud() && envs.HasBilling() && !ns.Billing.IsActive() {
		if err := s.store.DeviceRemovedInsert(ctx, tenant, device); err != nil {
			return NewErrDeviceRemovedInsert(err)
		}

		if err := s.store.NamespaceIncrementDeviceCount(ctx, tenant, models.DeviceStatusRemoved, 1); err != nil { //nolint:revive
			return err
		}
	}

	if err := s.store.DeviceDelete(ctx, uid); err != nil {
		return err
	}

	if err := s.store.NamespaceIncrementDeviceCount(ctx, tenant, device.Status, -1); err != nil { //nolint:revive
		return err
	}

	return nil
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error {
	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid), s.store.Options().InNamespace(tenant))
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	updatedDevice := &models.Device{
		UID:        device.UID,
		Name:       strings.ToLower(name),
		Identity:   device.Identity,
		Info:       device.Info,
		PublicKey:  device.PublicKey,
		TenantID:   device.TenantID,
		LastSeen:   device.LastSeen,
		Online:     device.Online,
		Namespace:  device.Namespace,
		Status:     device.Status,
		CreatedAt:  time.Time{},
		RemoteAddr: "",
		Position:   &models.DevicePosition{},
		Tags:       []string{},
	}

	if ok, err := s.validator.Struct(updatedDevice); !ok || err != nil {
		return NewErrDeviceInvalid(nil, err)
	}

	if device.Name == updatedDevice.Name {
		return nil
	}

	otherDevice, err := s.store.DeviceResolve(ctx, store.DeviceHostnameResolver, updatedDevice.Name, s.store.Options().WithDeviceStatus(models.DeviceStatusAccepted), s.store.Options().InNamespace(tenant))
	if err != nil && err != store.ErrNoDocuments {
		return NewErrDeviceNotFound(models.UID(updatedDevice.UID), err)
	}

	if otherDevice != nil {
		return NewErrDeviceDuplicated(otherDevice.Name, err)
	}

	return s.store.DeviceRename(ctx, uid, name)
}

// LookupDevice looks for a device in a namespace.
//
// It receives a context, used to "control" the request flow and, the namespace name from a models.Namespace and a
// device name from models.Device.
func (s *service) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	n, err := s.store.NamespaceGetByName(ctx, namespace)
	if err != nil {
		return nil, NewErrNamespaceNotFound(namespace, err)
	}

	device, err := s.store.DeviceResolve(ctx, store.DeviceHostnameResolver, name, s.store.Options().InNamespace(n.TenantID))
	if err != nil || device == nil {
		return nil, NewErrDeviceNotFound(models.UID(name), err)
	}

	return device, nil
}

func (s *service) OfflineDevice(ctx context.Context, uid models.UID) error {
	now := clock.Now()
	if err := s.store.DeviceUpdate(ctx, "", string(uid), &models.DeviceChanges{DisconnectedAt: &now}); err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrDeviceNotFound(uid, err)
		}

		return err
	}

	return nil
}

func (s *service) UpdateDeviceStatus(ctx context.Context, req *requests.UpdateDeviceStatus) error {
	return s.store.WithTransaction(ctx, s.updateDeviceStatus(req))
}

func (s *service) updateDeviceStatus(req *requests.UpdateDeviceStatus) store.TransactionCb {
	return func(ctx context.Context) error {
		namespace, err := s.store.NamespaceGet(ctx, req.TenantID)
		if err != nil {
			return NewErrNamespaceNotFound(req.TenantID, err)
		}

		device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, req.UID, s.store.Options().InNamespace(namespace.TenantID))
		if err != nil {
			return NewErrDeviceNotFound(models.UID(req.UID), err)
		}

		if device.Status == models.DeviceStatusAccepted {
			log.WithFields(log.Fields{"device_uid": device.UID}).
				Warn("cannot change status - device already accepted")

			return NewErrDeviceStatusAccepted(nil)
		}

		oldStatus := device.Status
		newStatus := models.DeviceStatus(req.Status)
		opts := []store.QueryOption{s.store.Options().WithDeviceStatus(models.DeviceStatusAccepted), s.store.Options().InNamespace(namespace.TenantID)}

		if newStatus == models.DeviceStatusAccepted {
			existingMacDevice, err := s.store.DeviceResolve(ctx, store.DeviceMACResolver, device.Identity.MAC, opts...)
			if err != nil && !errors.Is(err, store.ErrNoDocuments) {
				log.WithError(err).
					WithFields(log.Fields{"mac": device.Identity.MAC}).
					Error("failed to retrieve device using MAC")

				return err
			}

			if existingMacDevice != nil && existingMacDevice.UID != device.UID {
				existingNameDevice, err := s.store.DeviceResolve(ctx, store.DeviceHostnameResolver, device.Name, opts...)
				if err != nil && !errors.Is(err, store.ErrNoDocuments) {
					log.WithError(err).
						WithFields(log.Fields{"name": device.Name}).
						Error("failed to retrieve device using name")

					return err
				}

				if existingNameDevice != nil && existingNameDevice.Identity.MAC != device.Identity.MAC {
					log.WithFields(log.Fields{"device_uid": device.UID, "device_mac": device.Identity.MAC, "conflicting_device_name": device.Name}).
						Error("device merge blocked - hostname already used by device with different MAC address")

					return NewErrDeviceDuplicated(device.Name, nil)
				}

				if err := s.mergeDevice(ctx, namespace.TenantID, existingMacDevice, device); err != nil {
					log.WithError(err).
						WithFields(log.Fields{"device_uid": device.UID, "existing_device_uid": existingMacDevice.UID, "device_mac": device.Identity.MAC}).
						Error("device merge operation failed")

					return err
				}
			} else {
				existingDevice, err := s.store.DeviceResolve(ctx, store.DeviceHostnameResolver, device.Name, opts...)
				if err != nil && !errors.Is(err, store.ErrNoDocuments) {
					log.WithError(err).
						WithFields(log.Fields{"name": device.Name}).
						Error("failed to retrieve device using name")

					return err
				}

				if existingDevice != nil {
					log.WithFields(log.Fields{"device_uid": device.UID, "conflicting_device_name": device.Name}).
						Error("device acceptance blocked - hostname already used by another device")

					return NewErrDeviceDuplicated(device.Name, nil)
				}

				if err := s.checkDeviceLimits(ctx, namespace, device); err != nil {
					log.WithError(err).WithFields(log.Fields{"device_uid": device.UID}).
						Error("namespace's limit reached - cannot accept another device")

					return err
				}

				if envs.IsCloud() {
					if namespace.Billing.IsActive() {
						if err := s.BillingReport(s.client, namespace.TenantID, ReportDeviceAccept); err != nil {
							return NewErrBillingReportNamespaceDelete(err)
						}
					} else {
						removed, err := s.store.DeviceRemovedGet(ctx, namespace.TenantID, models.UID(device.UID))
						if err != nil && err != store.ErrNoDocuments {
							return NewErrDeviceRemovedGet(err)
						}

						if removed != nil {
							oldStatus = models.DeviceStatusRemoved
							if err := s.store.DeviceRemovedDelete(ctx, namespace.TenantID, models.UID(device.UID)); err != nil {
								return NewErrDeviceRemovedDelete(err)
							}
						}

						ok, err := s.BillingEvaluate(s.client, namespace.TenantID)
						switch {
						case err != nil:
							return NewErrBillingEvaluate(err)
						case !ok:
							return ErrDeviceLimit
						}
					}
				}
			}
		}

		if err := s.store.DeviceUpdate(ctx, namespace.TenantID, device.UID, &models.DeviceChanges{Status: newStatus}); err != nil {
			return err
		}

		for status, count := range map[models.DeviceStatus]int64{oldStatus: -1, newStatus: 1} {
			if err := s.store.NamespaceIncrementDeviceCount(ctx, namespace.TenantID, status, count); err != nil {
				return err
			}
		}

		return nil
	}
}

// mergeDevice merges an old device into a new device. It transfers all sessions from the old device to the new one and
// renames the new device to preserve the old device's identity. The old device is then deleted and the namespace's device count is decremented.
func (s *service) mergeDevice(ctx context.Context, tenantID string, oldDevice *models.Device, newDevice *models.Device) error {
	// TODO: update tunnels as well?

	if err := s.store.SessionUpdateDeviceUID(ctx, models.UID(oldDevice.UID), models.UID(newDevice.UID)); err != nil && !errors.Is(err, store.ErrNoDocuments) {
		return err
	}

	if err := s.store.DeviceUpdate(ctx, tenantID, newDevice.UID, &models.DeviceChanges{Name: oldDevice.Name}); err != nil {
		return err
	}

	if err := s.store.DeviceDelete(ctx, models.UID(oldDevice.UID)); err != nil {
		return err
	}

	if err := s.store.NamespaceIncrementDeviceCount(ctx, tenantID, oldDevice.Status, -1); err != nil { //nolint:revive
		return err
	}

	return nil
}

// checkDeviceLimits validates if the namespace can accept more devices based on environment-specific limits.
func (s *service) checkDeviceLimits(ctx context.Context, namespace *models.Namespace, device *models.Device) error {
	switch {
	case envs.IsCommunity(), envs.IsEnterprise():
		if namespace.HasMaxDevices() && namespace.HasMaxDevicesReached() {
			return NewErrDeviceMaxDevicesReached(namespace.MaxDevices)
		}

		return nil
	case envs.IsCloud():
		if !namespace.Billing.IsActive() {
			_, err := s.store.DeviceRemovedGet(ctx, namespace.TenantID, models.UID(device.UID))
			if err != nil {
				if !errors.Is(err, store.ErrNoDocuments) {
					return NewErrDeviceRemovedGet(err)
				}

				count, err := s.store.DeviceRemovedCount(ctx, namespace.TenantID)
				if err != nil {
					return NewErrDeviceRemovedCount(err)
				}

				if namespace.HasMaxDevices() && namespace.HasLimitDevicesReached(count) {
					return NewErrDeviceRemovedFull(namespace.MaxDevices, nil)
				}
			}
		}

		return nil
	default:
		return nil
	}
}
