package services

import (
	"context"
	"crypto/rand"
	"errors"
	"slices"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
)

const webhookLimit = 10

// WebhookService manages the webhooks a namespace's events are sent to.
type WebhookService interface {
	// CreateWebhook subscribes req.URL to req.Events in req.TenantID's namespace, enabled, with a
	// signing secret it generates and returns once, alongside the webhook. It counts and inserts
	// under the namespace's webhook lock, so concurrent creates cannot pass the limit. It returns
	// ErrForbidden for an empty tenant, ErrWebhookInvalidField for a URL that is not http or https
	// or an event list that is empty or names an unknown type, and ErrWebhookLimit when the
	// namespace already has the most webhooks it may. Any other failure returns the underlying
	// error: the encryption's for the secret, or the store's, a ctx cancelled or past its deadline
	// included.
	CreateWebhook(ctx context.Context, req *requests.CreateWebhook) (*responses.CreateWebhook, error)

	// ListWebhooks returns a page of the namespace's webhooks, oldest first, and how many there
	// are. No webhook carries its secret. It returns ErrForbidden for an empty tenant and the
	// store's error otherwise, a ctx cancelled or past its deadline included.
	ListWebhooks(ctx context.Context, req *requests.ListWebhooks) ([]models.Webhook, int64, error)

	// GetWebhook returns the namespace's webhook with req.ID. It returns ErrForbidden for an empty
	// tenant, ErrWebhookNotFound when the namespace has no such webhook and the store's error
	// otherwise, a ctx cancelled or past its deadline included.
	GetWebhook(ctx context.Context, req *requests.WebhookID) (*models.Webhook, error)

	// UpdateWebhook changes the URL, events and enabled flag req sets, leaving the nil ones as they
	// are, and returns the result. It holds the namespace's webhook lock while it reads and writes
	// the webhook, so two updates of it apply one after the other. It returns ErrForbidden for an
	// empty tenant, ErrWebhookNotFound, the ErrWebhookInvalidField cases of CreateWebhook and the
	// store's error otherwise, a ctx cancelled or past its deadline included.
	UpdateWebhook(ctx context.Context, req *requests.UpdateWebhook) (*models.Webhook, error)

	// DeleteWebhook deletes the namespace's webhook with req.ID and its deliveries. It returns
	// ErrForbidden for an empty tenant, ErrWebhookNotFound when the namespace has no such webhook
	// and the store's error otherwise, a ctx cancelled or past its deadline included.
	DeleteWebhook(ctx context.Context, req *requests.WebhookID) error

	// RevealWebhookSecret returns the signing secret of the namespace's webhook with req.ID. It
	// returns ErrForbidden for an empty tenant, ErrWebhookNotFound when the namespace has no such
	// webhook, the decryption's error when the stored secret cannot be decrypted and the store's
	// error otherwise, a ctx cancelled or past its deadline included.
	RevealWebhookSecret(ctx context.Context, req *requests.WebhookID) (string, error)
}

func uniqueWebhookEvents(events []models.WebhookEventType) []models.WebhookEventType {
	return slices.Compact(slices.Sorted(slices.Values(events)))
}

func validateWebhook(url string, events []models.WebhookEventType) error {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return NewErrWebhookInvalidField(map[string]string{"url": "must be an http or https URL"})
	}

	if len(events) == 0 {
		return NewErrWebhookInvalidField(map[string]string{"events": "at least one event type is required"})
	}

	for _, event := range events {
		if !slices.Contains(models.WebhookEventTypes, event) {
			return NewErrWebhookInvalidField(map[string]string{"events": "unknown event type " + string(event)})
		}
	}

	return nil
}

func (s *service) CreateWebhook(ctx context.Context, req *requests.CreateWebhook) (*responses.CreateWebhook, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	events := uniqueWebhookEvents(req.Events)
	if err := validateWebhook(req.URL, events); err != nil {
		return nil, err
	}

	secret := rand.Text()

	encrypted, err := s.encryptAtRest(secret)
	if err != nil {
		return nil, err
	}

	webhook := &models.Webhook{
		TenantID:        req.TenantID,
		URL:             req.URL,
		Events:          events,
		Enabled:         true,
		SecretEncrypted: encrypted,
	}

	if err := s.store.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.store.WebhookLockNamespace(ctx, sc); err != nil {
			return err
		}

		_, count, err := s.store.WebhookList(ctx, sc, s.store.Options().Paginate(&query.Paginator{Page: 1, PerPage: 1}))
		if err != nil {
			return err
		}

		if count >= webhookLimit {
			return NewErrWebhookLimit(webhookLimit)
		}

		_, err = s.store.WebhookCreate(ctx, webhook)

		return err
	}); err != nil {
		return nil, err
	}

	return &responses.CreateWebhook{Webhook: *webhook, Secret: secret}, nil
}

func (s *service) ListWebhooks(ctx context.Context, req *requests.ListWebhooks) ([]models.Webhook, int64, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, 0, err
	}

	return s.store.WebhookList(ctx, sc, s.store.Options().Paginate(&req.Paginator))
}

func (s *service) resolveWebhook(ctx context.Context, tenantID, id string) (*models.Webhook, error) {
	sc, err := BoundTo(tenantID)
	if err != nil {
		return nil, err
	}

	return s.resolveWebhookIn(ctx, sc, id)
}

func (s *service) resolveWebhookIn(ctx context.Context, sc scope.Scope, id string) (*models.Webhook, error) {
	webhook, err := s.store.WebhookResolve(ctx, sc, id)
	if err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return nil, NewErrWebhookNotFound(id, err)
		}

		return nil, err
	}

	return webhook, nil
}

func (s *service) GetWebhook(ctx context.Context, req *requests.WebhookID) (*models.Webhook, error) {
	return s.resolveWebhook(ctx, req.TenantID, req.ID)
}

func (s *service) UpdateWebhook(ctx context.Context, req *requests.UpdateWebhook) (*models.Webhook, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	var webhook *models.Webhook
	if err := s.store.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.store.WebhookLockNamespace(ctx, sc); err != nil {
			return err
		}

		var err error
		webhook, err = s.resolveWebhookIn(ctx, sc, req.ID)
		if err != nil {
			return err
		}

		if req.URL != nil {
			webhook.URL = *req.URL
		}

		if req.Events != nil {
			webhook.Events = uniqueWebhookEvents(req.Events)
		}

		if req.Enabled != nil {
			webhook.Enabled = *req.Enabled
		}

		if err := validateWebhook(webhook.URL, webhook.Events); err != nil {
			return err
		}

		if err := s.store.WebhookUpdate(ctx, webhook); err != nil {
			if errors.Is(err, store.ErrNoDocuments) {
				return NewErrWebhookNotFound(req.ID, err)
			}

			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return webhook, nil
}

func (s *service) DeleteWebhook(ctx context.Context, req *requests.WebhookID) error {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	if err := s.store.WebhookDelete(ctx, sc, req.ID); err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrWebhookNotFound(req.ID, err)
		}

		return err
	}

	return nil
}

func (s *service) RevealWebhookSecret(ctx context.Context, req *requests.WebhookID) (string, error) {
	webhook, err := s.resolveWebhook(ctx, req.TenantID, req.ID)
	if err != nil {
		return "", err
	}

	return s.decryptAtRest(webhook.SecretEncrypted)
}
