package services

import (
	"context"
	"errors"
	"strings"

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

	// RenameDevice renames the specified device.
	// This method is deprecated, use [DeviceService#UpdateDevice] instead.
	RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error

	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	OfflineDevice(ctx context.Context, uid models.UID) error

	UpdateDevice(ctx context.Context, req *requests.DeviceUpdate) error
	// UpdateDeviceStatus updates a device's status. Devices that are already accepted cannot change their status.
	//
	// When accepting, if a device with the same MAC address is already accepted within the same namespace, it
	// merges these devices unless a third device with the same hostname already exists and has a different MAC
	// address. The merge transfers all sessions from the old device to the new one, renames the new device to
	// preserve the old device's identity, and deletes the old device. Also, if another accepted device already
	// uses the same hostname but has a different MAC address, the operation fails.
	//
	// Environment-specific Acceptance Rules:
	//   - Community/Enterprise: Only checks the namespace's device limit
	//   - Cloud (billing active): Reports device acceptance to billing service for quota/payment validation
	//   - Cloud (billing inactive): Checks if the device is removed and evaluates namespace capabilities:
	//     * If device was previously removed: removes from removed list, then evaluates billing
	//     * If device was not removed: counts total removed devices and checks against limits, then evaluates billing
	//     * Billing evaluation determines if the namespace can accept more devices based on subscription status
	//
	// All operations are performed within a database transaction to ensure consistency during device merging
	// and counter updates.
	UpdateDeviceStatus(ctx context.Context, req *requests.DeviceUpdateStatus) error
}

func (s *service) ListDevices(ctx context.Context, req *requests.DeviceList) ([]models.Device, int, error) {
	opts := []store.QueryOption{}

	if req.DeviceStatus != "" {
		opts = append(opts, s.store.Options().WithDeviceStatus(req.DeviceStatus))
	}

	if req.TenantID != "" {
		opts = append(opts, s.store.Options().InNamespace(req.TenantID))
	}

	opts = append(opts, s.store.Options().Match(&req.Filters), s.store.Options().Sort(&req.Sorter), s.store.Options().Paginate(&req.Paginator))

	if req.DeviceStatus == models.DeviceStatusRemoved {
		return s.store.DeviceList(ctx, store.DeviceAcceptableFromRemoved, opts...)
	}

	if req.TenantID != "" {
		ns, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
		if err != nil {
			return nil, 0, NewErrNamespaceNotFound(req.TenantID, err)
		}

		if ns.HasMaxDevices() {
			switch {
			case envs.IsCloud():
				if ns.HasLimitDevicesReached() {
					return s.store.DeviceList(ctx, store.DeviceAcceptableFromRemoved, opts...)
				}
			case envs.IsEnterprise():
				fallthrough
			case envs.IsCommunity():
				if ns.HasMaxDevicesReached() {
					return s.store.DeviceList(ctx, store.DeviceAcceptableAsFalse, opts...)
				}
			}
		}
	}

	return s.store.DeviceList(ctx, store.DeviceAcceptableIfNotAccepted, opts...)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid))
	if err != nil {
		return nil, NewErrDeviceNotFound(uid, err)
	}

	return device, nil
}

func (s *service) ResolveDevice(ctx context.Context, req *requests.ResolveDevice) (*models.Device, error) {
	n, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
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

	ns, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenant)
	if err != nil {
		return NewErrNamespaceNotFound(tenant, err)
	}

	// NOTE: If the namespace has a limit of devices, we change the device's slot status to removed only when it is
	// [models.DeviceStatusAccepted]. This way, we can keep track of the number of devices that were removed from the
	// namespace and void the device switching.
	if envs.IsCloud() && !ns.Billing.IsActive() && device.Status == models.DeviceStatusAccepted {
		now := clock.Now()
		changes := &models.DeviceChanges{DisconnectedAt: device.DisconnectedAt, RemovedAt: &now, Status: models.DeviceStatusRemoved}
		if err := s.store.DeviceUpdate(ctx, tenant, string(uid), changes); err != nil {
			return err
		}

		if err := s.store.NamespaceIncrementDeviceCount(ctx, tenant, models.DeviceStatusRemoved, 1); err != nil {
			return err
		}
	} else {
		if err := s.store.DeviceDelete(ctx, uid); err != nil {
			return err
		}
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

	if strings.EqualFold(device.Name, name) {
		return nil
	}

	changes := &models.DeviceChanges{DisconnectedAt: device.DisconnectedAt, RemovedAt: device.RemovedAt, Name: strings.ToLower(name)}
	if err := s.store.DeviceUpdate(ctx, device.TenantID, string(uid), changes); err != nil { // nolint:revive
		return err
	}

	return nil
}

// LookupDevice looks for a device in a namespace.
//
// It receives a context, used to "control" the request flow and, the namespace name from a models.Namespace and a
// device name from models.Device.
func (s *service) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	n, err := s.store.NamespaceResolve(ctx, store.NamespaceNameResolver, strings.ToLower(namespace))
	if err != nil {
		return nil, NewErrNamespaceNotFound(namespace, err)
	}

	opts := []store.QueryOption{
		s.store.Options().InNamespace(n.TenantID),
		s.store.Options().WithDeviceStatus(models.DeviceStatusAccepted),
	}

	device, err := s.store.DeviceResolve(ctx, store.DeviceHostnameResolver, name, opts...)
	if err != nil || device == nil {
		return nil, NewErrDeviceNotFound(models.UID(name), err)
	}

	return device, nil
}

