package main

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/moby/moby/api/types/container"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	enrollmentWebhookPort   = "9876"
	enrollmentWebhookSecret = "s3cr3t"
)

// TestEnrollmentWebhook enrolls devices with webhook-mode keys whose integrator is a stub that runs
// in the server's network namespace: it answers each call with the decision its path names and
// logs every call it receives. The server reaches it on loopback, which the stack allows the
// webhook to call; every other private address stays behind the SSRF guard.
func TestEnrollmentWebhook(t *testing.T) {
	compose := newConfiguredSSHEnvironment(t, t.Context(),
		environment.New(t, run).WithEnv("SHELLHUB_PROVISIONING_KEY_WEBHOOK_ALLOWED_CIDRS", "127.0.0.0/8"),
		models.SSHAccessModeLegacy)
	webhook := startEnrollmentWebhook(t, compose)

	newKey := func(t *testing.T, req *requests.CreateProvisioningKey) string {
		t.Helper()

		req.Mode = string(models.ProvisioningKeyModeWebhook)
		req.WebhookSecret = enrollmentWebhookSecret

		return compose.CreateProvisioningKey(t, req).Key
	}

	t.Run("the integrator's answer decides the device", func(t *testing.T) {
		cases := []struct {
			decision string
			status   models.DeviceStatus
			mac      string
		}{
			{decision: "accept", status: models.DeviceStatusAccepted, mac: "02:00:00:00:80:01"},
			{decision: "reject", status: models.DeviceStatusRejected, mac: "02:00:00:00:80:02"},
			{decision: "defer", status: models.DeviceStatusPending, mac: "02:00:00:00:80:03"},
		}

		for _, tc := range cases {
			t.Run("an answer to "+tc.decision, func(t *testing.T) {
				name := "webhook-" + tc.decision
				key := newKey(t, &requests.CreateProvisioningKey{Name: name, WebhookURL: webhookURL("/" + tc.decision)})

				device := enroll(t, compose, newKeyedDeviceAuthRequest(t, key, name, tc.mac))

				assert.Equal(t, tc.status, device.Status)
				webhook.awaitCalls(t, device.UID, 1)
			})
		}
	})

	t.Run("the call is signed with the key's secret and describes the enrollment", func(t *testing.T) {
		key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
			Name:          "webhook-payload",
			Mode:          string(models.ProvisioningKeyModeWebhook),
			WebhookURL:    webhookURL("/defer"),
			WebhookSecret: enrollmentWebhookSecret,
		})
		req := newKeyedDeviceAuthRequest(t, key.Key, "webhook-payload", "02:00:00:00:80:04")
		before := clock.Now()

		device := enroll(t, compose, req)

		call := webhook.awaitCalls(t, device.UID, 1)[0]

		mac := hmac.New(sha256.New, []byte(enrollmentWebhookSecret))
		mac.Write([]byte(call.Body))
		assert.Equal(t, hex.EncodeToString(mac.Sum(nil)), call.Signature)

		payload := call.payload(t)
		assert.Equal(t, ShellHubNamespace, payload.TenantID)
		assert.Equal(t, key.ID, payload.ProvisioningKeyID)
		assert.Equal(t, "webhook-payload", payload.ProvisioningKeyName)
		assert.Equal(t, device.UID, payload.DeviceUID)
		assert.Equal(t, "02:00:00:00:80:04", payload.Identity)
		assert.Equal(t, "webhook-payload", payload.Hostname)
		assert.Equal(t, req.Info, payload.Info)
		assert.NotEmpty(t, payload.SourceIP)
		assert.WithinRange(t, payload.Timestamp, before.Add(-time.Minute), clock.Now().Add(time.Minute))

		callback, err := url.Parse(payload.CallbackURL)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(callback.Path, "/api/devices/enroll/callback/"), payload.CallbackURL)
	})

	t.Run("an integrator slower than the key's timeout leaves the device pending", func(t *testing.T) {
		const delay = 10 * time.Second

		key := newKey(t, &requests.CreateProvisioningKey{
			Name:           "webhook-slow",
			WebhookURL:     webhookURL("/accept?delay=" + delay.String()),
			WebhookTimeout: 1,
		})

		started := clock.Now()
		device := enroll(t, compose, newKeyedDeviceAuthRequest(t, key, "webhook-slow", "02:00:00:00:80:05"))

		assert.Less(t, clock.Now().Sub(started), models.ProvisioningKeyWebhookDefaultTimeout*time.Second,
			"the server waited for the integrator past the key's timeout")
		assert.Equal(t, models.DeviceStatusPending, device.Status)
		webhook.awaitCalls(t, device.UID, 1)
	})

	t.Run("a deferred decision is delivered through the callback", func(t *testing.T) {
		cases := []struct {
			decision string
			status   models.DeviceStatus
			mac      string
		}{
			{decision: "accept", status: models.DeviceStatusAccepted, mac: "02:00:00:00:80:06"},
			{decision: "reject", status: models.DeviceStatusRejected, mac: "02:00:00:00:80:07"},
		}

		for _, tc := range cases {
			t.Run("a callback to "+tc.decision, func(t *testing.T) {
				name := "callback-" + tc.decision
				key := newKey(t, &requests.CreateProvisioningKey{Name: name, WebhookURL: webhookURL("/defer")})

				device := enroll(t, compose, newKeyedDeviceAuthRequest(t, key, name, tc.mac))
				require.Equal(t, models.DeviceStatusPending, device.Status)

				callback := webhook.awaitCalls(t, device.UID, 1)[0].payload(t).CallbackURL

				resp := postEnrollmentCallback(t, compose, callback, tc.decision)
				require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

				assert.Equal(t, tc.status, requireDevice(t, compose, device.UID).Status)
			})
		}
	})

	t.Run("a callback redeems its decision only once", func(t *testing.T) {
		key := newKey(t, &requests.CreateProvisioningKey{Name: "callback-once", WebhookURL: webhookURL("/defer")})

		device := enroll(t, compose, newKeyedDeviceAuthRequest(t, key, "callback-once", "02:00:00:00:80:08"))
		callback := webhook.awaitCalls(t, device.UID, 1)[0].payload(t).CallbackURL

		resp := postEnrollmentCallback(t, compose, callback, "accept")
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		resp = postEnrollmentCallback(t, compose, callback, "reject")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), resp.String())

		assert.Equal(t, models.DeviceStatusAccepted, requireDevice(t, compose, device.UID).Status)
	})

	t.Run("a callback past the key's callback window decides nothing", func(t *testing.T) {
		key := newKey(t, &requests.CreateProvisioningKey{
			Name:               "callback-expired",
			WebhookURL:         webhookURL("/defer"),
			WebhookCallbackTTL: 1,
		})

		device := enroll(t, compose, newKeyedDeviceAuthRequest(t, key, "callback-expired", "02:00:00:00:80:09"))
		callback := webhook.awaitCalls(t, device.UID, 1)[0].payload(t).CallbackURL

		time.Sleep(3 * time.Second)

		resp := postEnrollmentCallback(t, compose, callback, "accept")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), resp.String())

		assert.Equal(t, models.DeviceStatusPending, requireDevice(t, compose, device.UID).Status)
	})

	t.Run("a pending device is decided on its next authentication once the integrator answers", func(t *testing.T) {
		key := newKey(t, &requests.CreateProvisioningKey{Name: "webhook-reconciled", WebhookURL: webhookURL("/defer")})
		req := newKeyedDeviceAuthRequest(t, key, "webhook-reconciled", "02:00:00:00:80:10")

		device := enroll(t, compose, req)
		require.Equal(t, models.DeviceStatusPending, device.Status)

		compose.UpdateProvisioningKey(t, "webhook-reconciled", map[string]any{"webhook_url": webhookURL("/accept")})

		awaitStatusOnReauth(t, compose, req, device.UID, models.DeviceStatusAccepted)

		calls := webhook.awaitCalls(t, device.UID, 2)
		assert.Equal(t, "/accept", calls[1].Path)
	})

	t.Run("a pending device asks the integrator again at most once a minute", func(t *testing.T) {
		key := newKey(t, &requests.CreateProvisioningKey{Name: "webhook-throttled", WebhookURL: webhookURL("/defer")})
		req := newKeyedDeviceAuthRequest(t, key, "webhook-throttled", "02:00:00:00:80:11")

		device := enroll(t, compose, req)

		var calls []enrollmentWebhookCall

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			resp, err := postDeviceAuth(t.Context(), compose, req)
			if !assert.NoError(tt, err) || !assert.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String()) {
				return
			}

			calls, err = webhook.calls(t.Context(), device.UID)
			assert.NoError(tt, err)
			assert.Len(tt, calls, 3)
		}, deviceAuthCacheTTL+2*models.EnrollmentReconcileInterval, 2*time.Second)

		reconciled, again := calls[1].payload(t).Timestamp, calls[2].payload(t).Timestamp
		assert.GreaterOrEqual(t, again.Sub(reconciled), models.EnrollmentReconcileInterval,
			"re-authenticating every 2 seconds past a %s cache asked the integrator again after only %s",
			deviceAuthCacheTTL, again.Sub(reconciled))
		assert.Equal(t, models.DeviceStatusPending, requireDevice(t, compose, device.UID).Status)
	})

	t.Run("an integrator on a private address outside the allowed ranges is never called", func(t *testing.T) {
		key := newKey(t, &requests.CreateProvisioningKey{
			Name:       "webhook-private",
			WebhookURL: "http://" + string(environment.ServiceServer) + ":" + enrollmentWebhookPort + "/accept",
		})

		device := enroll(t, compose, newKeyedDeviceAuthRequest(t, key, "webhook-private", "02:00:00:00:80:12"))

		assert.Equal(t, models.DeviceStatusPending, device.Status)
		assert.Never(t, func() bool {
			calls, err := webhook.calls(t.Context(), device.UID)

			return err != nil || len(calls) > 0
		}, 5*time.Second, time.Second, "the server called the integrator at the server's own network address")
	})
}

