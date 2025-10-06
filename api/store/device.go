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

type DeviceResolver uint

const (
	DeviceUIDResolver DeviceResolver = iota + 1
	DeviceHostnameResolver
	DeviceMACResolver
)

type DeviceStore interface {
	// DeviceCreate creates a new device. It returns the inserted UID and an error, if any.
	DeviceCreate(ctx context.Context, device *models.Device) (insertedUID string, err error)

	DeviceList(ctx context.Context, acceptable DeviceAcceptable, opts ...QueryOption) ([]models.Device, int, error)

	// DeviceResolve fetches a device using a specific resolver within a given tenant ID.
	//
	// It returns the resolved device if found and an error, if any.
	DeviceResolve(ctx context.Context, resolver DeviceResolver, value string, opts ...QueryOption) (*models.Device, error)

	// DeviceConflicts reports whether the target contains conflicting attributes with the database. Pass zero values for
	// attributes you do not wish to match on. For example, the following call checks for conflicts based on email only:
	//
	//  ctx := context.Background()
	//  conflicts, has, err := store.DeviceConflicts(ctx, &models.DeviceConflicts{Name: "mydevice"})
	//
	// It returns an array of conflicting attribute fields and an error, if any.
	DeviceConflicts(ctx context.Context, target *models.DeviceConflicts) (conflicts []string, has bool, err error)

	// DeviceUpdate updates a device. It returns [ErrNoDocuments] if none device is found.
	DeviceUpdate(ctx context.Context, device *models.Device) error
	// DeviceHeartbeat updates the last_seen timestamp and sets disconnected_at to nil for multiple devices.
	// It returns the number of modified devices and an error if any.
	DeviceHeartbeat(ctx context.Context, uids []string, lastSeen time.Time) (modifiedCount int64, err error)

	DeviceDelete(ctx context.Context, device *models.Device) error
	// DeviceDeleteMany deletes multiple devices by their UIDs.
	DeviceDeleteMany(ctx context.Context, uids []string) (deletedCount int64, err error)
}
