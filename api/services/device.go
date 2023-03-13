package services

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	req "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/sirupsen/logrus"
)

const StatusAccepted = "accepted"

type DeviceService interface {
	ListDevices(ctx context.Context, tenant string, pagination paginator.Query, filter []models.Filter, status models.DeviceStatus, sort, order string) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	GetDeviceByPublicURLAddress(ctx context.Context, address string) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID, tenant string) error
	RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error
	UpdatePendingStatus(ctx context.Context, uid models.UID, status models.DeviceStatus, tenant string) error
	SetDevicePosition(ctx context.Context, uid models.UID, ip string) error
	DeviceHeartbeat(ctx context.Context, uid models.UID) error
	UpdateDevice(ctx context.Context, tenant string, uid models.UID, name *string, publicURL *bool) error
}

func (s *service) ListDevices(ctx context.Context, tenant string, pagination paginator.Query, filter []models.Filter, status models.DeviceStatus, sort, order string) ([]models.Device, int, error) {
	switch status {
	case models.DeviceStatusPending, models.DeviceStatusRejected:
		ns, err := s.store.NamespaceGet(ctx, tenant)
		if err != nil {
			return nil, 0, NewErrNamespaceNotFound(tenant, err)
		}

		count, err := s.store.DeviceRemovedCount(ctx, ns.TenantID)
		if err != nil {
			return nil, 0, NewErrDeviceRemovedCount(err)
		}

		if ns.HasMaxDevicesReached(count) {
			return s.store.DeviceList(ctx, pagination, filter, status, sort, order, store.DeviceListModeMaxDeviceReached)
		}
	case models.DeviceStatusRemoved:
		removed, count, err := s.store.DeviceRemovedList(ctx, tenant, pagination, filter, sort, order)
		if err != nil {
			return nil, 0, err
		}

		devices := make([]models.Device, 0, len(removed))
		for _, device := range removed {
			devices = append(devices, *device.Device)
		}

		return devices, count, nil
	}

	return s.store.DeviceList(ctx, pagination, filter, status, sort, order, store.DeviceListModeDefault)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil {
		return nil, NewErrDeviceNotFound(uid, err)
	}

	return device, nil
}

