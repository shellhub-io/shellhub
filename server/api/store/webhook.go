package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// WebhookStore persists namespaces' webhooks.
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
}
