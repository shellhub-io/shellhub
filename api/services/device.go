package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	utils "github.com/shellhub-io/shellhub/api/pkg/namespace"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

func (s *service) ListDevices(ctx context.Context, pagination paginator.Query, filterB64 string, status string, sort string, order string) ([]models.Device, int, error) {
	raw, err := base64.StdEncoding.DecodeString(filterB64)
	if err != nil {
		return nil, 0, err
	}

	var filter []models.Filter
	if err := json.Unmarshal(raw, &filter); len(raw) > 0 && err != nil {
		return nil, 0, err
	}

	return s.store.DeviceList(ctx, pagination, filter, status, sort, order)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	return s.store.DeviceGet(ctx, uid)
}

func (s *service) DeleteDevice(ctx context.Context, uid models.UID, tenant, ownerID string) error {
	if err := utils.IsNamespaceOwner(ctx, s.store, tenant, ownerID); err != nil {
		return ErrUnauthorized
	}

	if _, err := s.store.DeviceGetByUID(ctx, uid, tenant); err != nil {
		return err
	}

	return s.store.DeviceDelete(ctx, uid)
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name, tenant, ownerID string) error {
	if err := utils.IsNamespaceOwner(ctx, s.store, tenant, ownerID); err != nil {
		return ErrUnauthorized
	}

	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
	if err != nil {
		return err
	}

	updatedDevice := &models.Device{
		UID:       device.UID,
		Name:      strings.ToLower(name),
		Identity:  device.Identity,
		Info:      device.Info,
		PublicKey: device.PublicKey,
		TenantID:  device.TenantID,
		LastSeen:  device.LastSeen,
		Online:    device.Online,
		Namespace: device.Namespace,
		Status:    device.Status,
	}

	if _, err = validator.ValidateStruct(updatedDevice); err != nil {
		return err
	}

	if device.Name == updatedDevice.Name {
		return nil
	}

	otherDevice, err := s.store.DeviceGetByName(ctx, updatedDevice.Name, tenant)
	if err != nil && err != store.ErrNoDocuments {
		return err
	}

	if otherDevice != nil {
		return ErrDuplicatedDeviceName
	}

	return s.store.DeviceRename(ctx, uid, name)
}

func (s *service) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	return s.store.DeviceLookup(ctx, namespace, name)
}

func (s *service) UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error {
	return s.store.DeviceSetOnline(ctx, uid, online)
}

func (s *service) UpdatePendingStatus(ctx context.Context, uid models.UID, status, tenant, ownerID string) error {
	if err := utils.IsNamespaceOwner(ctx, s.store, tenant, ownerID); err != nil {
		return ErrUnauthorized
	}

	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
	if err != nil {
		return err
	}

	if status == "accepted" { //nolint:nestif
		sameMacDev, err := s.store.DeviceGetByMac(ctx, device.Identity.MAC, device.TenantID, "accepted")
		if err != nil && err != store.ErrNoDocuments {
			return err
		}

		if sameMacDev != nil && sameMacDev.UID != device.UID {
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
				return err
			}

			if ns.MaxDevices > 0 && ns.MaxDevices <= ns.DevicesCount {
				return ErrMaxDeviceCountReached
			}
		}
	}

	return s.store.DeviceUpdateStatus(ctx, uid, status)
}
