package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/services"
	servicemock "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	webhookTestTenant = "00000000-0000-4000-0000-000000000000"
	webhookTestID     = "6f1c2b3a-0000-4000-8000-000000000001"
)

type webhookCaller struct {
	role   string
	apiKey bool
}

func (c webhookCaller) headers() map[string]string {
	headers := map[string]string{"Content-Type": "application/json", "X-Tenant-ID": webhookTestTenant, "X-Role": c.role}
	if c.apiKey {
		headers["X-API-KEY"] = "b2f7cc0e-d933-4aad-9ab2-b557f2f2554f"
	} else {
		headers["X-ID"] = "000000000000000000000000"
	}

	return headers
}

var (
	webhookOwner    = webhookCaller{role: "owner"}
	webhookAdmin    = webhookCaller{role: "administrator"}
	webhookOperator = webhookCaller{role: "operator"}
	webhookObserver = webhookCaller{role: "observer"}
	webhookAdminKey = webhookCaller{role: "administrator", apiKey: true}
)

func serveWebhookRoute(t *testing.T, svc services.Service, method, path string, caller webhookCaller, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload string
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		payload = string(data)
	}

	req := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(payload))
	for k, v := range caller.headers() {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	newTestRouter(t, svc).ServeHTTP(rec, req)

	return rec
}

func TestCreateWebhookRoute(t *testing.T) {
	valid := map[string]any{"url": "https://hooks.example.com/shellhub", "events": []string{"device.accepted"}}

	cases := []struct {
		description string
		caller      webhookCaller
		body        map[string]any
		mocks       func(*servicemock.MockService)
		status      int
	}{
		{description: "refuses an API key, since the response carries the secret", caller: webhookAdminKey, body: valid, status: http.StatusForbidden},
		{description: "refuses an operator", caller: webhookOperator, body: valid, status: http.StatusForbidden},
		{description: "refuses an observer", caller: webhookObserver, body: valid, status: http.StatusForbidden},
		{description: "refuses a missing URL", caller: webhookOwner, body: map[string]any{"events": []string{"device.accepted"}}, status: http.StatusBadRequest},
		{description: "refuses a value that is not a URL", caller: webhookOwner, body: map[string]any{"url": "hooks", "events": []string{"device.accepted"}}, status: http.StatusBadRequest},
		{description: "refuses no event types", caller: webhookOwner, body: map[string]any{"url": "https://hooks.example.com", "events": []string{}}, status: http.StatusBadRequest},
		{description: "refuses an unknown event type", caller: webhookOwner, body: map[string]any{"url": "https://hooks.example.com", "events": []string{"device.renamed"}}, status: http.StatusBadRequest},
		{
			description: "reports the namespace limit",
			caller:      webhookOwner,
			body:        valid,
			mocks: func(svc *servicemock.MockService) {
				svc.On("CreateWebhook", mock.Anything, mock.Anything).Return(nil, services.NewErrWebhookLimit(10)).Once()
			},
			status: http.StatusForbidden,
		},
		{
			description: "creates as an administrator",
			caller:      webhookAdmin,
			body:        valid,
			mocks: func(svc *servicemock.MockService) {
				svc.On("CreateWebhook", mock.Anything, &requests.CreateWebhook{
					TenantID: webhookTestTenant,
					URL:      "https://hooks.example.com/shellhub",
					Events:   []models.WebhookEventType{models.WebhookEventDeviceAccepted},
				}).Return(&responses.CreateWebhook{Webhook: models.Webhook{ID: webhookTestID}, Secret: "s3cr3t"}, nil).Once()
			},
			status: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svc := servicemock.NewMockService(t)
			if tc.mocks != nil {
				tc.mocks(svc)
			}

			rec := serveWebhookRoute(t, svc, http.MethodPost, "/api/namespaces/webhooks", tc.caller, tc.body)
			require.Equal(t, tc.status, rec.Code)

			if tc.status == http.StatusOK {
				var got map[string]any
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
				require.Equal(t, "s3cr3t", got["secret"], "the secret is shown once, on create")
			}
		})
	}
}

func TestListAndGetWebhookRoutesNeverReturnTheSecret(t *testing.T) {
	stored := models.Webhook{ID: webhookTestID, TenantID: webhookTestTenant, URL: "https://hooks.example.com", SecretEncrypted: "ciphertext"}

	t.Run("list", func(t *testing.T) {
		svc := servicemock.NewMockService(t)
		svc.On("ListWebhooks", mock.Anything, mock.MatchedBy(func(req *requests.ListWebhooks) bool { return req.TenantID == webhookTestTenant })).
			Return([]models.Webhook{stored}, int64(1), nil).Once()

		rec := serveWebhookRoute(t, svc, http.MethodGet, "/api/namespaces/webhooks", webhookAdminKey, nil)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "1", rec.Header().Get("X-Total-Count"))
		require.NotContains(t, rec.Body.String(), "secret")
		require.NotContains(t, rec.Body.String(), "ciphertext")
	})

	t.Run("get", func(t *testing.T) {
		svc := servicemock.NewMockService(t)
		svc.On("GetWebhook", mock.Anything, &requests.WebhookID{TenantID: webhookTestTenant, ID: webhookTestID}).Return(&stored, nil).Once()

		rec := serveWebhookRoute(t, svc, http.MethodGet, "/api/namespaces/webhooks/"+webhookTestID, webhookAdmin, nil)
		require.Equal(t, http.StatusOK, rec.Code)
		require.NotContains(t, rec.Body.String(), "secret")
		require.NotContains(t, rec.Body.String(), "ciphertext")
	})
}

