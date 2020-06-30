package deviceadm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"gopkg.in/go-playground/validator.v9"
)

var ErrUnauthorized = errors.New("unauthorized")

type Service interface {
	ListDevices(ctx context.Context, pagination paginator.Query, filter string, status string) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID, tenant string) error
	RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error
	UpdatePendingStatus(ctx context.Context, uid models.UID, status string) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) ListDevices(ctx context.Context, pagination paginator.Query, filterB64 string, status string) ([]models.Device, int, error) {
	raw, err := base64.StdEncoding.DecodeString(filterB64)
	if err != nil {
		return nil, 0, err
	}

	var filter []models.Filter

	if err := json.Unmarshal([]byte(raw), &filter); len(raw) > 0 && err != nil {
		return nil, 0, err
	}

	return s.store.ListDevices(ctx, pagination, filter, status)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	return s.store.GetDevice(ctx, uid)
}

func (s *service) DeleteDevice(ctx context.Context, uid models.UID, tenant string) error {
	device, _ := s.store.GetDeviceByUID(ctx, uid, tenant)
	if device != nil {
		return s.store.DeleteDevice(ctx, uid)
	}
	return ErrUnauthorized
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error {
	device, _ := s.store.GetDeviceByUID(ctx, uid, tenant)
	validate := validator.New()
	if device != nil {
		if device.Name != name {
			device.Name = name
			if err := validate.Struct(device); err == nil {
				otherDevice, _ := s.store.GetDeviceByName(ctx, name, tenant)
				if otherDevice == nil {
					return s.store.RenameDevice(ctx, uid, name)
				}
			}
		}
	}
	return ErrUnauthorized
}

func (s *service) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	return s.store.LookupDevice(ctx, namespace, name)
}

func (s *service) UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error {
	return s.store.UpdateDeviceStatus(ctx, uid, online)
}

func (s *service) UpdatePendingStatus(ctx context.Context, uid models.UID, status string) error {
	return s.store.UpdatePendingStatus(ctx, uid, status)
}
