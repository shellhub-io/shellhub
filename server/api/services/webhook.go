package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/shellhub-io/shellhub/server/api/pkg/webhook"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

const (
	webhookLimit = 10

	webhookMaxAttempts        = 12
	webhookRetryBase          = time.Minute
	webhookRetryMax           = time.Hour
	webhookRetryMaxDoublings  = 6
	webhookDeliveryBatch      = 10
	webhookDeliveryLease      = time.Minute
	webhookRequestTimeout     = 10 * time.Second
	webhookResponseDrainLimit = 64 << 10

	webhookDeliveryRetention = 7 * 24 * time.Hour
)

// CronWebhookDeliverySweep and CronWebhookDeliveryCleanup are when the webhook delivery jobs run:
// the sweep every minute, the cleanup at night like the other cleanups.
const (
	CronWebhookDeliverySweep   = worker.CronSpec("* * * * *")
	CronWebhookDeliveryCleanup = worker.CronSpec("30 4 * * *")
)

// TaskWebhookDelivery is the task a committed event submits so its deliveries go out at once,
// rather than at the next CronWebhookDeliverySweep.
const TaskWebhookDelivery = worker.TaskPattern("webhook:deliver")

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

	// ListWebhookDeliveries returns a page of the deliveries to the namespace's webhook with
	// req.ID, newest first, and how many there are. It returns ErrForbidden for an empty tenant,
	// ErrWebhookNotFound when the namespace has no such webhook and the store's error otherwise,
	// a ctx cancelled or past its deadline included.
	ListWebhookDeliveries(ctx context.Context, req *requests.ListWebhookDeliveries) ([]models.WebhookDelivery, int64, error)

	// DeliverWebhooks returns the task that sends every delivery due now, in batches, until none
	// is left. A delivery answered 2xx is delivered; any other answer, a timeout or a redirect
	// schedules a retry with backoff, and the last allowed attempt failing abandons it. The task
	// returns only a failure to claim deliveries; a failed send is recorded on the delivery.
	DeliverWebhooks() worker.TaskHandler

	// WebhookDeliverySweep returns the cron job that does what DeliverWebhooks does, for the
	// deliveries no submitted task reached and the retries that came due.
	WebhookDeliverySweep() worker.CronHandler

	// WebhookDeliveryCleanup returns the cron job that deletes the deliveries delivered or
	// abandoned more than seven days ago, and the events left with no delivery.
	WebhookDeliveryCleanup() worker.CronHandler
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

func (s *service) ListWebhookDeliveries(ctx context.Context, req *requests.ListWebhookDeliveries) ([]models.WebhookDelivery, int64, error) {
	webhook, err := s.resolveWebhook(ctx, req.TenantID, req.ID)
	if err != nil {
		return nil, 0, err
	}

	return s.store.WebhookDeliveryList(ctx, scope.MustBounded(webhook.TenantID), webhook.ID, s.store.Options().Paginate(&req.Paginator))
}

func deviceStatusEvent(status models.DeviceStatus) (models.WebhookEventType, bool) {
	switch status {
	case models.DeviceStatusPending:
		return models.WebhookEventDevicePending, true
	case models.DeviceStatusAccepted:
		return models.WebhookEventDeviceAccepted, true
	case models.DeviceStatusRejected:
		return models.WebhookEventDeviceRejected, true
	case models.DeviceStatusRemoved:
		return models.WebhookEventDeviceRemoved, true
	default:
		return "", false
	}
}

func (s *service) removedDevice(device models.Device) *models.Device {
	now := s.clock.Now()
	device.Status = models.DeviceStatusRemoved
	device.RemovedAt = &now

	return &device
}

func (s *service) recordDeviceEvent(ctx context.Context, eventType models.WebhookEventType, device *models.Device) error {
	return recordDeviceEvent(ctx, s.store, s.clock, eventType, device)
}

type webhookDevice struct {
	UID             string                 `json:"uid"`
	Name            string                 `json:"name"`
	Status          models.DeviceStatus    `json:"status"`
	TenantID        string                 `json:"tenant_id"`
	Namespace       string                 `json:"namespace"`
	Identity        *models.DeviceIdentity `json:"identity"`
	Info            *models.DeviceInfo     `json:"info"`
	Tags            []models.Tag           `json:"tags"`
	CreatedAt       time.Time              `json:"created_at"`
	StatusUpdatedAt time.Time              `json:"status_updated_at"`
	RemovedAt       *time.Time             `json:"removed_at"`
}

func newWebhookDevice(device *models.Device) webhookDevice {
	tags := device.Tags
	if tags == nil {
		tags = []models.Tag{}
	}

	return webhookDevice{
		UID:             device.UID,
		Name:            device.Name,
		Status:          device.Status,
		TenantID:        device.TenantID,
		Namespace:       device.Namespace,
		Identity:        device.Identity,
		Info:            device.Info,
		Tags:            tags,
		CreatedAt:       device.CreatedAt,
		StatusUpdatedAt: device.StatusUpdatedAt,
		RemovedAt:       device.RemovedAt,
	}
}

func recordDeviceEvent(ctx context.Context, st store.WebhookStore, clk clock.Clock, eventType models.WebhookEventType, device *models.Device) error {
	data, err := json.Marshal(map[string]webhookDevice{"device": newWebhookDevice(device)})
	if err != nil {
		return err
	}

	_, err = st.WebhookEventRecord(ctx, &models.WebhookEvent{
		ID:        uuid.Generate(),
		TenantID:  device.TenantID,
		Type:      eventType,
		Data:      data,
		CreatedAt: clk.Now(),
	})

	return err
}

