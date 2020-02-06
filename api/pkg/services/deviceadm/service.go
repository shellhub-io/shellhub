package deviceadm

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/models"
	"github.com/shellhub-io/shellhub/api/pkg/store"
)

type Service interface {
	ListDevices(ctx context.Context) ([]models.Device, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID) error
	RenameDevice(ctx context.Context, uid models.UID, name string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) ListDevices(ctx context.Context) ([]models.Device, error) {
	return s.store.ListDevices(ctx)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	return s.store.GetDevice(ctx, uid)
}

func (s *service) DeleteDevice(ctx context.Context, uid models.UID) error {
	return s.store.DeleteDevice(ctx, uid)
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name string) error {
	return s.store.RenameDevice(ctx, uid, name)
}

func (s *service) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	return s.store.LookupDevice(ctx, namespace, name)
}

func (s *service) UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error {
	return s.store.UpdateDeviceStatus(ctx, uid, online)
}
