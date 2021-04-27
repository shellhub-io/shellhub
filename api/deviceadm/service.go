package deviceadm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"gopkg.in/go-playground/validator.v9"
)

var (
	ErrUnauthorized          = errors.New("unauthorized")
	ErrMaxDeviceCountReached = errors.New("maximum number of accepted devices reached")
	ErrDuplicatedDeviceName  = errors.New("the name already exists in the namespace")
)

type Service interface {
	ListDevices(ctx context.Context, pagination paginator.Query, filter string, status string, sort string, order string) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID, tenant, username string) error
	RenameDevice(ctx context.Context, uid models.UID, name, tenant, username string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error
	UpdatePendingStatus(ctx context.Context, uid models.UID, status, tenant, username string) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) isNamespaceOnwer(ctx context.Context, tenant, username string) error {
	namespace, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil {
		return err
	}

	user, err := s.store.UserGetByUsername(ctx, username)
	if err != nil {
		return err
	}

	if user.ID != namespace.Owner {
		return ErrUnauthorized
	}
	return nil
}

func (s *service) ListDevices(ctx context.Context, pagination paginator.Query, filterB64 string, status string, sort string, order string) ([]models.Device, int, error) {
	raw, err := base64.StdEncoding.DecodeString(filterB64)
	if err != nil {
		return nil, 0, err
	}

	var filter []models.Filter

	if err := json.Unmarshal([]byte(raw), &filter); len(raw) > 0 && err != nil {
		return nil, 0, err
	}

	return s.store.DeviceList(ctx, pagination, filter, status, sort, order)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	return s.store.DeviceGet(ctx, uid)
}

func (s *service) DeleteDevice(ctx context.Context, uid models.UID, tenant, username string) error {
	if err := s.isNamespaceOnwer(ctx, tenant, username); err != nil {
		return ErrUnauthorized
	}

	if _, err := s.store.DeviceGetByUID(ctx, uid, tenant); err != nil {
		return err
	}

	return s.store.DeviceDelete(ctx, uid)
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name, tenant, username string) error {
	if err := s.isNamespaceOnwer(ctx, tenant, username); err != nil {
		return ErrUnauthorized
	}

	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
	if err != nil {
		return err
	}

	name = strings.ToLower(name)
	if device.Name == name {
		return nil
	}

	validate := validator.New()
	err = validate.Struct(device)
	if err != nil {
		return err
	}

	otherDevice, err := s.store.DeviceGetByName(ctx, name, tenant)
	if err != nil && err != store.ErrDeviceNoDocuments {
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

func (s *service) UpdatePendingStatus(ctx context.Context, uid models.UID, status, tenant, username string) error {
	if err := s.isNamespaceOnwer(ctx, tenant, username); err != nil {
		return ErrUnauthorized
	}

	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
	if err != nil {
		return err
	}

	if status == "accepted" {
		sameMacDev, err := s.store.DeviceGetByMac(ctx, device.Identity.MAC, device.TenantID, "accepted")
		if err != nil && err != store.ErrDeviceNoDocuments {
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
