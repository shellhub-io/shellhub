package store

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type DeviceAcceptable uint

const (
	// DeviceAcceptableIfNotAccepted is used to indicate the all devices not accepted will be defined as "acceptabled".
	DeviceAcceptableIfNotAccepted DeviceAcceptable = iota + 1
	// DeviceAcceptableFromRemoved is used to indicate that the namepsace's device maxium number of devices has been
	// reached and should set the "acceptable" value to true for devices that were recently removed.
	DeviceAcceptableFromRemoved
	// DeviceAcceptableAsFalse set acceptable to false to all returned devices.
	DeviceAcceptableAsFalse
)

type DeviceIdent string

const (
	DeviceIdentUID  DeviceIdent = "id"
	DeviceIdentName DeviceIdent = "name"
)

type DeviceStore interface {
	DeviceCreate(ctx context.Context, device *models.Device) (string, error)
	DeviceConflicts(ctx context.Context, target *models.DeviceConflicts) ([]string, bool, error)
	DeviceList(ctx context.Context, opts ...QueryOption) ([]models.Device, int, error)
	DeviceGet(ctx context.Context, ident DeviceIdent, val string) (*models.Device, error)
	DeviceSave(ctx context.Context, device *models.Device) error
	DeviceUpdateSeenAt(ctx context.Context, ids []string, to time.Time) (int64, error)
	DeviceDelete(ctx context.Context, device *models.Device) error
}