func (s *service) GetDeviceByPublicURLAddress(ctx context.Context, address string) (*models.Device, error) {
	device, err := s.store.DeviceGetByPublicURLAddress(ctx, address)
	if err != nil {
		return nil, NewErrDeviceNotFound(models.UID(address), err)
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
	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
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
	if envs.IsCloud() && ns.HasMaxDevices() {
		if err := s.store.DeviceRemovedInsert(ctx, tenant, device); err != nil {
			return NewErrDeviceRemovedInsert(err)
		}
	}

	if err = createReportUsage(s.client.(req.Client), ns, false, device); err != nil {
		return err
	}

	return s.store.DeviceDelete(ctx, uid)
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error {
	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
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
		PublicURL:  false,
	}

	if data, err := validator.ValidateStructFields(updatedDevice); err != nil {
		return NewErrDeviceInvalid(data, err)
	}

	if device.Name == updatedDevice.Name {
		return nil
	}

	otherDevice, err := s.store.DeviceGetByName(ctx, updatedDevice.Name, tenant)
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
	device, err := s.store.DeviceLookup(ctx, namespace, name)
	if err != nil || device == nil {
		return nil, NewErrDeviceLookupNotFound(namespace, name, err)
	}

	return device, nil
}

func (s *service) UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error {
	err := s.store.DeviceSetOnline(ctx, uid, online)
	if err == store.ErrNoDocuments {
		return NewErrDeviceNotFound(uid, err)
	}

	return err
}

func (s *service) UpdatePendingStatus(ctx context.Context, uid models.UID, status models.DeviceStatus, tenant string) error {
	validateStatus := map[string]bool{
		"accepted": true,
		"pending":  true,
		"rejected": true,
	}

	if _, ok := validateStatus[string(status)]; !ok {
		return NewErrDeviceStatusInvalid(string(status), nil)
	}

	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	if device.Status == models.DeviceStatusAccepted {
		return NewErrDeviceStatusAccepted(nil)
	}

	ns, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil {
		return NewErrNamespaceNotFound(tenant, err)
	}

	if status != models.DeviceStatusAccepted {
		return s.store.DeviceUpdateStatus(ctx, uid, status)
	}

	// NOTICE: The logic below is only executed when the new status is "accepted".

	if envs.IsCloud() && ns.HasMaxDevices() {
		removed, err := s.store.DeviceRemovedGet(ctx, tenant, uid)
		if err != nil && err != store.ErrNoDocuments {
			return NewErrDeviceRemovedGet(err)
		}

		if removed != nil {
			if err := s.store.DeviceRemovedDelete(ctx, tenant, uid); err != nil {
				return NewErrDeviceRemovedDelete(err)
			}
		} else {
			count, err := s.store.DeviceRemovedCount(ctx, tenant)
			if err != nil {
				return NewErrDeviceRemovedCount(err)
			}

			if ns.HasMaxDevicesReached(count) {
				return NewErrDeviceRemovedFull(ns.MaxDevices, nil)
			}
		}
	}

	sameMacDev, err := s.store.DeviceGetByMac(ctx, device.Identity.MAC, device.TenantID, "accepted")
	if err != nil && err != store.ErrNoDocuments {
		return NewErrDeviceNotFound(models.UID(device.UID), err)
	}

	if sameMacDev != nil && sameMacDev.UID != device.UID { //nolint:nestif
		// TODO: decide what to do with these errors.
		if err := s.store.SessionUpdateDeviceUID(ctx, models.UID(sameMacDev.UID), models.UID(device.UID)); err != nil {
			return err
		}
		if err := s.store.DeviceDelete(ctx, models.UID(sameMacDev.UID)); err != nil {
			return err
		}
		if err := s.store.DeviceRename(ctx, models.UID(device.UID), sameMacDev.Name); err != nil {
			return err
		}
	} else {
		ns, err := s.store.NamespaceGet(ctx, device.TenantID)
		if err != nil {
			return NewErrNamespaceNotFound(device.TenantID, err)
		}

		if err := createReportUsage(s.client.(req.Client), ns, true, device); err != nil {
			return err
		}

		if ns.MaxDevices > 0 && ns.MaxDevices <= ns.DevicesCount {
			return NewErrDeviceLimit(ns.MaxDevices, nil)
		}
	}

	return s.store.DeviceUpdateStatus(ctx, uid, status)
}

// SetDevicePosition sets the position to a device from its IP.
func (s *service) SetDevicePosition(ctx context.Context, uid models.UID, ip string) error {
	ipParsed := net.ParseIP(ip)
	position, err := s.locator.GetPosition(ipParsed)
	if err != nil {
		logrus.
			WithError(err).
			WithFields(logrus.Fields{
				"uid": uid,
				"ip":  ip,
			}).Error("Failed to get device's position")
	}

	devicePosition := models.DevicePosition{
		Longitude: position.Longitude,
		Latitude:  position.Latitude,
	}

	err = s.store.DeviceSetPosition(ctx, uid, devicePosition)
	if err != nil {
		logrus.
			WithError(err).
			WithFields(logrus.Fields{
				"uid": uid,
				"ip":  ip,
			}).Error("Failed to set device's position to database")

		return err
	}
	logrus.WithFields(logrus.Fields{
		"uid":      uid,
		"ip":       ip,
		"position": position,
	}).Debug("Success to set device's position")

	return nil
}

func (s *service) DeviceHeartbeat(ctx context.Context, uid models.UID) error {
	if err := s.store.DeviceSetOnline(ctx, uid, true); err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	return nil
}

func (s *service) UpdateDevice(ctx context.Context, tenant string, uid models.UID, name *string, publicURL *bool) error {
	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	if name != nil {
		*name = strings.ToLower(*name)

		if device.Name == *name {
			return nil
		}

		otherDevice, err := s.store.DeviceGetByName(ctx, *name, tenant)
		if err != nil && err != store.ErrNoDocuments {
			return NewErrDeviceNotFound(models.UID(device.UID), fmt.Errorf("failed to get device by name: %w", err))
		}

		if otherDevice != nil {
			return NewErrDeviceDuplicated(otherDevice.Name, err)
		}
	}

	if publicURL != nil {
		if device.PublicURLAddress == "" && *publicURL {
			if err := s.store.DeviceCreatePublicURLAddress(ctx, models.UID(device.UID)); err != nil {
				return err
			}
		}
	}

	return s.store.DeviceUpdate(ctx, uid, name, publicURL)
}
