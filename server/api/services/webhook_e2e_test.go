package services

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
)

func (e *enrollmentE2E) createWebhook(t *testing.T, url string, events ...models.WebhookEventType) (string, string) {
	t.Helper()

	created, err := e.svc.CreateWebhook(context.Background(), &requests.CreateWebhook{
		TenantID: e.tenantID,
		URL:      url,
		Events:   events,
	})
	require.NoError(t, err)

	return created.ID, created.Secret
}

func TestWebhookE2E_CreateRevealsTheSecretOnlyOnCreateAndReveal(t *testing.T) {
	e := setupEnrollmentE2E(t)
	ctx := context.Background()

	id, secret := e.createWebhook(t, "https://hooks.example.com/shellhub", models.WebhookEventDeviceAccepted)
	require.NotEmpty(t, secret)

	webhooks, count, err := e.svc.ListWebhooks(ctx, &requests.ListWebhooks{TenantID: e.tenantID, Paginator: query.Paginator{Page: 1, PerPage: 10}})
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
	require.Equal(t, "https://hooks.example.com/shellhub", webhooks[0].URL)
	require.Equal(t, []models.WebhookEventType{models.WebhookEventDeviceAccepted}, webhooks[0].Events)
	require.True(t, webhooks[0].Enabled, "a new webhook starts enabled")

	revealed, err := e.svc.RevealWebhookSecret(ctx, &requests.WebhookID{TenantID: e.tenantID, ID: id})
	require.NoError(t, err)
	require.Equal(t, secret, revealed)

	_, otherSecret := e.createWebhook(t, "https://hooks.example.com/other", models.WebhookEventDeviceAccepted)
	require.NotEqual(t, secret, otherSecret, "each webhook gets its own secret")
}

func TestWebhookE2E_UpdateChangesOnlyTheGivenFields(t *testing.T) {
	e := setupEnrollmentE2E(t)
	ctx := context.Background()

	id, _ := e.createWebhook(t, "https://hooks.example.com/a", models.WebhookEventDeviceAccepted)

	disabled := false
	_, err := e.svc.UpdateWebhook(ctx, &requests.UpdateWebhook{
		TenantID: e.tenantID,
		ID:       id,
		Events:   []models.WebhookEventType{models.WebhookEventDevicePending, models.WebhookEventDeviceRemoved},
		Enabled:  &disabled,
	})
	require.NoError(t, err)

	webhook, err := e.svc.GetWebhook(ctx, &requests.WebhookID{TenantID: e.tenantID, ID: id})
	require.NoError(t, err)
	require.Equal(t, "https://hooks.example.com/a", webhook.URL, "an omitted URL is left unchanged")
	require.ElementsMatch(t, []models.WebhookEventType{models.WebhookEventDevicePending, models.WebhookEventDeviceRemoved}, webhook.Events)
	require.False(t, webhook.Enabled)

	url := "http://hooks.example.com/b"
	_, err = e.svc.UpdateWebhook(ctx, &requests.UpdateWebhook{TenantID: e.tenantID, ID: id, URL: &url})
	require.NoError(t, err)

	webhook, err = e.svc.GetWebhook(ctx, &requests.WebhookID{TenantID: e.tenantID, ID: id})
	require.NoError(t, err)
	require.Equal(t, url, webhook.URL)
	require.False(t, webhook.Enabled, "an omitted enabled flag is left unchanged")
}

func TestWebhookE2E_ConcurrentUpdatesKeepBothChanges(t *testing.T) {
	e := setupEnrollmentE2E(t)
	ctx := context.Background()
	sc := scope.MustBounded(e.tenantID)

	id, _ := e.createWebhook(t, "https://hooks.example.com/a", models.WebhookEventDeviceAccepted)
	events := []models.WebhookEventType{models.WebhookEventDevicePending}

	commit := e.heldOpen(t, func(ctx context.Context) error {
		if err := e.st.WebhookLockNamespace(ctx, sc); err != nil {
			return err
		}

		webhook, err := e.st.WebhookResolve(ctx, sc, id)
		if err != nil {
			return err
		}

		webhook.Events = events

		return e.st.WebhookUpdate(ctx, webhook)
	})

	updated := make(chan error, 1)
	go func() {
		disabled := false
		_, err := e.svc.UpdateWebhook(ctx, &requests.UpdateWebhook{TenantID: e.tenantID, ID: id, Enabled: &disabled})
		updated <- err
	}()

	e.waitUntilBlockedOnLock(t)
	commit()
	require.NoError(t, <-updated)

	webhook, err := e.svc.GetWebhook(ctx, &requests.WebhookID{TenantID: e.tenantID, ID: id})
	require.NoError(t, err)
	require.ElementsMatch(t, events, webhook.Events)
	require.False(t, webhook.Enabled)
}

