package store

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// DeletedDevice is a device [DeviceStore.DeviceDeleteMany] deleted, with the namespace and status
// it had, which are what a caller adjusts the namespace's device counts by.
type DeletedDevice struct {
	UID      string
	TenantID string
	Status   models.DeviceStatus
}

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

// DeviceBeat says the device UID of the namespace TenantID was seen alive at At. At is taken when
// the device is observed (its authentication, or its tunnel answering a keep-alive), not when the
// beat is written.
type DeviceBeat struct {
	UID      string
	TenantID string
	At       time.Time
}

// DeviceStore persists devices and the counters a namespace's device limit is enforced from.
type DeviceStore interface {
	// DeviceCreate creates device in the namespace sc is bounded to and returns its UID. It returns
	// ErrInvalidScope, writing nothing, when sc is not bounded or device's TenantID is not sc's
	// tenant, and the store's error when the insert fails.
	DeviceCreate(ctx context.Context, sc scope.Scope, device *models.Device) (insertedUID string, err error)

	DeviceList(ctx context.Context, sc scope.Scope, acceptable DeviceAcceptable, opts ...QueryOption) ([]models.Device, int64, error)

	// DeviceLocations groups the scope's accepted devices by position, rounded to
	// [models.DeviceLocationPrecision] decimal places, and counts the devices and the online
	// devices at each. A device with no position, which the store keeps as 0,0, counts nowhere. The
	// order of the locations is unspecified. It reads under ctx, within the transaction ctx carries,
	// if any. It returns ErrInvalidScope for a scope of no known kind, ErrNoDocuments for a bounded
	// scope whose tenant is not a UUID, an option's own error, ctx's error unwrapped when ctx ends,
	// and an error wrapping [ErrInternal] for any other database failure.
	DeviceLocations(ctx context.Context, sc scope.Scope, opts ...QueryOption) ([]models.DeviceLocation, error)

	// DeviceListAfter reads up to size devices ordered by creation time, then id, ascending,
	// starting after the after cursor, or from the first device when after is nil. A size below
	// one reads one. It applies sc, acceptable and opts as DeviceList does, takes no Sort or
	// Paginate option, counts nothing, and reads within the transaction ctx carries, if any. next
	// is the cursor of the last device returned when more devices follow, and nil on the last
	// page, so a page holding devices may be the last one but no empty page follows a non-nil
	// next. It returns ErrInvalidScope for a scope of no known kind, ErrNoDocuments for a bounded
	// scope whose tenant is not a UUID, an option's own error, an error wrapping ctx's error when
	// ctx ends, and an error wrapping [ErrInternal] for any other database failure.
	DeviceListAfter(ctx context.Context, sc scope.Scope, acceptable DeviceAcceptable, after *query.Cursor, size int64, opts ...QueryOption) (devices []models.Device, next *query.Cursor, err error)

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

	// DeviceUpdate updates a device within the namespace sc is bounded to. It returns
	// ErrInvalidScope, writing nothing, when sc is not bounded or device's TenantID is not sc's
	// tenant, [ErrNoDocuments] when sc holds no such device, and the store's error when the update
	// fails or, after the row is written, when the follow-up write of KeyVerifiedAt fails. It writes KeyVerifiedAt only while the stored one is unset and the device
	// is not removed, so a snapshot read before a verification cannot undo it and a removal racing a
	// verification wins, and clears it when the device is written as removed, so a removed device
	// accepts unsigned authentication again unless its namespace requires verification.
	DeviceUpdate(ctx context.Context, sc scope.Scope, device *models.Device) error
	// DeviceUpdateUnlessRemoved updates a device the way DeviceUpdate does, but only while it is not
	// removed. It returns [ErrNoDocuments] when the device was removed after the caller read it, so
	// writing back a snapshot taken before a removal cannot undo it, and ErrInvalidScope as
	// DeviceUpdate does.
	DeviceUpdateUnlessRemoved(ctx context.Context, sc scope.Scope, device *models.Device) error
	// DeviceAssertNotRemoved writes nothing and returns [ErrNoDocuments] when the device is missing
	// from the namespace sc is bounded to or removed, the check DeviceUpdateUnlessRemoved makes, for a
	// caller whose snapshot has nothing to write back. It reads only the device's UID and TenantID,
	// and returns ErrInvalidScope as DeviceUpdate does, ctx's error when ctx ends and an error
	// wrapping [ErrInternal] for any other database failure.
	DeviceAssertNotRemoved(ctx context.Context, sc scope.Scope, device *models.Device) error
	// DeviceHeartbeat sets last_seen to the beat's time for every beat's device that exists in the
	// beat's namespace, and in sc's when sc is bounded, and is not removed, and clears its
	// disconnected_at unless the disconnect is newer than the beat, so a beat written late cannot
	// bring a disconnected device back online. A device with no disconnected_at
	// whose last_seen is less than skipWithin before its beat, or after it, is left unwritten, so a
	// skipWithin of zero or less never moves such a device's last_seen backwards. A device carrying a
	// disconnect is written whatever its last_seen, so a late beat can still move its last_seen back.
	// Each UID must appear in beats at most once. It returns the other uids, those deleted, removed
	// or not in the beat's namespace or a bounded sc's, in the order given and never nil, whether or
	// not their beat was written, or a nil gone and ErrInvalidScope for a scope of no known kind or
	// the store error when the update fails.
	DeviceHeartbeat(ctx context.Context, sc scope.Scope, beats []DeviceBeat, skipWithin time.Duration) (gone []string, err error)

	// DeviceOffline stamps disconnected_at to mark a device offline: the targeted counterpart to
	// DeviceHeartbeat, since disconnected_at is skipupdate. It returns ErrInvalidScope, writing
	// nothing, when sc is not bounded, [ErrNoDocuments] when sc holds no such device, and the
	// store's error when the update fails.
	DeviceOffline(ctx context.Context, sc scope.Scope, uid string, disconnectedAt time.Time) error

	// DeviceSweepStalePresence stamps disconnected_at on up to limit devices of sc whose tunnel close
	// was never written: devices that are not removed, carry no disconnected_at and were last seen
	// before staleBefore. Each gets last_seen plus [models.DeviceOnlineWindow], the instant it stopped
	// reading online, so the stamp is in the past and any later beat clears it as DeviceHeartbeat
	// does for a real close. A device never seen is left alone. It returns how many devices it
	// stamped, ErrInvalidScope for a scope of no known kind, [ErrNoDocuments] for a bounded scope
	// whose tenant is not a UUID, ctx's error when ctx ends, and an error wrapping [ErrInternal] for
	// any other database failure.
	DeviceSweepStalePresence(ctx context.Context, sc scope.Scope, staleBefore time.Time, limit int64) (stamped int64, err error)

	// DeviceSetCustomField sets or updates a single custom_fields entry on the device atomically. It
	// returns ErrInvalidScope, writing nothing, when sc is not bounded, [ErrNoDocuments] when sc
	// holds no such device, and the store's error when the update fails.
	DeviceSetCustomField(ctx context.Context, sc scope.Scope, uid, key, value string) error
	// DeviceDeleteCustomField removes a single custom_fields entry from the device atomically.
	// It is idempotent: removing a non-existent key is not an error. It returns the errors
	// DeviceSetCustomField does.
	DeviceDeleteCustomField(ctx context.Context, sc scope.Scope, uid, key string) error

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

	// DeviceDelete deletes device, with its sessions, from the namespace sc is bounded to. It returns
	// ErrInvalidScope, deleting nothing, when sc is not bounded or device's TenantID is not sc's
	// tenant, [ErrNoDocuments] when sc holds no such device, and the store's error otherwise.
	DeviceDelete(ctx context.Context, sc scope.Scope, device *models.Device) error
	// DeviceDeleteMany deletes the devices of sc that uids names, with their sessions, and returns the
	// ones it deleted. A uid that names no device of sc, including one another call deleted first, is
	// left out, so the result can be shorter than uids, and an empty uids deletes nothing. It returns
	// ErrInvalidScope for a scope of no known kind, ctx's error when ctx ends, and an error wrapping
	// [ErrInternal] for any other database failure.
	DeviceDeleteMany(ctx context.Context, sc scope.Scope, uids []string) (deleted []DeletedDevice, err error)

	// DeviceListExpiredEphemeral lists up to 1000 ephemeral devices of sc the cleanup can remove: those
	// that do not read online and whose last tunnel beat and last status change, whichever is later,
	// are older than their per-device ephemeral timeout (in minutes). A device that is not accepted
	// is listed only once the server has seen its tunnel close since its last status change. It
	// returns ErrInvalidScope for a scope of no known kind, [ErrNoDocuments] for a bounded scope
	// whose tenant is not a UUID, ctx's error when ctx ends, and an error wrapping [ErrInternal] for
	// any other database failure.
	DeviceListExpiredEphemeral(ctx context.Context, sc scope.Scope) (devices []models.Device, err error)
}
