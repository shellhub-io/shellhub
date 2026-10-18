package store

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// ProvisioningKeyResolver names the field a provisioning key is looked up by.
type ProvisioningKeyResolver uint

// The fields a provisioning key can be resolved by. The zero value is not one, so an unset
// resolver cannot silently mean the first.
const (
	ProvisioningKeyIDResolver ProvisioningKeyResolver = iota + 1
	ProvisioningKeyNameResolver
)

// ProvisioningKeyStore persists provisioning keys, their usage counters, and the record of which
// devices each enrolled.
type ProvisioningKeyStore interface {
	// ProvisioningKeyCreate creates a provisioning key with the provided data, linked to the tags its
	// TagIDs name, in one transaction, and returns the inserted ID. A key digest or (namespace, name)
	// already in use fails the create with [ErrDuplicate], a ctx cancelled or past its deadline fails
	// it with ctx's error, and any other failure, such as a TagIDs entry no tag holds, fails it with
	// [ErrInternal]. Whatever fails, nothing is written.
	ProvisioningKeyCreate(ctx context.Context, provisioningKey *models.ProvisioningKey) (insertedID string, err error)

	// ProvisioningKeyResolve fetches a provisioning key and its tags using a specific resolver within
	// the given namespace scope. The digest is unique across namespaces, so resolving by ID unbounded
	// finds the namespace that owns the key. A name is unique only within a namespace, and an
	// unbounded resolve by name that matches keys in several namespaces returns [ErrAmbiguous]. It
	// returns [ErrResolverNotFound] for a resolver it does not know, [ErrInvalidScope] for a scope
	// never constructed, [ErrNoDocuments] when no key matches, ctx's error when it is cancelled or
	// its deadline passes and [ErrInternal] for any other failure.
	ProvisioningKeyResolve(ctx context.Context, sc scope.Scope, resolver ProvisioningKeyResolver, value string, opts ...QueryOption) (*models.ProvisioningKey, error)

	// ProvisioningKeyResolveSystem fetches the namespace's system-managed legacy key and its tags. It
	// returns [ErrInvalidScope] for a scope not bounded to a namespace, [ErrNoDocuments] when the
	// namespace has no such key, ctx's error when it is cancelled or its deadline passes and
	// [ErrInternal] for any other failure.
	ProvisioningKeyResolveSystem(ctx context.Context, sc scope.Scope) (*models.ProvisioningKey, error)

	// ProvisioningKeyResolveSystemPairing fetches the namespace's system-managed pairing key and its
	// tags: the source attributed to devices accepted through the tenant-less pairing-code flow. It
	// returns [ErrInvalidScope] for a scope not bounded to a namespace, [ErrNoDocuments] when the
	// namespace has no such key, ctx's error when it is cancelled or its deadline passes and
	// [ErrInternal] for any other failure.
	ProvisioningKeyResolveSystemPairing(ctx context.Context, sc scope.Scope) (*models.ProvisioningKey, error)

	// ProvisioningKeyConflicts reports whether the target contains conflicting attributes with the database. Pass zero
	// values for attributes you do not wish to match on. It returns an array of conflicting attribute fields and
	// an error, if any.
	//
	// Provisioning key attributes can be duplicated at document level when the tenant ID is different.
	ProvisioningKeyConflicts(ctx context.Context, sc scope.Scope, target *models.ProvisioningKeyConflicts) (conflicts []string, has bool, err error)

	// ProvisioningKeyList retrieves a list of provisioning keys within the given namespace scope.
	// Returns the list of provisioning keys, the total count of matched documents, and an error if any.
	ProvisioningKeyList(ctx context.Context, sc scope.Scope, opts ...QueryOption) (provisioningKeys []models.ProvisioningKey, count int64, err error)

	// ProvisioningKeyUpdate updates a provisioning key and, unless its TagIDs are nil, replaces its tag
	// links with the ones TagIDs name, in one transaction; empty TagIDs drop every link. It returns
	// [ErrNoDocuments] when no key matches, [ErrDuplicate] when the (namespace, name) pair is taken,
	// ctx's error when it is cancelled or its deadline passes and [ErrInternal] for any other failure,
	// such as a TagIDs entry no tag holds, with nothing written.
	ProvisioningKeyUpdate(ctx context.Context, provisioningKey *models.ProvisioningKey) (err error)

	// ProvisioningKeyIncrementUsage atomically increments the provisioning key's usage counter and sets its last-used
	// timestamp, only when the key still has usage left (unlimited or used_times < usage_limit). It returns
	// [ErrNoDocuments] when the key is already overused, closing the race between concurrent enrollments.
	ProvisioningKeyIncrementUsage(ctx context.Context, provisioningKey *models.ProvisioningKey) (err error)

	// ProvisioningKeyEventCreate appends one immutable row to a provisioning key's enrollment history. The store
	// stamps the event ID and timestamp. It returns an error, if any.
	ProvisioningKeyEventCreate(ctx context.Context, event *models.ProvisioningKeyEvent) (err error)

	// ProvisioningKeyEventStampDecision freezes the enrollment outcome (accepted/rejected + when) on the
	// device's newest history event, so the audit survives the device being removed. Best-effort: a
	// device with no event stamps nothing.
	ProvisioningKeyEventStampDecision(ctx context.Context, sc scope.Scope, deviceUID string, status models.DeviceStatus, at time.Time) (err error)

	// ProvisioningKeyEventList retrieves the enrollment history of the provisioning key identified by its digest
	// within the given namespace scope, newest first. It returns the events, the total count, and an
	// error, if any.
	ProvisioningKeyEventList(ctx context.Context, sc scope.Scope, keyDigest string, opts ...QueryOption) (events []models.ProvisioningKeyEvent, count int64, err error)

	// EnrollmentCallbackRedeem atomically claims a deferred-decision callback token by its JWT id,
	// making it single-use. It returns true when the token was claimed for the first time and false
	// when it had already been redeemed (a replay), plus an error, if any.
	EnrollmentCallbackRedeem(ctx context.Context, jti string, at time.Time) (redeemed bool, err error)

	// EnrollmentCallbackCleanup deletes callback redemption records older than the given cutoff (past
	// the maximum token TTL, so they can no longer gate a replay). It returns the number of rows
	// deleted and an error, if any.
	EnrollmentCallbackCleanup(ctx context.Context, before time.Time) (deleted int64, err error)
}