func TestWebhookRoutePermissions(t *testing.T) {
	cases := []struct {
		description string
		method      string
		path        string
		body        any
		caller      webhookCaller
		mocks       func(*servicemock.MockService)
		status      int
	}{
		{description: "an observer cannot list", method: http.MethodGet, path: "/api/namespaces/webhooks", caller: webhookObserver, status: http.StatusForbidden},
		{description: "an operator cannot get", method: http.MethodGet, path: "/api/namespaces/webhooks/" + webhookTestID, caller: webhookOperator, status: http.StatusForbidden},
		{description: "an API key cannot update, since it could point the URL at itself", method: http.MethodPatch, path: "/api/namespaces/webhooks/" + webhookTestID, body: map[string]any{"url": "https://attacker.example.com"}, caller: webhookAdminKey, status: http.StatusForbidden},
		{description: "an operator cannot update", method: http.MethodPatch, path: "/api/namespaces/webhooks/" + webhookTestID, body: map[string]any{"enabled": false}, caller: webhookOperator, status: http.StatusForbidden},
		{description: "an administrator can update", method: http.MethodPatch, path: "/api/namespaces/webhooks/" + webhookTestID, body: map[string]any{"enabled": false}, caller: webhookAdmin, status: http.StatusOK, mocks: func(svc *servicemock.MockService) {
			disabled := false
			svc.On("UpdateWebhook", mock.Anything, &requests.UpdateWebhook{TenantID: webhookTestTenant, ID: webhookTestID, Enabled: &disabled}).Return(&models.Webhook{ID: webhookTestID}, nil).Once()
		}},
		{description: "an update with an unknown event type is refused", method: http.MethodPatch, path: "/api/namespaces/webhooks/" + webhookTestID, body: map[string]any{"events": []string{"device.renamed"}}, caller: webhookOwner, status: http.StatusBadRequest},
		{description: "an operator cannot delete", method: http.MethodDelete, path: "/api/namespaces/webhooks/" + webhookTestID, caller: webhookOperator, status: http.StatusForbidden},
		{description: "an owner can delete", method: http.MethodDelete, path: "/api/namespaces/webhooks/" + webhookTestID, caller: webhookOwner, status: http.StatusOK, mocks: func(svc *servicemock.MockService) {
			svc.On("DeleteWebhook", mock.Anything, &requests.WebhookID{TenantID: webhookTestTenant, ID: webhookTestID}).Return(nil).Once()
		}},
		{description: "a missing webhook is not found", method: http.MethodDelete, path: "/api/namespaces/webhooks/" + webhookTestID, caller: webhookOwner, status: http.StatusNotFound, mocks: func(svc *servicemock.MockService) {
			svc.On("DeleteWebhook", mock.Anything, mock.Anything).Return(services.NewErrWebhookNotFound(webhookTestID, nil)).Once()
		}},
		{description: "an ID that is not a UUID is refused", method: http.MethodDelete, path: "/api/namespaces/webhooks/not-a-uuid", caller: webhookOwner, status: http.StatusBadRequest},
		{description: "an API key cannot reveal the secret", method: http.MethodGet, path: "/api/namespaces/webhooks/" + webhookTestID + "/reveal", caller: webhookAdminKey, status: http.StatusForbidden},
		{description: "an operator cannot reveal the secret", method: http.MethodGet, path: "/api/namespaces/webhooks/" + webhookTestID + "/reveal", caller: webhookOperator, status: http.StatusForbidden},
		{description: "an administrator can reveal the secret", method: http.MethodGet, path: "/api/namespaces/webhooks/" + webhookTestID + "/reveal", caller: webhookAdmin, status: http.StatusOK, mocks: func(svc *servicemock.MockService) {
			svc.On("RevealWebhookSecret", mock.Anything, &requests.WebhookID{TenantID: webhookTestTenant, ID: webhookTestID}).Return("s3cr3t", nil).Once()
		}},
		{description: "an observer cannot list deliveries", method: http.MethodGet, path: "/api/namespaces/webhooks/" + webhookTestID + "/deliveries", caller: webhookObserver, status: http.StatusForbidden},
		{description: "an API key can list deliveries", method: http.MethodGet, path: "/api/namespaces/webhooks/" + webhookTestID + "/deliveries", caller: webhookAdminKey, status: http.StatusOK, mocks: func(svc *servicemock.MockService) {
			svc.On("ListWebhookDeliveries", mock.Anything, mock.MatchedBy(func(req *requests.ListWebhookDeliveries) bool { return req.ID == webhookTestID })).
				Return([]models.WebhookDelivery{}, int64(0), nil).Once()
		}},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svc := servicemock.NewMockService(t)
			if tc.mocks != nil {
				tc.mocks(svc)
			}

			rec := serveWebhookRoute(t, svc, tc.method, tc.path, tc.caller, tc.body)
			require.Equal(t, tc.status, rec.Code)
		})
	}
}
