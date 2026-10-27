package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type receivedWebhook struct {
	header http.Header
	body   []byte
}

type webhookReceiver struct {
	*httptest.Server

	mu       sync.Mutex
	received []receivedWebhook
	status   int
}

func newWebhookReceiver(t *testing.T) *webhookReceiver {
	t.Helper()

	r := &webhookReceiver{status: http.StatusOK}
	r.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)

		r.mu.Lock()
		r.received = append(r.received, receivedWebhook{header: req.Header.Clone(), body: body})
		status := r.status
		r.mu.Unlock()

		w.WriteHeader(status)
	}))
	t.Cleanup(r.Close)

	return r
}

func (r *webhookReceiver) answer(status int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.status = status
}

func (r *webhookReceiver) requests() []receivedWebhook {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]receivedWebhook(nil), r.received...)
}

type deliveredEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	TenantID  string    `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
	Data      struct {
		Device models.Device `json:"device"`
	} `json:"data"`
}

func (e *enrollmentE2E) deliver(t *testing.T) {
	t.Helper()

	require.NoError(t, e.svc.DeliverWebhooks()(context.Background(), nil))
}

func (e *enrollmentE2E) setStatus(t *testing.T, uid string, status models.DeviceStatus) error {
	t.Helper()

	return e.svc.UpdateDeviceStatus(context.Background(), &requests.DeviceUpdateStatus{TenantID: e.tenantID, UID: uid, Status: string(status)})
}

func (e *enrollmentE2E) deliveries(t *testing.T, webhookID string) []models.WebhookDelivery {
	t.Helper()

	deliveries, _, err := e.svc.ListWebhookDeliveries(context.Background(), &requests.ListWebhookDeliveries{
		TenantID:  e.tenantID,
		ID:        webhookID,
		Paginator: query.Paginator{Page: 1, PerPage: 100},
	})
	require.NoError(t, err)

	return deliveries
}

func TestWebhookDeliveryE2E_AcceptSendsASignedEvent(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	receiver := newWebhookReceiver(t)
	_, secret := e.createWebhook(t, receiver.URL, models.WebhookEventDeviceAccepted)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusAccepted))

	e.deliver(t)

	got := receiver.requests()
	require.Len(t, got, 1)

	var event deliveredEvent
	require.NoError(t, json.Unmarshal(got[0].body, &event))
	require.Equal(t, "device.accepted", event.Type)
	require.Equal(t, e.tenantID, event.TenantID)
	require.Equal(t, uid, event.Data.Device.UID)
	require.Equal(t, models.DeviceStatusAccepted, event.Data.Device.Status)

	require.Equal(t, "application/json", got[0].header.Get("Content-Type"))
	require.Equal(t, event.ID, got[0].header.Get("X-ShellHub-Event-ID"))
	require.Equal(t, "device.accepted", got[0].header.Get("X-ShellHub-Event-Type"))

	timestamp := strconv.FormatInt(now.Unix(), 10)
	require.Equal(t, timestamp, got[0].header.Get("X-ShellHub-Timestamp"))

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(got[0].body)
	require.Equal(t, "v1="+hex.EncodeToString(mac.Sum(nil)), got[0].header.Get("X-ShellHub-Signature"))

	e.deliver(t)
	require.Len(t, receiver.requests(), 1, "a delivered event is not sent again")
}

func TestWebhookDeliveryE2E_AnEventCarriesOnlyTheDevicesPublicFields(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	receiver := newWebhookReceiver(t)
	e.createWebhook(t, receiver.URL, models.WebhookEventDeviceAccepted)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusAccepted))

	e.deliver(t)

	got := receiver.requests()
	require.Len(t, got, 1)

	var event struct {
		Data struct {
			Device map[string]json.RawMessage `json:"device"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(got[0].body, &event))
	require.ElementsMatch(t, []string{
		"uid", "name", "status", "tenant_id", "namespace", "identity", "info", "tags",
		"created_at", "status_updated_at", "removed_at",
	}, slices.Collect(maps.Keys(event.Data.Device)))
	require.JSONEq(t, "[]", string(event.Data.Device["tags"]))
}

