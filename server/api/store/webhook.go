package store

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// WebhookDispatch is a claimed delivery with what sending it takes: the event, where it goes and
// the secret that signs it.
type WebhookDispatch struct {
	Delivery        models.WebhookDelivery
	Event           models.WebhookEvent
	URL             string
	SecretEncrypted string
}

// WebhookStore persists namespaces' webhooks, the events recorded for them and the deliveries of
// those events.
type WebhookStore interface {
	// WebhookCreate stores webhook, stamping its ID and timestamps, and returns the ID. It returns
	// ctx's error when it is cancelled or its deadline passes and [ErrInternal] for any other
	// failure, such as a namespace that does not exist.
	WebhookCreate(ctx context.Context, webhook *models.Webhook) (string, error)

	// WebhookLockNamespace blocks until no other transaction holds the scope's namespace webhook
	// lock, then holds it until the transaction in ctx ends. It returns [ErrInvalidScope] for an
	// unbounded scope, [ErrLockOutsideTransaction] when ctx carries no transaction, ctx's error
	// when it is cancelled or its deadline passes and [ErrInternal] for any other failure.
	WebhookLockNamespace(ctx context.Context, sc scope.Scope) error

	// WebhookList returns a page of the scope's webhooks, oldest first, and how many there are in
	// all. It returns [ErrInvalidScope] for a scope never constructed, ctx's error when it is
	// cancelled or its deadline passes and [ErrInternal] for any other failure.
	WebhookList(ctx context.Context, sc scope.Scope, opts ...QueryOption) ([]models.Webhook, int64, error)

	// WebhookResolve fetches the scope's webhook with id. It returns [ErrNoDocuments] when the
	// scope holds no such webhook, an id that is not a UUID included, [ErrInvalidScope] for a scope
	// never constructed, ctx's error when it is cancelled or its deadline passes and [ErrInternal]
	// for any other failure.
	WebhookResolve(ctx context.Context, sc scope.Scope, id string) (*models.Webhook, error)

	// WebhookUpdate writes webhook's URL, events and enabled flag over the stored ones and stamps
	// its update time. It returns [ErrNoDocuments] when no webhook of webhook's namespace has its
	// ID, ctx's error when it is cancelled or its deadline passes and [ErrInternal] for any other
	// failure.
	WebhookUpdate(ctx context.Context, webhook *models.Webhook) error

	// WebhookDelete deletes the scope's webhook with id, and its deliveries with it. It returns
	// [ErrNoDocuments] when the scope holds no such webhook, [ErrInvalidScope] for a scope not
	// bounded to a namespace, ctx's error when it is cancelled or its deadline passes and
	// [ErrInternal] for any other failure.
	WebhookDelete(ctx context.Context, sc scope.Scope, id string) error

	// WebhookEventRecord stores event with one pending delivery, due at the event's creation, for
	// each enabled webhook of the event's namespace subscribed to its type, and returns how many
	// deliveries it stored. With no such webhook it stores nothing. Call it with the transaction
	// of the change the event describes, so the event exists exactly when the change commits. It
	// returns ctx's error when it is cancelled or its deadline passes and [ErrInternal] for any
	// other failure.
	WebhookEventRecord(ctx context.Context, event *models.WebhookEvent) (int64, error)

	// WebhookDeliveryClaim leases up to limit pending deliveries due at now whose lease, if any,
	// expired by now, oldest due first, until leaseUntil, and returns them for sending. A delivery
	// another claim leased, or is leasing at the same moment, is skipped, so two claims never
	// return the same delivery while its lease holds. A delivery to a disabled webhook is not due,
	// so it waits until the webhook is enabled again. It reads every namespace. It returns ctx's
	// error when it is cancelled or its deadline passes and [ErrInternal] for any other failure.
	WebhookDeliveryClaim(ctx context.Context, now, leaseUntil time.Time, limit int) ([]WebhookDispatch, error)

	// WebhookDeliveryFinish writes delivery's status, attempts, next attempt, last status code,
	// last error and delivery time, releases its lease and stamps its update time. It returns
	// [ErrNoDocuments] when no delivery has its ID, ctx's error when it is cancelled or its
	// deadline passes and [ErrInternal] for any other failure.
	WebhookDeliveryFinish(ctx context.Context, delivery *models.WebhookDelivery) error

	// WebhookDeliveryList returns a page of the scope's deliveries to the webhook with webhookID,
	// newest first, and how many there are in all. A webhookID that is not a UUID names no webhook,
	// so it returns an empty page and no error. It returns [ErrInvalidScope] for a scope never
	// constructed, ctx's error when it is cancelled or its deadline passes and [ErrInternal] for
	// any other failure.
	WebhookDeliveryList(ctx context.Context, sc scope.Scope, webhookID string, opts ...QueryOption) ([]models.WebhookDelivery, int64, error)

	// WebhookDeliveryPrune deletes the delivered and failed deliveries last updated before before,
	// then the events created before before that no delivery points at any more, and returns how
	// many deliveries it deleted. It reads every namespace. It returns ctx's error when it is
	// cancelled or its deadline passes and [ErrInternal] for any other failure.
	WebhookDeliveryPrune(ctx context.Context, before time.Time) (int64, error)
}