// RecordDepartureRemovals records a device.removed event for each device departed removed, as the
// store holds it after the departure and stamped with clk's time. Call it with the ctx of the
// transaction that departed the member, so the events commit or roll back with the departure; the
// next delivery sweep sends them. A departure that removed no device records nothing. It returns
// ErrForbidden when departed names no tenant, and the store's error when a removed device cannot be
// resolved or an event cannot be recorded, a ctx cancelled or past its deadline included.
func RecordDepartureRemovals(ctx context.Context, st store.Store, clk clock.Clock, departed *models.MemberDeparted) error {
	if len(departed.RemovedDevices) == 0 {
		return nil
	}

	sc, err := BoundTo(departed.TenantID)
	if err != nil {
		return err
	}

	for _, uid := range departed.RemovedDevices {
		device, err := st.DeviceResolve(ctx, sc, store.DeviceUIDResolver, uid)
		if err != nil {
			return err
		}

		if err := recordDeviceEvent(ctx, st, clk, models.WebhookEventDeviceRemoved, device); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) DeliverWebhooks() worker.TaskHandler {
	return func(ctx context.Context, _ []byte) error {
		return s.deliverWebhooks(ctx)
	}
}

func (s *service) deliverWebhooks(ctx context.Context) error {
	client := webhook.NewClient(s.webhookCIDRs)
	defer client.CloseIdleConnections()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	for {
		now := s.clock.Now()

		dispatches, err := s.store.WebhookDeliveryClaim(ctx, now, now.Add(webhookDeliveryLease), webhookDeliveryBatch)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		for i := range dispatches {
			wg.Go(func() { s.dispatchWebhook(ctx, client, &dispatches[i]) })
		}
		wg.Wait()

		if len(dispatches) < webhookDeliveryBatch {
			return nil
		}
	}
}

func webhookRetryDelay(attempts int) time.Duration {
	delay := min(webhookRetryBase<<min(attempts-1, webhookRetryMaxDoublings), webhookRetryMax)

	jitter, err := rand.Int(rand.Reader, big.NewInt(int64(delay/10)))
	if err != nil {
		return delay
	}

	return delay + time.Duration(jitter.Int64())
}

func (s *service) dispatchWebhook(ctx context.Context, client *http.Client, dispatch *store.WebhookDispatch) {
	delivery := dispatch.Delivery

	statusCode, sendErr := s.sendWebhook(ctx, client, dispatch)
	now := s.clock.Now()

	delivery.Attempts++
	delivery.LastStatusCode = statusCode

	switch {
	case sendErr == nil:
		delivery.Status = models.WebhookDeliveryDelivered
		delivery.DeliveredAt = &now
		delivery.LastError = ""
	case delivery.Attempts >= webhookMaxAttempts:
		delivery.Status = models.WebhookDeliveryFailed
		delivery.LastError = webhook.Describe(sendErr)
	default:
		delivery.NextAttemptAt = now.Add(webhookRetryDelay(delivery.Attempts))
		delivery.LastError = webhook.Describe(sendErr)
	}

	if sendErr != nil {
		log.WithError(sendErr).WithFields(log.Fields{"delivery_id": delivery.ID, "webhook_id": delivery.WebhookID, "attempts": delivery.Attempts}).
			Warn("webhook delivery attempt failed")
	}

	if err := s.store.WebhookDeliveryFinish(ctx, &delivery); err != nil {
		log.WithError(err).WithFields(log.Fields{"delivery_id": delivery.ID, "webhook_id": delivery.WebhookID}).
			Error("failed to record the webhook delivery attempt; it is retried when its lease expires")
	}
}

func (s *service) sendWebhook(ctx context.Context, client *http.Client, dispatch *store.WebhookDispatch) (int, error) {
	secret, err := s.decryptAtRest(dispatch.SecretEncrypted)
	if err != nil {
		return 0, err
	}

	body, err := json.Marshal(dispatch.Event)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(ctx, webhookRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, dispatch.URL, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	timestamp := s.clock.Now().Unix()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ShellHub-Event-ID", dispatch.Event.ID)
	req.Header.Set("X-ShellHub-Event-Type", string(dispatch.Event.Type))
	req.Header.Set("X-ShellHub-Timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("X-ShellHub-Signature", webhook.SignV1(secret, timestamp, body))

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close() //nolint:errcheck // the body is only drained, so a close error changes nothing

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, webhookResponseDrainLimit))

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return resp.StatusCode, webhook.StatusError(resp.StatusCode)
	}

	return resp.StatusCode, nil
}

func (s *service) nudgeWebhookDelivery(ctx context.Context) {
	if s.webhookDispatcher == nil {
		return
	}

	if err := s.webhookDispatcher.Submit(ctx, TaskWebhookDelivery, nil); err != nil {
		log.WithError(err).Warn("failed to submit the webhook delivery task; the next sweep sends the deliveries")
	}
}

func (s *service) WebhookDeliverySweep() worker.CronHandler {
	return s.deliverWebhooks
}

func (s *service) WebhookDeliveryCleanup() worker.CronHandler {
	return func(ctx context.Context) error {
		deleted, err := s.store.WebhookDeliveryPrune(ctx, s.clock.Now().Add(-webhookDeliveryRetention))
		if err != nil {
			return err
		}

		log.WithField("deleted", deleted).Info("webhook delivery cleanup completed")

		return nil
	}
}