func TestWebhookDeliveryE2E_AFailedAttemptSaysWhyWithoutTheNetworksDetails(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)

	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	untrusted := httptest.NewTLSServer(http.NotFoundHandler())
	t.Cleanup(untrusted.Close)

	cases := []struct {
		name string
		url  string
		want string
	}{
		{"an unknown host", "http://receiver.invalid/", "could not resolve the host"},
		{"a refused connection", closed.URL, "the connection was refused"},
		{"an address outside the allowlist", "http://10.0.0.1/", "the address is not allowed"},
		{"an untrusted certificate", untrusted.URL, "the certificate could not be verified"},
	}

	webhooks := make([]string, len(cases))
	for i, tc := range cases {
		webhooks[i], _ = e.createWebhook(t, tc.url, models.WebhookEventDeviceAccepted)
	}

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusAccepted))

	e.deliver(t)

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			delivery := e.deliveries(t, webhooks[i])[0]
			require.Equal(t, models.WebhookDeliveryPending, delivery.Status)
			require.Equal(t, 1, delivery.Attempts)
			require.Equal(t, tc.want, delivery.LastError)
		})
	}
}

func TestWebhookDeliveryE2E_SendsOnlyTheSubscribedTypesToEnabledWebhooks(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	ctx := context.Background()

	acceptedOnly := newWebhookReceiver(t)
	e.createWebhook(t, acceptedOnly.URL, models.WebhookEventDeviceAccepted)

	rejectedOnly := newWebhookReceiver(t)
	e.createWebhook(t, rejectedOnly.URL, models.WebhookEventDeviceRejected)

	paused := newWebhookReceiver(t)
	pausedID, _ := e.createWebhook(t, paused.URL, models.WebhookEventDeviceAccepted, models.WebhookEventDeviceRejected)
	disabled := false
	_, err := e.svc.UpdateWebhook(ctx, &requests.UpdateWebhook{TenantID: e.tenantID, ID: pausedID, Enabled: &disabled})
	require.NoError(t, err)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusRejected))
	e.deliver(t)

	require.Empty(t, acceptedOnly.requests())
	require.Len(t, rejectedOnly.requests(), 1)
	require.Empty(t, paused.requests(), "a paused webhook receives nothing")
}

func TestWebhookDeliveryE2E_PausingHoldsQueuedDeliveriesUntilResumed(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	ctx := context.Background()
	receiver := newWebhookReceiver(t)
	id, _ := e.createWebhook(t, receiver.URL, models.WebhookEventDeviceAccepted)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusAccepted))

	paused, resumed := false, true
	_, err := e.svc.UpdateWebhook(ctx, &requests.UpdateWebhook{TenantID: e.tenantID, ID: id, Enabled: &paused})
	require.NoError(t, err)
	e.deliver(t)
	require.Empty(t, receiver.requests(), "a paused webhook is sent nothing")

	_, err = e.svc.UpdateWebhook(ctx, &requests.UpdateWebhook{TenantID: e.tenantID, ID: id, Enabled: &resumed})
	require.NoError(t, err)
	e.deliver(t)
	require.Len(t, receiver.requests(), 1, "the delivery queued before the pause is sent once resumed")
}

func TestWebhookDeliveryE2E_AnUnchangedStatusRecordsNothing(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	receiver := newWebhookReceiver(t)
	e.createWebhook(t, receiver.URL, models.WebhookEventDeviceRejected)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusRejected))
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusRejected))
	e.deliver(t)

	require.Len(t, receiver.requests(), 1)
}

func TestWebhookDeliveryE2E_ARolledBackChangeRecordsNothing(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	receiver := newWebhookReceiver(t)
	e.createWebhook(t, receiver.URL, models.WebhookEventDeviceAccepted)

	first := e.enrollAs(t, "same-host", "aa:bb:cc:dd:ee:01", "pk-1", "")
	second := e.enrollAs(t, "same-host", "aa:bb:cc:dd:ee:02", "pk-2", "")

	require.NoError(t, e.setStatus(t, first, models.DeviceStatusAccepted))
	require.ErrorIs(t, e.setStatus(t, second, models.DeviceStatusAccepted), ErrDeviceDuplicated)
	e.deliver(t)

	got := receiver.requests()
	require.Len(t, got, 1, "the refused accept must not leave an event behind")

	var event deliveredEvent
	require.NoError(t, json.Unmarshal(got[0].body, &event))
	require.Equal(t, first, event.Data.Device.UID)
}