func TestWebhookE2E_DeleteRemovesTheWebhook(t *testing.T) {
	e := setupEnrollmentE2E(t)
	ctx := context.Background()

	id, _ := e.createWebhook(t, "https://hooks.example.com/a", models.WebhookEventDeviceAccepted)

	require.NoError(t, e.svc.DeleteWebhook(ctx, &requests.WebhookID{TenantID: e.tenantID, ID: id}))

	_, err := e.svc.GetWebhook(ctx, &requests.WebhookID{TenantID: e.tenantID, ID: id})
	require.ErrorIs(t, err, ErrWebhookNotFound)

	require.ErrorIs(t, e.svc.DeleteWebhook(ctx, &requests.WebhookID{TenantID: e.tenantID, ID: id}), ErrWebhookNotFound)
}

func TestWebhookE2E_AnotherNamespaceCannotReachTheWebhook(t *testing.T) {
	e := setupEnrollmentE2E(t)
	ctx := context.Background()

	id, _ := e.createWebhook(t, "https://hooks.example.com/a", models.WebhookEventDeviceAccepted)
	other := "00000000-0000-4000-0000-0000000000ff"

	_, err := e.svc.GetWebhook(ctx, &requests.WebhookID{TenantID: other, ID: id})
	require.ErrorIs(t, err, ErrWebhookNotFound)

	_, err = e.svc.RevealWebhookSecret(ctx, &requests.WebhookID{TenantID: other, ID: id})
	require.ErrorIs(t, err, ErrWebhookNotFound)

	require.ErrorIs(t, e.svc.DeleteWebhook(ctx, &requests.WebhookID{TenantID: other, ID: id}), ErrWebhookNotFound)
}

func TestWebhookE2E_CreateRefusesAnInvalidWebhook(t *testing.T) {
	e := setupEnrollmentE2E(t)
	ctx := context.Background()

	cases := []struct {
		description string
		url         string
		events      []models.WebhookEventType
	}{
		{description: "a URL that is not http or https", url: "ftp://hooks.example.com", events: []models.WebhookEventType{models.WebhookEventDeviceAccepted}},
		{description: "no event types", url: "https://hooks.example.com", events: nil},
		{description: "an unknown event type", url: "https://hooks.example.com", events: []models.WebhookEventType{"device.renamed"}},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			_, err := e.svc.CreateWebhook(ctx, &requests.CreateWebhook{TenantID: e.tenantID, URL: tc.url, Events: tc.events})
			require.ErrorIs(t, err, ErrWebhookInvalidField)
		})
	}
}

func TestWebhookE2E_CreateStopsAtTheNamespaceLimit(t *testing.T) {
	e := setupEnrollmentE2E(t)

	for range webhookLimit {
		e.createWebhook(t, "https://hooks.example.com/a", models.WebhookEventDeviceAccepted)
	}

	_, err := e.svc.CreateWebhook(context.Background(), &requests.CreateWebhook{
		TenantID: e.tenantID,
		URL:      "https://hooks.example.com/a",
		Events:   []models.WebhookEventType{models.WebhookEventDeviceAccepted},
	})
	require.ErrorIs(t, err, ErrWebhookLimit)
}

func TestWebhookE2E_ConcurrentCreatesStopAtTheNamespaceLimit(t *testing.T) {
	e := setupEnrollmentE2E(t)

	var (
		wg      sync.WaitGroup
		created atomic.Int64
		refused atomic.Int64
	)
	for range 2 * webhookLimit {
		wg.Go(func() {
			_, err := e.svc.CreateWebhook(context.Background(), &requests.CreateWebhook{
				TenantID: e.tenantID,
				URL:      "https://hooks.example.com/a",
				Events:   []models.WebhookEventType{models.WebhookEventDeviceAccepted},
			})
			switch {
			case err == nil:
				created.Add(1)
			case errors.Is(err, ErrWebhookLimit):
				refused.Add(1)
			default:
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
	wg.Wait()

	require.Equal(t, int64(webhookLimit), created.Load())
	require.Equal(t, int64(webhookLimit), refused.Load())
}
