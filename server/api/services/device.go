package services

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/authctx"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

const StatusAccepted = "accepted"

// DeviceFilterFields maps each filter field the device list endpoint accepts
// to the set of operators valid for it. Operators that the database rejects
// on a given column type (e.g. ILIKE on the status enum) are omitted so the
// handler returns HTTP 400 instead of letting the store produce a 500.
var DeviceFilterFields = query.NewFieldConstraints(map[string][]string{
	"name":          {"contains", "eq", "ne"},
	"status":        {"eq", "ne"},
	"mac":           {"contains", "eq", "ne"},
	"platform":      {"contains", "eq", "ne"},
	"tags.name":     {"contains", "eq"},
	"online":        {"bool", "eq"},
	"custom_fields": {"contains"},

	"info.platform": {"contains", "eq", "ne"},
	"identity.mac":  {"contains", "eq", "ne"},
},
	"online",
)

// DeviceSortFields is the set of field names accepted in the sort_by query
// parameter when listing devices.
var DeviceSortFields = query.NewFieldSet(
	"name",
	"status",
	"last_seen",
	"created_at",
)

type DeviceService interface {
	ListDevices(ctx context.Context, sc scope.Scope, req *requests.DeviceList) ([]models.Device, int, error)

	// GetDevice fetches a device within the given namespace scope. The scope is an explicit
	// parameter rather than something recovered from the request context, so a caller cannot
	// receive a cross-namespace read by omission.
	GetDevice(ctx context.Context, sc scope.Scope, uid models.UID) (*models.Device, error)

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

	// SetDeviceCustomField sets or updates a single custom_fields entry on the device.
	// It enforces the per-device entry cap defined by [maxCustomFieldsPerDevice] when
	// adding a new key.
	SetDeviceCustomField(ctx context.Context, req *requests.DeviceSetCustomField) error
	// DeleteDeviceCustomField removes a single custom_fields entry from the device.
	// It is idempotent: removing a non-existent key is not an error.
	DeleteDeviceCustomField(ctx context.Context, req *requests.DeviceDeleteCustomField) error
}

func (s *service) deviceLimit(ctx context.Context, tenantID string) (models.NamespaceDeviceLimit, error) {
	if limit, ok := authctx.NamespaceDeviceLimit(ctx, tenantID); ok {
		return limit, nil
	}

	return s.store.NamespaceGetDeviceLimit(ctx, tenantID)
}

func (s *service) ListDevices(ctx context.Context, sc scope.Scope, req *requests.DeviceList) ([]models.Device, int, error) {
	opts := []store.QueryOption{}

	if req.DeviceStatus != "" {
		opts = append(opts, s.store.Options().WithDeviceStatus(req.DeviceStatus))
	}

	if req.Sorter.By == "" {
		req.Sorter.By = "last_seen"
	}

	req.Sorter.Tiebreak = "id"

	opts = append(opts, s.store.Options().Match(&req.Filters), s.store.Options().Sort(&req.Sorter), s.store.Options().Paginate(&req.Paginator))

	if req.DeviceStatus == models.DeviceStatusRemoved {
		return s.store.DeviceList(ctx, sc, store.DeviceAcceptableFromRemoved, opts...)
	}

	acceptable := store.DeviceAcceptableIfNotAccepted

	if sc.IsBounded() {
		limit, err := s.deviceLimit(ctx, req.TenantID)
		if err != nil {
			return nil, 0, NewErrNamespaceNotFound(req.TenantID, err)
		}

		if limit.HasMax() && limit.IsReached() {
			acceptable = store.DeviceAcceptableAsFalse
		}
	}

	return s.store.DeviceList(ctx, sc, acceptable, opts...)
}

