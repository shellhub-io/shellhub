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
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/sirupsen/logrus"
)

const StatusAccepted = "accepted"

type DeviceService interface {
	ListDevices(ctx context.Context, pagination paginator.Query, filter []models.Filter, status string, sort string, order string) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID, tenant string) error
	RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error
	UpdatePendingStatus(ctx context.Context, uid models.UID, status, tenant string) error
	SetDevicePosition(ctx context.Context, uid models.UID, ip string) error
	DeviceHeartbeat(ctx context.Context, uid models.UID) error
	UpdateDevice(ctx context.Context, tenant string, uid models.UID, name *string, publicURL *bool) error
}

func (s *service) ListDevices(ctx context.Context, pagination paginator.Query, filter []models.Filter, status string, sort string, order string) ([]models.Device, int, error) {
	return s.store.DeviceList(ctx, pagination, filter, status, sort, order)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	device, err := s.store.DeviceGet(ctx, uid)
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
	if ns.MaxDevices > 0 {
		if err := s.store.SlotSet(ctx, tenant, models.UID(device.UID), "removed"); err != nil {
			return NewErrSlotSet(err)
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

func (s *service) UpdatePendingStatus(ctx context.Context, uid models.UID, status, tenant string) error {
	validateStatus := map[string]bool{
		"accepted": true,
		"pending":  true,
		"rejected": true,
	}

	if _, ok := validateStatus[status]; !ok {
		return NewErrDeviceStatusInvalid(status, nil)
	}

	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	if device.Status == StatusAccepted {
		return NewErrDeviceStatusAccepted(nil)
	}

	ns, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil {
		return NewErrNamespaceNotFound(tenant, err)
	}

	// Here I introduce the concept of "slot", which is a way to control the number of devices that a namespace can
	// have in a given time and when there is a device limitation. The "slot" is a document that is created when a
	// device is "accepted" and its status change to "removed" when the device is deleted.
	// A "slot" with the status "removed" expires after a period of time, allowing a new device to be accepted at its
	// place.
	if ns.MaxDevices > 0 && status == StatusAccepted {
		// List all devices' slots.
		slots, err := s.store.SlotsList(ctx, tenant)
		if err != nil {
			return NewErrSlotList(err)
		}

		// When we set it to true, it means that the device's slot was updated either because it already exists or
		// another slot has expired.
		var set bool

		// NOTICE: I have those functions to improve the readability of the code, as I believe that it is easier to
		// understand what is going on when we have a function with a name that describes what it does.

		// isSlotRemoved checks if the slot's status is removed.
		isSlotRemoved := func(slot *models.Slot) bool {
			return slot.Status == "removed"
		}

		// isSlotAccepted checks if the slot's status is accepted.
		isSlotAccepted := func(slot *models.Slot) bool {
			return slot.Status == "accepted"
		}

		// isSlotExpired checks if the slot's has expired.
		// It is expired if the current time is after the slot's updated time plus the slot's duration.
		isSlotExpired := func(slot *models.Slot, duration time.Duration) bool {
			return time.Now().After(slot.UpdatedAt.Add(duration))
		}

		for _, slot := range slots {
			// Check if the device is into a slot.
			if models.UID(device.UID) == slot.UID {
				// If the device is into the slot and its status is accepted, we return an error.
				if isSlotAccepted(&slot) && status == StatusAccepted {
					return NewErrSlotOccupied(nil)
				}

				// Otherwise, if the device is into the slot, but the slot's status is removed, we update the slot
				// status to the new one.
				if isSlotRemoved(&slot) && status == StatusAccepted {
					if err := s.store.SlotSet(ctx, tenant, models.UID(device.UID), status); err != nil {
						return NewErrSlotSet(err)
					}

					// We set the variable to avoid creating a new slot for the device.
					set = true

					break
				}
			}

			// If the device is not into the slot, but there is a device's slot with its status as removed, and it is
			// expired, we need to delete the slot and create a new one for the device.
			// TODO: What could be the best way to get the slot's duration? A environment variable? A value hardcoded?
			if isSlotRemoved(&slot) && isSlotExpired(&slot, 20*time.Minute) && status == StatusAccepted {
				if err := s.store.SlotDelete(ctx, tenant, slot.UID); err != nil {
					return err
				}

				if err := s.store.SlotSet(ctx, tenant, models.UID(device.UID), status); err != nil {
					return NewErrSlotSet(err)
				}

				// We set the variable to avoid creating a new slot for the device.
				set = true

				break
			}
		}

		// If the device's slot does not apply to any of the above cases, we need to create a new slot for it.
		if !set {
			// The number of slots must be less than the maximum number of devices allowed to avoid device siting.
			if len(slots) >= ns.MaxDevices {
				return NewErrSlotsFull(nil, ns.MaxDevices)
			}

			if status == StatusAccepted {
				if err := s.store.SlotSet(ctx, tenant, models.UID(device.UID), status); err != nil {
					return NewErrSlotSet(err)
				}
			}
		}
	}

	if status != StatusAccepted {
		return s.store.DeviceUpdateStatus(ctx, uid, status)
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

	return s.store.DeviceUpdate(ctx, uid, name, publicURL)
}