func webhookURL(path string) string {
	return "http://127.0.0.1:" + enrollmentWebhookPort + path
}

func postEnrollmentCallback(t *testing.T, compose *environment.DockerCompose, callbackURL, decision string) *resty.Response {
	t.Helper()

	resp, err := compose.Anonymous(t.Context()).
		SetBody(map[string]string{"decision": decision}).
		Post(callbackURL)
	require.NoError(t, err)

	return resp
}

type enrollmentWebhook struct {
	container testcontainers.Container
}

type enrollmentWebhookCall struct {
	Path      string `json:"path"`
	Signature string `json:"signature"`
	Body      string `json:"body"`
}

type enrollmentWebhookPayload struct {
	TenantID            string               `json:"tenant_id"`
	ProvisioningKeyID   string               `json:"provisioning_key_id"`
	ProvisioningKeyName string               `json:"provisioning_key_name"`
	DeviceUID           string               `json:"device_uid"`
	Identity            string               `json:"identity"`
	Hostname            string               `json:"hostname"`
	Info                *requests.DeviceInfo `json:"info"`
	SourceIP            string               `json:"source_ip"`
	Timestamp           time.Time            `json:"timestamp"`
	CallbackURL         string               `json:"callback_url"`
}

func (c enrollmentWebhookCall) payload(t *testing.T) enrollmentWebhookPayload {
	t.Helper()

	var payload enrollmentWebhookPayload
	require.NoError(t, json.Unmarshal([]byte(c.Body), &payload))

	return payload
}

