package deviceadm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/shellhub-io/shellhub/api/pkg/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

var UnauthorizedErr = errors.New("unauthorized")

type Service interface {
	CountDevices(ctx context.Context) (int64, error)
	ListDevices(ctx context.Context, perPage int, page int, filter string) ([]models.Device, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID, tenant string) error
	RenameDevice(ctx context.Context, uid models.UID, name string, tenant string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}
func (s *service) CountDevices(ctx context.Context) (int64, error) {
	return s.store.CountDevices(ctx)
}

func (s *service) ListDevices(ctx context.Context, perPage int, page int, filterB64 string) ([]models.Device, error) {

	raw, err := base64.StdEncoding.DecodeString(filterB64)
	if err != nil {
		return err
	}

	var filter []models.Filter

	if err := json.Unmarshal([]byte(raw), &filter); err != nil {
		panic(err)
	}

	return s.store.ListDevices(ctx, perPage, page, filter)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	return s.store.GetDevice(ctx, uid)
}

func (s *service) DeleteDevice(ctx context.Context, uid models.UID, tenant string) error {
	device, _ := s.store.GetDeviceByUid(ctx, uid, tenant)
	if device != nil {
		return s.store.DeleteDevice(ctx, uid)
	}
	return UnauthorizedErr
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name string, tenant string) error {
	device, _ := s.store.GetDeviceByUid(ctx, uid, tenant)
	if device != nil {
		if device.Name != name {
			return s.store.RenameDevice(ctx, uid, name)
		}
	}
	return UnauthorizedErr
}

func (s *service) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	return s.store.LookupDevice(ctx, namespace, name)
}

func (s *service) UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error {
	return s.store.UpdateDeviceStatus(ctx, uid, online)
}
