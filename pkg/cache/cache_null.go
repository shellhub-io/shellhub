package cache

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type nullCache struct{}

var _ Cache = &nullCache{}

func NewNullCache() Cache {
	return &nullCache{}
}

func (n *nullCache) Get(_ context.Context, _ string, _ interface{}) error {
	return nil
}

func (n *nullCache) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
	return nil
}

func (n *nullCache) Delete(_ context.Context, _ string) error {
	return nil
}

func (n *nullCache) GetLastSeen(_ context.Context, _, _ string) (time.Time, bool, error) {
	return time.Time{}, true, nil
}

func (n *nullCache) SetLastSeen(_ context.Context, _, _ string, _ time.Time) error {
	return nil
}

func (n *nullCache) DelLastSeen(_ context.Context, _, _ string) error {
	return nil
}

func (n *nullCache) CountConnectedDevices(_ context.Context, _ string, _ models.DeviceStatus) (int64, error) {
	return 0, nil
}
func (n *nullCache) IncreaseConnectedDevices(_ context.Context, _ string, _ models.DeviceStatus, _ int) error {
	return nil
}

func (n *nullCache) DecreaseConnectedDevices(_ context.Context, _ string, _ models.DeviceStatus, _ int) error {
	return nil
}