func (s *service) GetDevice(ctx context.Context, sc scope.Scope, uid models.UID) (*models.Device, error) {
	device, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(uid))
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

	sc := scope.MustBounded(n.TenantID)

	var device *models.Device
	switch {
	case req.UID != "":
		device, err = s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, req.UID)
	case req.Hostname != "":
		device, err = s.store.DeviceResolve(ctx, sc, store.DeviceHostnameResolver, req.Hostname)
	}

	if err != nil {
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
	sc, err := BoundTo(tenant)
	if err != nil {
		return err
	}

	device, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(uid))
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	if device.Status == models.DeviceStatusAccepted {
		now := clock.Now()

		deviceCopy := *device
		deviceCopy.Status = models.DeviceStatusRemoved
		deviceCopy.RemovedAt = &now
		if err := s.store.DeviceUpdate(ctx, &deviceCopy); err != nil {
			return err
		}

		if err := s.store.NamespaceIncrementDeviceCount(ctx, sc, models.DeviceStatusRemoved, 1); err != nil {
			return err
		}
	} else {
		if err := s.store.DeviceDelete(ctx, device); err != nil {
			return err
		}
	}

	if err := s.store.NamespaceIncrementDeviceCount(ctx, sc, device.Status, -1); err != nil {
		return err
	}

	return nil
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error {
	sc, err := BoundTo(tenant)
	if err != nil {
		return err
	}

	device, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(uid))
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	if strings.EqualFold(device.Name, name) {
		return nil
	}

	device.Name = strings.ToLower(name)
	if err := s.store.DeviceUpdate(ctx, device); err != nil {
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

	device, err := s.store.DeviceResolve(ctx, scope.MustBounded(n.TenantID), store.DeviceHostnameResolver, name,
		s.store.Options().WithDeviceStatus(models.DeviceStatusAccepted))
	if err != nil || device == nil {
		return nil, NewErrDeviceNotFound(models.UID(name), err)
	}

	return device, nil
}

func (s *service) OfflineDevice(ctx context.Context, uid models.UID) error {
	if err := s.store.DeviceOffline(ctx, string(uid), clock.Now()); err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrDeviceNotFound(uid, err)
		}

		return err
	}

	return nil
}

func (s *service) UpdateDeviceStatus(ctx context.Context, req *requests.DeviceUpdateStatus) error {
	if err := s.store.WithTransaction(ctx, s.updateDeviceStatus(req)); err != nil {
		return err
	}

	status := models.DeviceStatus(req.Status)
	if status == models.DeviceStatusAccepted || status == models.DeviceStatusRejected {
		sc, err := BoundTo(req.TenantID)
		if err != nil {
			return err
		}

		if err := s.store.InstallKeyEventStampDecision(ctx, sc, req.UID, status, clock.Now()); err != nil {
			log.WithError(err).WithFields(log.Fields{"device_uid": req.UID, "status": req.Status}).
				Warn("failed to stamp the enrollment decision on the history event")
		}
	}

	return nil
}

func (s *service) updateDeviceStatus(req *requests.DeviceUpdateStatus) store.TransactionCb {
	return func(ctx context.Context) error {
		namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
		if err != nil {
			return NewErrNamespaceNotFound(req.TenantID, err)
		}

		sc := scope.MustBounded(namespace.TenantID)

		device, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, req.UID)
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
			opts := []store.QueryOption{s.store.Options().WithDeviceStatus(models.DeviceStatusAccepted)}
			existingMacDevice, err := s.store.DeviceResolve(ctx, sc, store.DeviceMACResolver, device.Identity.MAC, opts...)
			if err != nil && !errors.Is(err, store.ErrNoDocuments) {
				log.WithError(err).
					WithFields(log.Fields{"mac": device.Identity.MAC}).
					Error("failed to retrieve device using MAC")

				return err
			}

			if existingMacDevice != nil && existingMacDevice.UID != device.UID {
				existingNameDevice, err := s.store.DeviceResolve(ctx, sc, store.DeviceHostnameResolver, device.Name, opts...)
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
				existingDevice, err := s.store.DeviceResolve(ctx, sc, store.DeviceHostnameResolver, device.Name, opts...)
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

				if err := s.validateDeviceAcceptance(ctx, namespace); err != nil {
					return err
				}
			}
		}

		device.Status = newStatus
		device.StatusUpdatedAt = clock.Now()
		if err := s.store.DeviceUpdate(ctx, device); err != nil {
			if newStatus == models.DeviceStatusAccepted && errors.Is(err, store.ErrDuplicate) {
				return NewErrDeviceDuplicated(device.Name, err)
			}

			return err
		}

		for status, count := range map[models.DeviceStatus]int64{oldStatus: -1, newStatus: 1} {
			if err := s.store.NamespaceIncrementDeviceCount(ctx, sc, status, count); err != nil {
				return err
			}
		}

		return nil
	}
}

