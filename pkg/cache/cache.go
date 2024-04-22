package cache

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type Cache interface {
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error

	// CountConnectedDevices counts the number of devices currently connected and having the specified status
	// for a given tenant. Use tenant="*" to count in all available tenants.
	CountConnectedDevices(ctx context.Context, tenant string, status models.DeviceStatus) (int64, error)

	// IncreaseConnectedDevices increases the count t times of currently connected devices with the specified status for
	// a given tenant by a specified value.
	IncreaseConnectedDevices(ctx context.Context, tenant string, status models.DeviceStatus, t int) error

	// DecreaseConnectedDevices decreases the count t times of currently connected devices with the specified status for
	// a given tenant by a specified value.
	DecreaseConnectedDevices(ctx context.Context, tenant string, status models.DeviceStatus, t int) error

	// GetLastSeen retrieves the last_seen associated with the provided tenant and uid.
	//
	// The last_seen is used to determine whether the device corresponding to the key is currently online.
	// If the key is not related to any timestamp, it false.
	GetLastSeen(ctx context.Context, tenant, uid string) (time.Time, bool, error)

	// SetLastSeen sets the last_seen tt for the provided tenant and uid.
	SetLastSeen(ctx context.Context, tenant, uid string, tt time.Time) error

	// DelLastSeen deletes the last_seen timestamp associated with the provided tenant and uid. It can be used to force
	// the device's status to offline.
	DelLastSeen(ctx context.Context, tenant, uid string) error
}
