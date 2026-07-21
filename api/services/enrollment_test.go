package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEvaluateEnrollment(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	svc := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	// The request MAC is upper-cased; the key's allowlist is stored normalized (lower), so a match
	// must be case-insensitive.
	req := requests.DeviceAuth{TenantID: "00000000-0000-4000-0000-000000000000", Identity: &requests.DeviceIdentity{MAC: "AA:BB:CC:DD:EE:FF"}}

	cases := []struct {
		description string
		key         *models.InstallKey
		paired      bool
		expected    enrollmentDecision
	}{
		{"a keyless enrollment (nil key) lands pending", nil, false, enrollPending},
		{"automatic accepts", &models.InstallKey{Mode: models.InstallKeyModeAutomatic}, false, enrollAccept},
		{"manual stays pending", &models.InstallKey{Mode: models.InstallKeyModeManual}, false, enrollPending},
		{"allowlist accepts a listed MAC (case-insensitive)", &models.InstallKey{Mode: models.InstallKeyModeAllowlist, AllowedMACs: []string{"aa:bb:cc:dd:ee:ff"}}, false, enrollAccept},
		{"allowlist rejects an unlisted MAC", &models.InstallKey{Mode: models.InstallKeyModeAllowlist, AllowedMACs: []string{"11:22:33:44:55:66"}}, false, enrollReject},
		{"an unknown mode stays pending", &models.InstallKey{Mode: "bogus"}, false, enrollPending},
		// The pairing-code flow is its own acceptance: paired accepts regardless of the key's mode, so a
		// manual/allowlist-miss pairing key still accepts (and never fires the webhook/reject).
		{"paired accepts despite a manual key", &models.InstallKey{Mode: models.InstallKeyModeManual}, true, enrollAccept},
		{"paired accepts despite an allowlist miss", &models.InstallKey{Mode: models.InstallKeyModeAllowlist, AllowedMACs: []string{"11:22:33:44:55:66"}}, true, enrollAccept},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			got := svc.evaluateEnrollment(context.Background(), tc.key, req, "uid", "host", tc.paired)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestEvaluateEnrollmentWebhook(t *testing.T) {
	// The integrator runs on loopback in these tests; permit it through the SSRF guard's allowlist.
	prevEnv := envs.DefaultBackend
	env := envmock.NewMockBackend(t)
	env.On("Get", enrollmentWebhookAllowedCIDRsEnv).Return("127.0.0.0/8,::1/128").Maybe()
	env.On("Get", mock.Anything).Return("").Maybe()
	envs.DefaultBackend = env
	t.Cleanup(func() { envs.DefaultBackend = prevEnv })

	// callEnrollmentWebhook stamps the payload with clock.Now(); the package clock is the shared mock.
	clockMock.On("Now").Return(now).Maybe()

	storeMock := storemock.NewMockStore(t)
	svc := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	req := requests.DeviceAuth{
		TenantID: "00000000-0000-4000-0000-000000000000",
		Identity: &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
		Info:     &requests.DeviceInfo{ID: "debian"},
		RealIP:   "203.0.113.7",
	}

	const secret = "s3cr3t"
	keyFor := func(url string) *models.InstallKey {
		return &models.InstallKey{Name: "ci", TenantID: req.TenantID, Mode: models.InstallKeyModeWebhook, WebhookURL: url, WebhookSecret: secret}
	}

	t.Run("honors the integrator decision and signs the request", func(t *testing.T) {
		for _, decision := range []enrollmentDecision{enrollAccept, enrollReject, enrollPending} {
			var gotSignature string
			var gotBody []byte

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotSignature = r.Header.Get("X-ShellHub-Signature")
				gotBody, _ = io.ReadAll(r.Body)
				_ = json.NewEncoder(w).Encode(enrollmentWebhookResponse{Decision: string(decision)})
			}))

			got := svc.evaluateEnrollment(context.Background(), keyFor(srv.URL), req, "uid", "host", false)
			require.Equal(t, decision, got)
			// The integrator can trust the request: the signature it received matches HMAC(secret, body).
			require.Equal(t, signEnrollmentWebhook(secret, gotBody), gotSignature)

			srv.Close()
		}
	})

	t.Run("fails closed to pending on a server error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		require.Equal(t, enrollPending, svc.evaluateEnrollment(context.Background(), keyFor(srv.URL), req, "uid", "host", false))
	})

	t.Run("fails closed to pending when the integrator is unreachable", func(t *testing.T) {
		require.Equal(t, enrollPending, svc.evaluateEnrollment(context.Background(), keyFor("http://127.0.0.1:1"), req, "uid", "host", false))
	})

	t.Run("fails closed to pending on an unrecognized decision", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(enrollmentWebhookResponse{Decision: "maybe"})
		}))
		defer srv.Close()

		require.Equal(t, enrollPending, svc.evaluateEnrollment(context.Background(), keyFor(srv.URL), req, "uid", "host", false))
	})

	t.Run("fails closed to pending on a cancelled context", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(enrollmentWebhookResponse{Decision: string(enrollAccept)})
		}))
		defer srv.Close()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		require.Equal(t, enrollPending, svc.evaluateEnrollment(ctx, keyFor(srv.URL), req, "uid", "host", false))
	})
}

func TestEnrollmentWebhookClientSSRF(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	withEnv := func(allowedCIDRs string) func() {
		env := envmock.NewMockBackend(t)
		env.On("Get", enrollmentWebhookAllowedCIDRsEnv).Return(allowedCIDRs).Maybe()
		env.On("Get", mock.Anything).Return("").Maybe()

		prev := envs.DefaultBackend
		envs.DefaultBackend = env

		return func() { envs.DefaultBackend = prev }
	}

	t.Run("blocks a loopback webhook target by default", func(t *testing.T) {
		defer withEnv("")()

		//nolint:noctx // the SSRF guard rejects the dial before any request is sent
		_, err := enrollmentWebhookClient().Get(srv.URL)
		require.Error(t, err, "a webhook to loopback must be refused so it can't reach the host's own services")
	})

	t.Run("allows a target the operator explicitly permits", func(t *testing.T) {
		defer withEnv("127.0.0.0/8,::1/128")()

		//nolint:noctx // exercising the dial path in a test
		resp, err := enrollmentWebhookClient().Get(srv.URL)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
	})
}
