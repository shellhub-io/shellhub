package store

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// DeviceAcceptable selects how the acceptable flag is computed for the devices a query
// returns. It depends on the namespace's limit, so it cannot be a stored column.
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

// DeviceResolver names the field a device is looked up by.
type DeviceResolver uint

// The fields a device can be resolved by. The zero value is not one, so an unset resolver
// cannot silently mean the first.
const (
	DeviceUIDResolver DeviceResolver = iota + 1
	DeviceHostnameResolver
	DeviceMACResolver
	// DevicePublicKeyResolver resolves a device by its public key. The key is the
	// same across namespaces, so callers typically pair it with a status filter
	// and an unbounded scope (e.g. to find where a key was already accepted).
	DevicePublicKeyResolver
)

// DeviceBeat says the device UID was seen alive at At. At is taken when the device is observed
// (its authentication, or its tunnel answering a keep-alive), not when the beat is written.
type DeviceBeat struct {
	UID string
	At  time.Time
}

// DeviceStore persists devices and the counters a namespace's device limit is enforced from.
type DeviceStore interface {
	// DeviceCreate creates a new device. It returns the inserted UID and an error, if any.
	DeviceCreate(ctx context.Context, device *models.Device) (insertedUID string, err error)

	DeviceList(ctx context.Context, sc scope.Scope, acceptable DeviceAcceptable, opts ...QueryOption) ([]models.Device, int64, error)

	// DeviceResolve fetches a device using a specific resolver within the given namespace scope.
	//
	// It returns the resolved device if found and an error, if any.
	DeviceResolve(ctx context.Context, sc scope.Scope, resolver DeviceResolver, value string, opts ...QueryOption) (*models.Device, error)

	// DeviceConflicts reports whether the target contains conflicting attributes with the database. Pass zero values for
	// attributes you do not wish to match on. For example, the following call checks for conflicts based on email only:
	//
	//  ctx := context.Background()
	//  conflicts, has, err := store.DeviceConflicts(ctx, &models.DeviceConflicts{Name: "mydevice"})
	//
	// It returns an array of conflicting attribute fields and an error, if any.
	//
	// Only accepted devices are considered conflicting, within the given namespace scope.
	DeviceConflicts(ctx context.Context, sc scope.Scope, target *models.DeviceConflicts, opts ...QueryOption) (conflicts []string, has bool, err error)

	// DeviceUpdate updates a device. It returns [ErrNoDocuments] if none device is found.
	DeviceUpdate(ctx context.Context, device *models.Device) error
	// DeviceUpdateUnlessRemoved updates a device the way DeviceUpdate does, but only while it is not
	// removed. It returns [ErrNoDocuments] when the device was removed after the caller read it, so
	// writing back a snapshot taken before a removal cannot undo it.
	DeviceUpdateUnlessRemoved(ctx context.Context, device *models.Device) error
	// DeviceHeartbeat sets last_seen to the beat's time for every beat's device that exists and is
	// not removed, and clears its disconnected_at unless the disconnect is newer than the beat, so a
	// beat written late cannot bring a disconnected device back online. Each UID must appear in beats
	// at most once. It returns the other uids, those deleted or removed, in the order given and never
	// nil, or a nil gone and the store error when the update fails.
	DeviceHeartbeat(ctx context.Context, beats []DeviceBeat) (gone []string, err error)

	// DeviceOffline stamps disconnected_at to mark a device offline: the targeted counterpart to
	// DeviceHeartbeat, since disconnected_at is skipupdate. Returns [ErrNoDocuments] if none found.
	DeviceOffline(ctx context.Context, uid string, disconnectedAt time.Time) error

	// DeviceSetCustomField sets or updates a single custom_fields entry on the device atomically.
	DeviceSetCustomField(ctx context.Context, uid, key, value string) error
	// DeviceDeleteCustomField removes a single custom_fields entry from the device atomically.
	// It is idempotent: removing a non-existent key is not an error.
	DeviceDeleteCustomField(ctx context.Context, uid, key string) error

	// DeviceSetOwner ties the device to the member ownerID, or makes it a team device when ownerID
	// is empty. It returns [ErrNoDocuments] when the scope holds no such device. An owner who is not
	// a member of the namespace fails when the enclosing transaction commits.
	DeviceSetOwner(ctx context.Context, sc scope.Scope, uid, ownerID string) error

	// DeviceLockMAC blocks until no other transaction holds the lock on mac in the scope's namespace,
	// then holds it until the transaction in ctx ends. It locks the MAC itself rather than a device
	// row, so a waiter is not released early when the device holding that MAC is deleted. It returns
	// ErrInvalidScope for an unbounded scope and ErrLockOutsideTransaction when ctx carries no
	// transaction.
	DeviceLockMAC(ctx context.Context, sc scope.Scope, mac string) error

	DeviceDelete(ctx context.Context, device *models.Device) error
	// DeviceDeleteMany deletes multiple devices by their UIDs.
	DeviceDeleteMany(ctx context.Context, uids []string) (deletedCount int64, err error)

	// DeviceListExpiredEphemeral lists ephemeral devices that have stayed offline longer than their
	// per-device ephemeral timeout (in minutes), so the cleanup can remove them.
	DeviceListExpiredEphemeral(ctx context.Context) (devices []models.Device, err error)
}