func (s *service) UpdateDevice(ctx context.Context, req *requests.DeviceUpdate) error {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	device, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, req.UID)
	if err != nil {
		return NewErrDeviceNotFound(models.UID(req.UID), err)
	}

	if req.Name == "" || strings.EqualFold(req.Name, device.Name) {
		return nil
	}

	conflictsTarget := &models.DeviceConflicts{Name: strings.ToLower(req.Name)}
	conflictsTarget.Distinct(device)
	if _, has, err := s.store.DeviceConflicts(ctx, sc, conflictsTarget); err != nil || has {
		return NewErrDeviceDuplicated(req.Name, err)
	}

	device.Name = strings.ToLower(req.Name)

	if err := s.store.DeviceUpdate(ctx, device); err != nil {
		return err
	}

	return nil
}

const maxCustomFieldsPerDevice = 20

func (s *service) SetDeviceCustomField(ctx context.Context, req *requests.DeviceSetCustomField) error {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	device, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, req.UID)
	if err != nil {
		return NewErrDeviceNotFound(models.UID(req.UID), err)
	}

	if _, exists := device.CustomFields[req.Key]; !exists && len(device.CustomFields) >= maxCustomFieldsPerDevice {
		return NewErrDeviceCustomFieldLimitReached(maxCustomFieldsPerDevice, nil)
	}

	if err := s.store.DeviceSetCustomField(ctx, req.UID, req.Key, req.Value); err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteDeviceCustomField(ctx context.Context, req *requests.DeviceDeleteCustomField) error {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	if _, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, req.UID); err != nil {
		return NewErrDeviceNotFound(models.UID(req.UID), err)
	}

	if err := s.store.DeviceDeleteCustomField(ctx, req.UID, req.Key); err != nil {
		return err
	}

	return nil
}

func (s *service) mergeDevice(ctx context.Context, tenantID string, oldDevice *models.Device, newDevice *models.Device) error {
	logFields := log.Fields{"tenant_id": tenantID, "old_device_uid": oldDevice.UID, "new_device_uid": newDevice.UID}

	log.WithFields(logFields).Debug("firing device merge hooks")
	if err := fireDeviceMerge(ctx, tenantID, oldDevice, newDevice); err != nil {
		log.WithError(err).WithFields(logFields).Error("device merge hook failed")

		return err
	}

	log.WithFields(logFields).Debug("transferring sessions from old device to new device")
	if err := s.store.SessionUpdateDeviceUID(ctx, models.UID(oldDevice.UID), models.UID(newDevice.UID)); err != nil && !errors.Is(err, store.ErrNoDocuments) {
		log.WithError(err).WithFields(logFields).Error("failed to transfer sessions")

		return err
	}

	log.WithFields(logFields).Debug("updating new device name to preserve old device identity")
	newDevice.Name = oldDevice.Name
	if err := s.store.DeviceUpdate(ctx, newDevice); err != nil {
		log.WithError(err).WithFields(logFields).Error("failed to update new device name")

		return err
	}

	log.WithFields(logFields).Debug("mergeDevice: deleting old device")
	if err := s.store.DeviceDelete(ctx, oldDevice); err != nil {
		log.WithError(err).WithFields(logFields).Error("failed to delete old device")

		return err
	}

	if err := s.store.NamespaceIncrementDeviceCount(ctx, scope.MustBounded(tenantID), oldDevice.Status, -1); err != nil {
		log.WithError(err).WithFields(logFields).Error("failed to decrement namespace device count")

		return err
	}

	log.WithFields(logFields).Info("device merge operation completed successfully")

	return nil
}

func (s *service) validateDeviceAcceptance(ctx context.Context, namespace *models.Namespace) error {
	if namespace.HasMaxDevices() && namespace.HasMaxDevicesReached() {
		if envs.IsCloud() && (namespace.Billing == nil || !namespace.Billing.IsActive()) {
			log.WithFields(log.Fields{"tenant": namespace.TenantID}).
				Error("namespace's limit reached - cannot accept another device")

			return NewErrDeviceLimit(namespace.MaxDevices, nil)
		}

		return NewErrDeviceMaxDevicesReached(namespace.MaxDevices)
	}

	if envs.IsCloud() && s.billing != nil {
		if err := s.validateBillingForDeviceAcceptance(ctx, namespace); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"tenant":         namespace.TenantID,
				"billing_active": namespace.Billing.IsActive(),
			}).Error("billing validation failed")

			return err
		}
	}

	if s.licenseEvaluator != nil {
		ok, err := s.licenseEvaluator.CanAcceptDevice(ctx)
		if err != nil {
			log.WithError(err).Warn("license evaluator returned an error; failing open and allowing device acceptance")

			return nil
		}

		if !ok {
			return ErrDeviceLicenseLimit
		}
	}

	return nil
}
