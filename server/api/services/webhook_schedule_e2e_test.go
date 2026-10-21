package services

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/shellhub-io/shellhub/server/api/pkg/webhook"
	"github.com/stretchr/testify/require"
)

type submittedTasks struct {
	mu       sync.Mutex
	patterns []worker.TaskPattern
}

func (s *submittedTasks) Submit(_ context.Context, pattern worker.TaskPattern, _ []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.patterns = append(s.patterns, pattern)

	return nil
}

func (*submittedTasks) Close() error { return nil }

func (s *submittedTasks) submitted() []worker.TaskPattern {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]worker.TaskPattern(nil), s.patterns...)
}

func TestWebhookScheduleE2E_ACommittedEventAsksForDeliveryAtOnce(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	tasks := &submittedTasks{}
	e.svc = NewService(e.st, privateKey, publicKey, storecache.NewNullCache(), WithClock(e.clock),
		WithEnrollmentWebhookCIDRs(webhook.ParseCIDRs("127.0.0.0/8,::1/128")), WithWebhookDispatcher(tasks))
	receiver := subscribeToEveryDeviceEvent(t, e)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.Equal(t, []worker.TaskPattern{TaskWebhookDelivery}, tasks.submitted(), "registration asks for delivery")

	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusAccepted))
	require.Len(t, tasks.submitted(), 2, "a status change asks for delivery")

	require.NoError(t, e.svc.DeleteDevice(context.Background(), models.UID(uid), e.tenantID))
	require.Len(t, tasks.submitted(), 3, "a removal asks for delivery")

	require.NoError(t, e.svc.WebhookDeliverySweep()(context.Background()))
	require.Len(t, receiver.requests(), 3, "the sweep sends what the asked-for deliveries would")
}

func TestWebhookScheduleE2E_CleanupDeletesOnlyOldFinishedDeliveries(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	receiver := newWebhookReceiver(t)
	webhookID, _ := e.createWebhook(t, receiver.URL, models.WebhookEventDeviceAccepted, models.WebhookEventDeviceRejected)

	delivered := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, delivered, models.DeviceStatusAccepted))
	e.deliver(t)

	receiver.answer(http.StatusInternalServerError)
	stillTrying := e.enroll(t, "aa:bb:cc:dd:ee:02", "")
	require.NoError(t, e.setStatus(t, stillTrying, models.DeviceStatusRejected))
	e.deliver(t)

	e.clock.Set(now.Add(webhookDeliveryRetention - time.Minute))
	require.NoError(t, e.svc.WebhookDeliveryCleanup()(context.Background()))
	require.Len(t, e.deliveries(t, webhookID), 2, "nothing is old enough yet")

	e.clock.Set(now.Add(webhookDeliveryRetention + time.Minute))
	require.NoError(t, e.svc.WebhookDeliveryCleanup()(context.Background()))

	left := e.deliveries(t, webhookID)
	require.Len(t, left, 1)
	require.Equal(t, models.WebhookDeliveryPending, left[0].Status, "a delivery still being tried is kept")
}

func TestWebhookScheduleE2E_DeletingTheNamespaceDeletesItsWebhooks(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	ctx := context.Background()
	e.createWebhook(t, "https://hooks.example.com/a", models.WebhookEventDeviceAccepted)

	require.NoError(t, e.svc.DeleteNamespace(ctx, e.tenantID))

	_, count, err := e.svc.ListWebhooks(ctx, &requests.ListWebhooks{TenantID: e.tenantID, Paginator: query.Paginator{Page: 1, PerPage: 10}})
	require.NoError(t, err)
	require.Zero(t, count)
}