func startEnrollmentWebhook(t *testing.T, compose *environment.DockerCompose) *enrollmentWebhook {
	t.Helper()

	server := compose.Service(environment.ServiceServer).GetContainerID()

	c, err := testcontainers.GenericContainer(t.Context(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Repo:      "enrollment-webhook",
				Tag:       "test",
				Context:   "cmd/enrollment-webhook",
				KeepImage: true,
			},
			Env: map[string]string{"ADDRESS": ":" + enrollmentWebhookPort},
			HostConfigModifier: func(hc *container.HostConfig) {
				hc.NetworkMode = container.NetworkMode("container:" + server)
			},
			WaitingFor: wait.ForLog("listening"),
		},
		Started: true,
		Logger:  log.New(io.Discard, "", log.LstdFlags),
	})
	require.NoError(t, err)

	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	return &enrollmentWebhook{container: c}
}

func (w *enrollmentWebhook) calls(ctx context.Context, uid string) ([]enrollmentWebhookCall, error) {
	reader, err := w.container.Logs(ctx)
	if err != nil {
		return nil, err
	}

	defer reader.Close() //nolint:errcheck // the logs are read in full before it closes

	calls := []enrollmentWebhookCall{}

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		var call enrollmentWebhookCall
		if json.Unmarshal(scanner.Bytes(), &call) != nil || call.Path == "" {
			continue
		}

		var payload enrollmentWebhookPayload
		if err := json.Unmarshal([]byte(call.Body), &payload); err != nil {
			return nil, err
		}

		if payload.DeviceUID == uid {
			calls = append(calls, call)
		}
	}

	return calls, scanner.Err()
}

func (w *enrollmentWebhook) awaitCalls(t *testing.T, uid string, count int) []enrollmentWebhookCall {
	t.Helper()

	var calls []enrollmentWebhookCall

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		var err error

		calls, err = w.calls(t.Context(), uid)
		assert.NoError(tt, err)
		assert.Len(tt, calls, count)
	}, 10*time.Second, 500*time.Millisecond)

	return calls
}
