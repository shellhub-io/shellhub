package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
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

	DeviceList(ctx context.Context, status models.DeviceStatus, pagination query.Paginator, filters query.Filters, sorter query.Sorter, acceptable DeviceAcceptable) ([]models.Device, int, error)

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

	// DeviceUpdate updates a device with the specified UID that belongs to the specified namespace. It returns [ErrNoDocuments] if none device is found.
	DeviceUpdate(ctx context.Context, tenant, uid string, changes *models.DeviceChanges) error
	// DeviceBulkdUpdate updates a list of devices. Different than [DeviceStore.DeviceUpdate], it does not differentiate namespaces.
	// It returns the number of  modified devices and an error if any.
	DeviceBulkUpdate(ctx context.Context, uids []string, changes *models.DeviceChanges) (modifiedCount int64, err error)

	DeviceDelete(ctx context.Context, uid models.UID) error
	DeviceRename(ctx context.Context, uid models.UID, hostname string) error
	DeviceUpdateStatus(ctx context.Context, uid models.UID, status models.DeviceStatus) error
}
