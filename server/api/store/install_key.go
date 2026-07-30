package store

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type InstallKeyResolver uint

const (
	InstallKeyIDResolver InstallKeyResolver = iota + 1
	InstallKeyNameResolver
)

type InstallKeyStore interface {
	// InstallKeyCreate creates a install key with the provided data. Returns the inserted ID and an error if any.
	InstallKeyCreate(ctx context.Context, installKey *models.InstallKey) (insertedID string, err error)

	// InstallKeyResolve fetches a install key using a specific resolver within a given tenant ID.
	//
	// It returns the resolved install key if found and an error, if any.
	InstallKeyResolve(ctx context.Context, resolver InstallKeyResolver, value string, opts ...QueryOption) (*models.InstallKey, error)

	// InstallKeyResolveSystem fetches the namespace's system-managed legacy key. It returns the key if
	// found and an error, if any.
	InstallKeyResolveSystem(ctx context.Context, tenantID string) (*models.InstallKey, error)

	// InstallKeyResolveSystemPairing fetches the namespace's system-managed pairing key: the source
	// attributed to devices accepted through the tenant-less pairing-code flow. It returns the key if
	// found and an error, if any.
	InstallKeyResolveSystemPairing(ctx context.Context, tenantID string) (*models.InstallKey, error)

	// InstallKeyConflicts reports whether the target contains conflicting attributes with the database. Pass zero
	// values for attributes you do not wish to match on. It returns an array of conflicting attribute fields and
	// an error, if any.
	//
	// Install key attributes can be duplicated at document level when the tenant ID is different.
	InstallKeyConflicts(ctx context.Context, tenantID string, target *models.InstallKeyConflicts) (conflicts []string, has bool, err error)

	// InstallKeyList retrieves a list of install keys.
	// Returns the list of install keys, the total count of matched documents, and an error if any.
	InstallKeyList(ctx context.Context, opts ...QueryOption) (installKeys []models.InstallKey, count int, err error)

	// InstallKeyUpdate updates a install key. It returns an error if any.
	InstallKeyUpdate(ctx context.Context, installKey *models.InstallKey) (err error)

	// InstallKeyIncrementUsage atomically increments the install key's usage counter and sets its last-used
	// timestamp, only when the key still has usage left (unlimited or used_times < usage_limit). It returns
	// [ErrNoDocuments] when the key is already overused, closing the race between concurrent enrollments.
	InstallKeyIncrementUsage(ctx context.Context, installKey *models.InstallKey) (err error)

	// InstallKeyDecrementUsage returns a use previously reserved by [InstallKeyIncrementUsage] when the
	// enrollment it was reserved for did not go through (e.g. the accept failed), guarding at zero so a
	// release never drives the counter negative. It returns [ErrNoDocuments] when there was nothing to
	// release (counter already at zero).
	InstallKeyDecrementUsage(ctx context.Context, installKey *models.InstallKey) (err error)

	// InstallKeyEventCreate appends one immutable row to an install key's enrollment history. The store
	// stamps the event ID and timestamp. It returns an error, if any.
	InstallKeyEventCreate(ctx context.Context, event *models.InstallKeyEvent) (err error)

	// InstallKeyEventStampDecision freezes the enrollment outcome (accepted/rejected + when) on the
	// device's newest history event, so the audit survives the device being removed. Best-effort: a
	// device with no event stamps nothing.
	InstallKeyEventStampDecision(ctx context.Context, tenantID, deviceUID string, status models.DeviceStatus, at time.Time) (err error)

	// InstallKeyEventList retrieves the enrollment history of the install key identified by its digest
	// within a tenant ID, newest first. It returns the events, the total count, and an error, if any.
	InstallKeyEventList(ctx context.Context, tenantID, keyDigest string, opts ...QueryOption) (events []models.InstallKeyEvent, count int, err error)

	// EnrollmentCallbackRedeem atomically claims a deferred-decision callback token by its JWT id,
	// making it single-use. It returns true when the token was claimed for the first time and false
	// when it had already been redeemed (a replay), plus an error, if any.
	EnrollmentCallbackRedeem(ctx context.Context, jti string, at time.Time) (redeemed bool, err error)

	// EnrollmentCallbackCleanup deletes callback redemption records older than the given cutoff (past
	// the maximum token TTL, so they can no longer gate a replay). It returns the number of rows
	// deleted and an error, if any.
	EnrollmentCallbackCleanup(ctx context.Context, before time.Time) (deleted int64, err error)
}