func (s *service) OfflineDevice(ctx context.Context, uid models.UID) error {
	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid))
	if err != nil || device == nil {
		return NewErrDeviceNotFound(uid, err)
	}

	now := clock.Now()
	changes := &models.DeviceChanges{RemovedAt: device.RemovedAt, DisconnectedAt: &now}
	if err := s.store.DeviceUpdate(ctx, "", string(uid), changes); err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrDeviceNotFound(uid, err)
		}

		return err
	}

	return nil
}

func (s *service) UpdateDeviceStatus(ctx context.Context, req *requests.DeviceUpdateStatus) error {
	return s.store.WithTransaction(ctx, s.updateDeviceStatus(req))
}

func (s *service) updateDeviceStatus(req *requests.DeviceUpdateStatus) store.TransactionCb {
	return func(ctx context.Context) error {
		namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
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

		if newStatus == device.Status {
			return nil
		}

		if newStatus == models.DeviceStatusAccepted {
			opts := []store.QueryOption{s.store.Options().WithDeviceStatus(models.DeviceStatusAccepted), s.store.Options().InNamespace(namespace.TenantID)}
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

				if envs.IsCloud() {
					hasBillingActive := namespace.Billing != nil && namespace.Billing.IsActive()
					hasRechedLimit := namespace.HasMaxDevices() && namespace.HasLimitDevicesReached()
					isDeviceStatusRemoved := device.RemovedAt != nil

					if !hasBillingActive && hasRechedLimit && !isDeviceStatusRemoved {
						log.WithError(err).WithFields(log.Fields{"device_uid": device.UID}).
							Error("namespace's limit reached - cannot accept another device")

						return NewErrDeviceRemovedFull(namespace.MaxDevices, nil)
					}

					if err := s.handleCloudBilling(ctx, namespace); err != nil {
						log.WithError(err).WithFields(log.Fields{"device_uid": device.UID, "billing_active": namespace.Billing.IsActive()}).
							Error("billing validation failed")

						return err
					}
				} else {
					if namespace.HasMaxDevices() && namespace.HasMaxDevicesReached() {
						return NewErrDeviceMaxDevicesReached(namespace.MaxDevices)
					}
				}
			}
		}

		changes := &models.DeviceChanges{RemovedAt: device.RemovedAt, DisconnectedAt: device.DisconnectedAt, Status: newStatus}
		if err := s.store.DeviceUpdate(ctx, namespace.TenantID, device.UID, changes); err != nil {
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

func (s *service) UpdateDevice(ctx context.Context, req *requests.DeviceUpdate) error {
	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, req.UID, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		return NewErrDeviceNotFound(models.UID(req.UID), err)
	}

	conflictsTarget := &models.DeviceConflicts{Name: req.Name}
	conflictsTarget.Distinct(device)
	if _, has, err := s.store.DeviceConflicts(ctx, conflictsTarget); err != nil || has {
		return NewErrDeviceDuplicated(req.Name, err)
	}

	// We pass DisconnectedAt because we don't want to update it to nil
	changes := &models.DeviceChanges{DisconnectedAt: device.DisconnectedAt, RemovedAt: device.RemovedAt}
	if req.Name != "" && strings.ToLower(req.Name) != device.Name {
		changes.Name = strings.ToLower(req.Name)
	}

	return s.store.DeviceUpdate(ctx, req.TenantID, req.UID, changes)
}

// mergeDevice merges an old device into a new device. It transfers all sessions from the old device to the new one and
// renames the new device to preserve the old device's identity. The old device is then deleted and the namespace's device count is decremented.
func (s *service) mergeDevice(ctx context.Context, tenantID string, oldDevice *models.Device, newDevice *models.Device) error {
	if err := s.store.TunnelUpdateDeviceUID(ctx, tenantID, oldDevice.UID, newDevice.UID); err != nil {
		return err
	}

	if err := s.store.SessionUpdateDeviceUID(ctx, models.UID(oldDevice.UID), models.UID(newDevice.UID)); err != nil && !errors.Is(err, store.ErrNoDocuments) {
		return err
	}

	changes := &models.DeviceChanges{DisconnectedAt: newDevice.DisconnectedAt, RemovedAt: newDevice.RemovedAt, Name: oldDevice.Name}
	if err := s.store.DeviceUpdate(ctx, tenantID, newDevice.UID, changes); err != nil {
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

// handleCloudBilling processes billing-related operations for Cloud environment.
// This function has side effects: it may delete removed devices and report to billing.
func (s *service) handleCloudBilling(ctx context.Context, namespace *models.Namespace) error {
		if namespace.Billing.IsActive() {
			if err := s.BillingReport(ctx, s.client, namespace.TenantID, ReportDeviceAccept); err != nil {
				return NewErrBillingReportNamespaceDelete(err)
			}
		} else {
			ok, err := s.BillingEvaluate(ctx, s.client, namespace.TenantID)
		switch {
		case err != nil:
			return NewErrBillingEvaluate(err)
		case !ok:
			return ErrDeviceLimit
		}
	}

	return nil
}