func TestWebhookDeliveryE2E_AFailedDeliveryIsRetriedWithBackoffThenAbandoned(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	receiver := newWebhookReceiver(t)
	receiver.answer(http.StatusInternalServerError)
	webhookID, _ := e.createWebhook(t, receiver.URL, models.WebhookEventDeviceAccepted)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusAccepted))

	e.deliver(t)
	require.Len(t, receiver.requests(), 1)

	delivery := e.deliveries(t, webhookID)[0]
	require.Equal(t, models.WebhookDeliveryPending, delivery.Status)
	require.Equal(t, 1, delivery.Attempts)
	require.Equal(t, http.StatusInternalServerError, delivery.LastStatusCode)
	require.Equal(t, "webhook returned status 500", delivery.LastError)
	require.WithinRange(t, delivery.NextAttemptAt, now.Add(time.Minute), now.Add(time.Minute+6*time.Second))

	e.deliver(t)
	require.Len(t, receiver.requests(), 1, "a retry waits for its backoff")

	for attempt := 2; attempt <= webhookMaxAttempts; attempt++ {
		e.clock.Set(e.deliveries(t, webhookID)[0].NextAttemptAt)
		e.deliver(t)
		require.Len(t, receiver.requests(), attempt)
	}

	delivery = e.deliveries(t, webhookID)[0]
	require.Equal(t, models.WebhookDeliveryFailed, delivery.Status)
	require.Equal(t, webhookMaxAttempts, delivery.Attempts)

	e.clock.Set(e.clock.Now().Add(24 * time.Hour))
	e.deliver(t)
	require.Len(t, receiver.requests(), webhookMaxAttempts, "an abandoned delivery is not sent again")
}

func TestWebhookDeliveryE2E_AReceiverThatNeverAnswersIsRetried(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	silent := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		<-r.Context().Done()
	}))
	t.Cleanup(silent.Close)
	webhookID, _ := e.createWebhook(t, silent.URL, models.WebhookEventDeviceAccepted)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusAccepted))

	started := time.Now() //nolint:forbidigo // measures how long the real request takes, which the test clock cannot
	e.deliver(t)
	require.Less(t, time.Since(started), webhookRequestTimeout+5*time.Second, "the request gives up at its timeout")

	delivery := e.deliveries(t, webhookID)[0]
	require.Equal(t, models.WebhookDeliveryPending, delivery.Status)
	require.Equal(t, 1, delivery.Attempts)
	require.Zero(t, delivery.LastStatusCode)
	require.Equal(t, "the request timed out", delivery.LastError)
	require.WithinRange(t, delivery.NextAttemptAt, now.Add(time.Minute), now.Add(time.Minute+6*time.Second))
}

func TestWebhookDeliveryE2E_ARetrySucceedsAndKeepsTheEventID(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	receiver := newWebhookReceiver(t)
	receiver.answer(http.StatusServiceUnavailable)
	webhookID, _ := e.createWebhook(t, receiver.URL, models.WebhookEventDeviceAccepted)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusAccepted))
	e.deliver(t)

	receiver.answer(http.StatusNoContent)
	e.clock.Set(e.deliveries(t, webhookID)[0].NextAttemptAt)
	e.deliver(t)

	got := receiver.requests()
	require.Len(t, got, 2)
	require.Equal(t, got[0].header.Get("X-ShellHub-Event-ID"), got[1].header.Get("X-ShellHub-Event-ID"))

	delivery := e.deliveries(t, webhookID)[0]
	require.Equal(t, models.WebhookDeliveryDelivered, delivery.Status)
	require.Equal(t, 2, delivery.Attempts)
	require.NotNil(t, delivery.DeliveredAt)
}

func TestWebhookDeliveryE2E_ConcurrentRunsSendEachEventOnce(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	receiver := newWebhookReceiver(t)
	e.createWebhook(t, receiver.URL, models.WebhookEventDeviceAccepted)

	const devices = 30
	for i := range devices {
		mac := "aa:bb:cc:dd:ee:" + strconv.FormatInt(int64(0x10+i), 16)
		uid := e.enrollAs(t, "host-"+strconv.Itoa(i), mac, "pk-"+mac, "")
		require.NoError(t, e.setStatus(t, uid, models.DeviceStatusAccepted))
	}

	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() { assert.NoError(t, e.svc.DeliverWebhooks()(context.Background(), nil)) })
	}
	wg.Wait()

	seen := map[string]int{}
	for _, got := range receiver.requests() {
		seen[got.header.Get("X-ShellHub-Event-ID")]++
	}

	require.Len(t, seen, devices)
	for id, count := range seen {
		require.Equal(t, 1, count, "event %s was sent more than once", id)
	}
}
