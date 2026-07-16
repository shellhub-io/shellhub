package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/storetest/pgprovider"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// plaintextFor / digest give a deterministic plaintext key and its stored SHA256 digest, so a test can
// present the plaintext at enrollment and have it resolve to the seeded key.
func plaintextFor(b byte) string { return fmt.Sprintf("install-key-%02x", b) }
func digest(b byte) string       { return hashInstallKey(plaintextFor(b)) }

// enrollmentE2E wires a real service over a real Postgres store (testcontainers) so enrollment can be
// exercised end to end: AuthDevice -> policy decision -> device status -> list-by-status -> accept,
// the exact path an external integrator drives.
type enrollmentE2E struct {
	svc      *APIService
	st       store.Store
	tenantID string
}

func setupEnrollmentE2E(t *testing.T) *enrollmentE2E {
	t.Helper()
	ctx := context.Background()

	provider, err := pgprovider.NewProvider(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { provider.Close(t) })

	st := provider.Store()

	// Isolate clock and env from the package-global mocks so this test controls them fully. Use the
	// real "now" captured at TestMain (not a fixed past date) so signed callback tokens' expiry is
	// valid against the JWT library's wall-clock check.
	fixed := now
	localClock := clockmock.NewMockClock(t)
	localClock.On("Now").Return(fixed).Maybe()
	prevClock := clock.DefaultBackend
	clock.DefaultBackend = localClock

	localEnv := envmock.NewMockBackend(t)
	// The webhook tests point their install key at an httptest server on loopback, which the SSRF guard
	// blocks by default; permit it via the operator allowlist so the real dial path is still exercised.
	localEnv.On("Get", "SHELLHUB_INSTALL_KEY_WEBHOOK_ALLOWED_CIDRS").Return("127.0.0.0/8,::1/128").Maybe()
	localEnv.On("Get", mock.Anything).Return("").Maybe() // community edition: not cloud, not enterprise
	prevEnv := envs.DefaultBackend
	envs.DefaultBackend = localEnv

	// A prior unit test may have leaked a fixed-value uuid mock as DefaultBackend; this e2e uses the
	// real store, which needs genuinely unique IDs (e.g. a re-registered device's second history event
	// must not collide with the first). Reset to the real generator.
	prevUUID := uuid.DefaultBackend
	uuid.DefaultBackend = realUUIDBackend

	t.Cleanup(func() {
		clock.DefaultBackend = prevClock
		envs.DefaultBackend = prevEnv
		uuid.DefaultBackend = prevUUID
	})

	owner, err := st.UserCreate(ctx, &models.User{
		Origin:        models.UserOriginLocal,
		Status:        models.UserStatusConfirmed,
		MaxNamespaces: -1,
		UserData:      models.UserData{Name: "owner", Email: "owner@example.com", Username: "owner"},
		Password:      models.UserPassword{Hash: "hash"},
	})
	require.NoError(t, err)

	tenantID, err := st.NamespaceCreate(ctx, &models.Namespace{
		Name:       "e2e",
		Owner:      owner,
		MaxDevices: -1,
		Members:    []models.Member{{ID: owner, Role: authorizer.RoleOwner}},
		Settings:   &models.NamespaceSettings{},
	})
	require.NoError(t, err)

	svc := NewService(st, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	return &enrollmentE2E{svc: svc, st: st, tenantID: tenantID}
}

// legacyKey inserts the namespace's system (legacy) install key in the given mode.
func (e *enrollmentE2E) installKey(t *testing.T, digest, name string, mode models.InstallKeyMode, system bool, opts func(*models.InstallKey)) {
	t.Helper()
	key := &models.InstallKey{
		ID:        digest,
		Name:      name,
		TenantID:  e.tenantID,
		Mode:      mode,
		System:    system,
		Reusable:  true,
		Tags:      []string{},
		CreatedBy: "00000000-0000-4000-0000-000000000009",
	}
	if opts != nil {
		opts(key)
	}
	_, err := e.st.InstallKeyCreate(context.Background(), key)
	require.NoError(t, err)
}

// enroll performs a device auth (enrollment) and returns the resulting device UID.
func (e *enrollmentE2E) enroll(t *testing.T, mac, installKey string) string {
	t.Helper()
	req := requests.DeviceAuth{
		TenantID:       e.tenantID,
		Hostname:       "host-" + mac,
		Identity:       &requests.DeviceIdentity{MAC: mac},
		Info:           &requests.DeviceInfo{ID: "debian", PrettyName: "Debian", Version: "v0.1.0", Arch: "amd64", Platform: "docker"},
		PublicKey:      "pk-" + mac,
		RealIP:         "203.0.113.7",
		ForwardedHost:  "shellhub.test",
		ForwardedProto: "https",
	}
	if installKey != "" {
		req.InstallKey = installKey
	}

	res, err := e.svc.AuthDevice(context.Background(), req)
	require.NoError(t, err)

	return res.UID
}

func (e *enrollmentE2E) status(t *testing.T, uid string) models.DeviceStatus {
	t.Helper()
	device, err := e.st.DeviceResolve(context.Background(), store.DeviceUIDResolver, uid)
	require.NoError(t, err)

	return device.Status
}

func (e *enrollmentE2E) usedTimes(t *testing.T, keyDigest string) int {
	t.Helper()
	key, err := e.st.InstallKeyResolve(context.Background(), store.InstallKeyIDResolver, keyDigest, e.st.Options().InNamespace(e.tenantID))
	require.NoError(t, err)

	return key.UsedTimes
}

// clearSecret blanks the reveal fields, keeping a seeded presentable key minimal.
func clearSecret(k *models.InstallKey) { k.KeyEncrypted, k.KeyHint = "", "" }

func (e *enrollmentE2E) events(t *testing.T, keyName string) []models.InstallKeyEvent {
	t.Helper()
	// History is keyed by the key's id (digest); resolve the seeded key's name to its id first.
	key, err := e.st.InstallKeyResolve(context.Background(), store.InstallKeyNameResolver, keyName, e.st.Options().InNamespace(e.tenantID))
	require.NoError(t, err)

	events, _, err := e.svc.ListInstallKeyEvents(context.Background(), &requests.ListInstallKeyEvents{
		TenantID:  e.tenantID,
		ID:        key.ID,
		Paginator: query.Paginator{Page: 1, PerPage: 100},
	})
	require.NoError(t, err)

	return events
}

func (e *enrollmentE2E) pendingCount(t *testing.T) int {
	t.Helper()
	_, count, err := e.svc.ListDevices(context.Background(), &requests.DeviceList{
		TenantID:     e.tenantID,
		DeviceStatus: models.DeviceStatusPending,
		Paginator:    query.Paginator{Page: 1, PerPage: 100},
	})
	require.NoError(t, err)

	return count
}

// deviceCounts returns the namespace's per-status device counters (accepted/pending/removed).
func (e *enrollmentE2E) deviceCounts(t *testing.T) models.Namespace {
	t.Helper()
	ns, err := e.st.NamespaceResolve(context.Background(), store.NamespaceTenantIDResolver, e.tenantID)
	require.NoError(t, err)

	return *ns
}

// TestEnrollmentE2E_LegacyKeyless is the compatibility gate: a device that presents only the tenant ID
// (no install key) — like a 40k field fleet — resolves the legacy manual key, lands pending, is
// listable via ?status=pending, and is accepted through the canonical status endpoint.
func TestEnrollmentE2E_LegacyKeyless(t *testing.T) {
	e := setupEnrollmentE2E(t)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")

	require.Equal(t, models.DeviceStatusPending, e.status(t, uid), "keyless device must land pending")
	require.Equal(t, 1, e.pendingCount(t), "device must be listable via ?status=pending")

	// The legacy key's history is the keyless-pending queue: the event carries the device's live
	// status, so the UI can accept from there.
	events := e.events(t, "legacy")
	require.Len(t, events, 1)
	require.Equal(t, uid, events[0].DeviceUID)
	require.Equal(t, models.DeviceStatusPending, events[0].DeviceStatus)

	// The integrator accepts it through the same endpoint it uses today.
	require.NoError(t, e.svc.UpdateDeviceStatus(context.Background(), &requests.DeviceUpdateStatus{
		TenantID: e.tenantID, UID: uid, Status: string(models.DeviceStatusAccepted),
	}))

	require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))
	require.Equal(t, 0, e.pendingCount(t), "device must leave the pending queue once accepted")

	// After acceptance the history reflects the new live status (drives the UI refresh).
	require.Equal(t, models.DeviceStatusAccepted, e.events(t, "legacy")[0].DeviceStatus)
}

// TestEnrollmentE2E_Modes covers the per-key policy outcomes end to end.
func TestEnrollmentE2E_Modes(t *testing.T) {
	e := setupEnrollmentE2E(t)

	t.Run("automatic accepts on enrollment", func(t *testing.T) {
		e.installKey(t, digest(0x01), "auto", models.InstallKeyModeAutomatic, false, func(k *models.InstallKey) {
			k.KeyEncrypted, k.KeyHint = "", ""
		})
		uid := e.enroll(t, "aa:bb:cc:dd:ee:10", plaintextFor(0x01))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))

		// The auto-accept must freeze its outcome onto the enrollment event so the audit keeps it after
		// the device is removed. Regression: the decision was stamped before the event was appended, so a
		// fresh automatic enrollment left DecidedStatus empty.
		evs := e.events(t, "auto")
		require.Len(t, evs, 1)
		require.Equal(t, models.DeviceStatusAccepted, evs[0].DecidedStatus)
		require.NotNil(t, evs[0].DecidedAt)
	})

	t.Run("manual key lands pending", func(t *testing.T) {
		e.installKey(t, digest(0x02), "manual", models.InstallKeyModeManual, false, func(k *models.InstallKey) {
			k.KeyEncrypted, k.KeyHint = "", ""
		})
		uid := e.enroll(t, "aa:bb:cc:dd:ee:20", plaintextFor(0x02))
		require.Equal(t, models.DeviceStatusPending, e.status(t, uid))
	})

	t.Run("allowlist accepts a listed MAC and rejects others", func(t *testing.T) {
		e.installKey(t, digest(0x03), "allow", models.InstallKeyModeAllowlist, false, func(k *models.InstallKey) {
			k.AllowedMACs = []string{"aa:bb:cc:dd:ee:31"}
			k.KeyEncrypted, k.KeyHint = "", ""
		})
		accepted := e.enroll(t, "aa:bb:cc:dd:ee:31", plaintextFor(0x03))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, accepted))

		rejected := e.enroll(t, "aa:bb:cc:dd:ee:32", plaintextFor(0x03))
		require.Equal(t, models.DeviceStatusRejected, e.status(t, rejected))

		// Both the auto-accept and the auto-reject must freeze their outcome onto their own event.
		decisions := map[string]models.DeviceStatus{}
		for _, ev := range e.events(t, "allow") {
			decisions[ev.DeviceUID] = ev.DecidedStatus
		}
		require.Equal(t, models.DeviceStatusAccepted, decisions[accepted])
		require.Equal(t, models.DeviceStatusRejected, decisions[rejected])
	})

	t.Run("the auth response carries the device status", func(t *testing.T) {
		// The response exposes the enrollment status so an agent can react to its authorization state
		// instead of connecting blind.
		e.installKey(t, digest(0x04), "auto-status", models.InstallKeyModeAutomatic, false, clearSecret)
		res, err := e.svc.AuthDevice(context.Background(), requests.DeviceAuth{
			TenantID:   e.tenantID,
			Hostname:   "host-status",
			Identity:   &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:44"},
			Info:       &requests.DeviceInfo{ID: "debian", PrettyName: "Debian", Version: "v0.1.0", Arch: "amd64", Platform: "docker"},
			PublicKey:  "pk-status",
			RealIP:     "203.0.113.7",
			InstallKey: plaintextFor(0x04),
		})
		require.NoError(t, err)
		require.Equal(t, models.DeviceStatusAccepted, res.Status)
	})
}

// TestEnrollmentE2E_ReregisterAndReaccept covers the re-registration (policy re-runs), reconnect
// (policy does not re-run), and re-accept (an automated reject is recoverable) semantics.
func TestEnrollmentE2E_ReregisterAndReaccept(t *testing.T) {
	e := setupEnrollmentE2E(t)

	t.Run("re-registration re-runs the policy and consumes another use", func(t *testing.T) {
		e.installKey(t, digest(0x40), "auto", models.InstallKeyModeAutomatic, false, clearSecret)

		uid := e.enroll(t, "aa:bb:cc:dd:ee:40", plaintextFor(0x40))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))
		require.Equal(t, 1, e.usedTimes(t, digest(0x40)))

		// Soft-remove the accepted device, then re-enroll: it is a fresh enrollment, so the policy runs
		// again and accepts it, consuming another use.
		require.NoError(t, e.svc.DeleteDevice(context.Background(), models.UID(uid), e.tenantID))

		uid2 := e.enroll(t, "aa:bb:cc:dd:ee:40", plaintextFor(0x40))
		require.Equal(t, uid, uid2)
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid2))
		require.Equal(t, 2, e.usedTimes(t, digest(0x40)))

		// The re-registration must leave the namespace counters consistent: one accepted device, nothing
		// pending or removed. Regression: re-evaluating the policy read the device's stale "removed"
		// status (the pending transition was not persisted before deciding), so the accept double-counted
		// the removed decrement and left a phantom pending.
		counts := e.deviceCounts(t)
		require.Equal(t, int64(1), counts.DevicesAcceptedCount)
		require.Equal(t, int64(0), counts.DevicesPendingCount)
		require.Equal(t, int64(0), counts.DevicesRemovedCount)
	})

	t.Run("a plain reconnect does not re-run the policy", func(t *testing.T) {
		e.installKey(t, digest(0x41), "auto2", models.InstallKeyModeAutomatic, false, clearSecret)

		uid := e.enroll(t, "aa:bb:cc:dd:ee:41", plaintextFor(0x41))
		require.Equal(t, 1, e.usedTimes(t, digest(0x41)))

		// Same identity, not removed: a reconnect must leave the status and the usage counter untouched.
		e.enroll(t, "aa:bb:cc:dd:ee:41", plaintextFor(0x41))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))
		require.Equal(t, 1, e.usedTimes(t, digest(0x41)), "reconnect must not re-run the policy or consume a use")
	})

	t.Run("an auto-rejected device can be manually re-accepted", func(t *testing.T) {
		e.installKey(t, digest(0x42), "allow", models.InstallKeyModeAllowlist, false, func(k *models.InstallKey) {
			k.AllowedMACs = []string{"aa:bb:cc:dd:ee:99"}
			clearSecret(k)
		})

		uid := e.enroll(t, "aa:bb:cc:dd:ee:42", plaintextFor(0x42))
		require.Equal(t, models.DeviceStatusRejected, e.status(t, uid))

		// Rejected is not a dead end: accept it through the canonical endpoint.
		require.NoError(t, e.svc.UpdateDeviceStatus(context.Background(), &requests.DeviceUpdateStatus{
			TenantID: e.tenantID, UID: uid, Status: string(models.DeviceStatusAccepted),
		}))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))
	})
}

// TestEnrollmentE2E_WebhookDeferCallback covers the async path: the integrator answers "defer", the
// device lands pending, and the emitted callback URL is redeemed later (no API key) to accept it.
func TestEnrollmentE2E_WebhookDeferCallback(t *testing.T) {
	e := setupEnrollmentE2E(t)

	var callbackURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			CallbackURL string `json:"callback_url"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		callbackURL = payload.CallbackURL
		_ = json.NewEncoder(w).Encode(map[string]string{"decision": "defer"})
	}))
	defer srv.Close()

	e.installKey(t, digest(0x50), "webhook", models.InstallKeyModeWebhook, false, func(k *models.InstallKey) {
		k.WebhookURL = srv.URL
		k.WebhookSecret = "s3cr3t"
		clearSecret(k)
	})

	uid := e.enroll(t, "aa:bb:cc:dd:ee:50", plaintextFor(0x50))
	require.Equal(t, models.DeviceStatusPending, e.status(t, uid), "defer lands the device pending")
	require.NotEmpty(t, callbackURL, "a callback URL must be offered to the integrator")

	token := callbackURL[strings.LastIndex(callbackURL, "/")+1:]

	t.Run("a garbage token is unauthorized", func(t *testing.T) {
		err := e.svc.ResolveEnrollmentCallback(context.Background(), &requests.EnrollmentCallback{Token: "garbage", Decision: "accept"})
		require.Error(t, err)
		require.Equal(t, models.DeviceStatusPending, e.status(t, uid))
	})

	t.Run("the real token redeems the deferred decision", func(t *testing.T) {
		require.NoError(t, e.svc.ResolveEnrollmentCallback(context.Background(), &requests.EnrollmentCallback{Token: token, Decision: "accept"}))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))
	})
}

// TestEnrollmentE2E_CallbackSingleUse proves the callback token is single-use: once redeemed, replaying
// the same URL is refused even for a decision the device state would otherwise still allow (rejected ->
// accepted), closing the accept<->reject flip a still-valid token would enable.
func TestEnrollmentE2E_CallbackSingleUse(t *testing.T) {
	e := setupEnrollmentE2E(t)

	var callbackURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			CallbackURL string `json:"callback_url"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		callbackURL = payload.CallbackURL
		_ = json.NewEncoder(w).Encode(map[string]string{"decision": "defer"})
	}))
	defer srv.Close()

	e.installKey(t, digest(0x60), "webhook", models.InstallKeyModeWebhook, false, func(k *models.InstallKey) {
		k.WebhookURL = srv.URL
		k.WebhookSecret = "s3cr3t"
		clearSecret(k)
	})

	uid := e.enroll(t, "aa:bb:cc:dd:ee:60", plaintextFor(0x60))
	require.Equal(t, models.DeviceStatusPending, e.status(t, uid))
	require.NotEmpty(t, callbackURL)
	token := callbackURL[strings.LastIndex(callbackURL, "/")+1:]

	// First redemption: reject. Rejected is not terminal for a later accept (the re-accept feature), so
	// the device-state guard alone would let a replayed accept flip it.
	require.NoError(t, e.svc.ResolveEnrollmentCallback(context.Background(), &requests.EnrollmentCallback{Token: token, Decision: "reject"}))
	require.Equal(t, models.DeviceStatusRejected, e.status(t, uid))

	// Replay the same URL with the opposite decision: the jti is already spent, so it is refused and the
	// device stays rejected. Without the single-use gate this would flip rejected -> accepted.
	err := e.svc.ResolveEnrollmentCallback(context.Background(), &requests.EnrollmentCallback{Token: token, Decision: "accept"})
	require.Error(t, err)
	require.Equal(t, models.DeviceStatusRejected, e.status(t, uid))
}

// TestEnrollmentE2E_CallbackHonorsKeyState proves the deferred-callback accept path mirrors the
// synchronous accept: it reserves a use against the key's limit and refuses once the key is no longer
// valid, so an outstanding token can't bypass the usage cap or accept with a key revoked after mint.
func TestEnrollmentE2E_CallbackHonorsKeyState(t *testing.T) {
	e := setupEnrollmentE2E(t)

	// enrollDeferred spins a defer-only integrator, enrolls a device with a fresh single-use webhook key,
	// and returns the device UID plus the single-use callback token the integrator was handed.
	enrollDeferred := func(keyByte byte, name, mac string) (string, string) {
		t.Helper()
		var callbackURL string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var payload struct {
				CallbackURL string `json:"callback_url"`
			}
			_ = json.NewDecoder(r.Body).Decode(&payload)
			callbackURL = payload.CallbackURL
			_ = json.NewEncoder(w).Encode(map[string]string{"decision": "defer"})
		}))
		defer srv.Close()

		e.installKey(t, digest(keyByte), name, models.InstallKeyModeWebhook, false, func(k *models.InstallKey) {
			k.WebhookURL = srv.URL
			k.WebhookSecret = "s3cr3t"
			k.UsageLimit = 1
			clearSecret(k)
		})

		uid := e.enroll(t, mac, plaintextFor(keyByte))
		require.Equal(t, models.DeviceStatusPending, e.status(t, uid))
		require.NotEmpty(t, callbackURL)

		return uid, callbackURL[strings.LastIndex(callbackURL, "/")+1:]
	}

	t.Run("the callback accept reserves a use against the key limit", func(t *testing.T) {
		uid, token := enrollDeferred(0x62, "webhook-limit", "aa:bb:cc:dd:ee:62")

		require.NoError(t, e.svc.ResolveEnrollmentCallback(context.Background(), &requests.EnrollmentCallback{Token: token, Decision: "accept"}))
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))
		require.Equal(t, 1, e.usedTimes(t, digest(0x62)), "a deferred accept must consume a use, like the synchronous path")
	})

	t.Run("a key revoked after the token is minted can't accept via callback", func(t *testing.T) {
		uid, token := enrollDeferred(0x63, "webhook-revoked", "aa:bb:cc:dd:ee:63")

		key, err := e.st.InstallKeyResolve(context.Background(), store.InstallKeyIDResolver, digest(0x63), e.st.Options().InNamespace(e.tenantID))
		require.NoError(t, err)
		key.Revoked = true
		require.NoError(t, e.st.InstallKeyUpdate(context.Background(), key))

		err = e.svc.ResolveEnrollmentCallback(context.Background(), &requests.EnrollmentCallback{Token: token, Decision: "accept"})
		require.Error(t, err, "a callback must not accept with a key revoked after the token was minted")
		require.Equal(t, models.DeviceStatusPending, e.status(t, uid))
		require.Equal(t, 0, e.usedTimes(t, digest(0x63)), "a refused accept must not consume a use")
	})
}

// TestEnrollmentE2E_ReconcilePending covers the reconcile path: a webhook device that landed pending
// (integrator deferred) is re-evaluated on the agent's later AuthDevice, throttled by
// EnrollmentReconcileInterval, and accepted once the integrator recovers — without a callback or a
// restart.
func TestEnrollmentE2E_ReconcilePending(t *testing.T) {
	e := setupEnrollmentE2E(t)

	// A mutable clock so the reconcile throttle window can be crossed deterministically.
	base := now
	cur := base
	clk := clockmock.NewMockClock(t)
	clk.On("Now").Return(func() time.Time { return cur }).Maybe()
	clock.DefaultBackend = clk

	decision := "defer"
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_ = json.NewEncoder(w).Encode(map[string]string{"decision": decision})
	}))
	defer srv.Close()

	e.installKey(t, digest(0x60), "webhook", models.InstallKeyModeWebhook, false, func(k *models.InstallKey) {
		k.WebhookURL = srv.URL
		k.WebhookSecret = "s3cr3t"
		clearSecret(k)
	})

	// Enrollment: the integrator defers, so the device lands pending (webhook consulted once).
	uid := e.enroll(t, "aa:bb:cc:dd:ee:60", plaintextFor(0x60))
	require.Equal(t, models.DeviceStatusPending, e.status(t, uid), "defer lands the device pending")
	require.Equal(t, 1, hits)

	// First phone-home after enrollment reconciles (no prior attempt to throttle against); the
	// integrator still defers, so it stays pending but the attempt is now stamped.
	e.enroll(t, "aa:bb:cc:dd:ee:60", plaintextFor(0x60))
	require.Equal(t, 2, hits)
	require.Equal(t, models.DeviceStatusPending, e.status(t, uid))

	// A second phone-home within the throttle window must not re-consult the integrator.
	e.enroll(t, "aa:bb:cc:dd:ee:60", plaintextFor(0x60))
	require.Equal(t, 2, hits, "a re-auth within the throttle window must not re-evaluate")
	require.Equal(t, models.DeviceStatusPending, e.status(t, uid))

	// Past the window, with the integrator now accepting, the next phone-home reconciles and accepts.
	decision = "accept"
	cur = base.Add(models.EnrollmentReconcileInterval + time.Minute)
	e.enroll(t, "aa:bb:cc:dd:ee:60", plaintextFor(0x60))
	require.Equal(t, 3, hits, "past the window the integrator is consulted again")
	require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid), "a recovered integrator resolves the device")
	require.Equal(t, 1, e.usedTimes(t, digest(0x60)), "the reconciled accept consumes one key use")

	// Reconcile is not a new enrollment: despite three re-evaluations (two pending, one accepting), the
	// history holds only the single original enrollment event. The device's outcome shows through the
	// history's live-joined status, not through extra rows.
	require.Len(t, e.events(t, "webhook"), 1, "reconcile must not append history events")

	t.Run("a terminal device is never reconciled", func(t *testing.T) {
		// The accepted device above keeps re-authing past the window: a terminal status is never
		// re-evaluated, so the integrator is not consulted again and the status stays put.
		cur = base.Add(10 * models.EnrollmentReconcileInterval)
		e.enroll(t, "aa:bb:cc:dd:ee:60", plaintextFor(0x60))
		require.Equal(t, 3, hits, "an accepted device must not re-consult the integrator")
		require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))
	})
}

// TestEnrollmentE2E_ReconcileSkipsInvalidKey proves reconcile respects the key's validity: a device
// left pending is not accepted on a later phone-home once its key has been revoked, matching the
// guarantee a fresh enrollment already has.
func TestEnrollmentE2E_ReconcileSkipsInvalidKey(t *testing.T) {
	e := setupEnrollmentE2E(t)

	base := now
	cur := base
	clk := clockmock.NewMockClock(t)
	clk.On("Now").Return(func() time.Time { return cur }).Maybe()
	clock.DefaultBackend = clk

	decision := "defer"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"decision": decision})
	}))
	defer srv.Close()

	e.installKey(t, digest(0x61), "webhook-revoked", models.InstallKeyModeWebhook, false, func(k *models.InstallKey) {
		k.WebhookURL = srv.URL
		k.WebhookSecret = "s3cr3t"
		clearSecret(k)
	})

	uid := e.enroll(t, "aa:bb:cc:dd:ee:61", plaintextFor(0x61))
	require.Equal(t, models.DeviceStatusPending, e.status(t, uid), "defer lands the device pending")

	// Revoke the key while the device sits pending.
	key, err := e.st.InstallKeyResolve(context.Background(), store.InstallKeyIDResolver, digest(0x61), e.st.Options().InNamespace(e.tenantID))
	require.NoError(t, err)
	key.Revoked = true
	require.NoError(t, e.st.InstallKeyUpdate(context.Background(), key))

	// Past the throttle window, with the integrator now accepting: reconcile must skip the revoked key,
	// so the device stays pending instead of being flipped to accepted.
	decision = "accept"
	cur = base.Add(models.EnrollmentReconcileInterval + time.Minute)
	e.enroll(t, "aa:bb:cc:dd:ee:61", plaintextFor(0x61))
	require.Equal(t, models.DeviceStatusPending, e.status(t, uid), "a revoked key must not reconcile a pending device to accepted")
}

// TestEnrollmentE2E_UsageLimitUnderConcurrency proves the usage limit holds when enrollments race: with
// a single-use key, many devices enrolling at once must yield exactly one accept, never more. A racer
// ends in one of three valid states: accepted (won the slot), pending (passed the stale pre-check but
// lost the reserve), or auth-invalid (the key was already exhausted when it resolved). The accept
// reserves a use against the atomic, limit-guarded counter first, so the counter — not the racy pre-
// check — decides who gets the slot, and it can never land past the limit.
func TestEnrollmentE2E_UsageLimitUnderConcurrency(t *testing.T) {
	e := setupEnrollmentE2E(t)
	e.installKey(t, digest(0x80), "single-use", models.InstallKeyModeAutomatic, false, func(k *models.InstallKey) {
		k.UsageLimit = 1
		clearSecret(k)
	})

	const racers = 8
	start := make(chan struct{})
	uids := make([]string, racers)
	errs := make([]error, racers)

	var wg sync.WaitGroup
	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := requests.DeviceAuth{
				TenantID:       e.tenantID,
				Hostname:       fmt.Sprintf("racer-%02d", i),
				Identity:       &requests.DeviceIdentity{MAC: fmt.Sprintf("aa:bb:cc:dd:80:%02x", i)},
				Info:           &requests.DeviceInfo{ID: "debian", PrettyName: "Debian", Version: "v0.1.0", Arch: "amd64", Platform: "docker"},
				PublicKey:      fmt.Sprintf("pk-80-%02d", i),
				RealIP:         "203.0.113.7",
				ForwardedHost:  "shellhub.test",
				ForwardedProto: "https",
				InstallKey:     plaintextFor(0x80),
			}

			<-start // release all racers together to maximize the overlap on the last slot
			res, err := e.svc.AuthDevice(context.Background(), req)
			if errs[i] = err; err == nil {
				uids[i] = res.UID
			}
		}(i)
	}
	close(start)
	wg.Wait()

	accepted := 0
	for i := 0; i < racers; i++ {
		if errs[i] != nil {
			// A racer that resolved the key after it was exhausted is rejected at auth (no device row);
			// that is the expected outcome, not a failure.
			require.ErrorContains(t, errs[i], "auth invalid")

			continue
		}
		if e.status(t, uids[i]) == models.DeviceStatusAccepted {
			accepted++
		}
	}

	require.Equal(t, 1, accepted, "a single-use key must accept exactly one racing device")
	require.Equal(t, 1, e.usedTimes(t, digest(0x80)), "the key's usage counter must land at its limit, not past it")
}

// TestEnrollmentE2E_HistoryCredential covers the history enrichment: the enrollment captures the
// device's public key, exposes it and (for a real key) its fingerprint, and live-joins the decision
// time (device status_updated_at) distinct from the immutable enrollment time.
func TestEnrollmentE2E_HistoryCredential(t *testing.T) {
	e := setupEnrollmentE2E(t)
	e.installKey(t, digest(0x70), "man", models.InstallKeyModeManual, false, clearSecret)

	// A mutable clock so the decision time can advance past the enrollment time deterministically.
	base := now
	cur := base
	clk := clockmock.NewMockClock(t)
	clk.On("Now").Return(func() time.Time { return cur }).Maybe()
	clock.DefaultBackend = clk

	uid := e.enroll(t, "aa:bb:cc:dd:ee:70", plaintextFor(0x70))
	require.Equal(t, models.DeviceStatusPending, e.status(t, uid))

	evs := e.events(t, "man")
	require.Len(t, evs, 1)
	require.Equal(t, "pk-aa:bb:cc:dd:ee:70", evs[0].PublicKey, "the enrollment key is captured")
	// The placeholder key is not a real PEM, so the fingerprint is empty (the SHA256 computation itself
	// is covered by the entity unit test). A pending device has no decision yet.
	require.Empty(t, evs[0].Fingerprint)
	require.Empty(t, evs[0].DecidedStatus)
	require.Nil(t, evs[0].DecidedAt)
	enrolledAt := evs[0].Timestamp

	// Accept it later: the event row's enrollment facts stay, and the decision is frozen onto it.
	cur = base.Add(time.Hour)
	require.NoError(t, e.svc.UpdateDeviceStatus(context.Background(), &requests.DeviceUpdateStatus{
		TenantID: e.tenantID, UID: uid, Status: string(models.DeviceStatusAccepted),
	}))

	evs = e.events(t, "man")
	require.Len(t, evs, 1, "accept does not add a history row")
	require.Equal(t, enrolledAt, evs[0].Timestamp, "the enrollment time stays fixed")
	require.Equal(t, models.DeviceStatusAccepted, evs[0].DecidedStatus, "the outcome is frozen on the event")
	require.NotNil(t, evs[0].DecidedAt)
	require.True(t, evs[0].DecidedAt.After(enrolledAt), "the decision time is the accept moment, after enrollment")

	// The whole point: remove the device and the frozen decision must remain in the audit (the live
	// status would go to removed and lose it).
	require.NoError(t, e.svc.DeleteDevice(context.Background(), models.UID(uid), e.tenantID))

	evs = e.events(t, "man")
	require.Len(t, evs, 1)
	require.Equal(t, models.DeviceStatusAccepted, evs[0].DecidedStatus, "the decision survives device removal")
	require.NotNil(t, evs[0].DecidedAt)
}

// TestEnrollmentE2E_HistoryCurrent covers the re-registration case: a device removed and re-enrolled
// with the same identity keeps one device row but gains a second event, and only the newest event is
// current (owns the live status/decision) so the older one doesn't borrow it.
func TestEnrollmentE2E_HistoryCurrent(t *testing.T) {
	e := setupEnrollmentE2E(t)
	e.installKey(t, digest(0x80), "man2", models.InstallKeyModeManual, false, clearSecret)

	// A mutable clock so the two events get distinct created_at (the "newest per device" window keys on
	// it); the shared setup clock is fixed, which would tie them.
	base := now
	cur := base
	clk := clockmock.NewMockClock(t)
	clk.On("Now").Return(func() time.Time { return cur }).Maybe()
	clock.DefaultBackend = clk

	accept := func(uid string) {
		require.NoError(t, e.svc.UpdateDeviceStatus(context.Background(), &requests.DeviceUpdateStatus{
			TenantID: e.tenantID, UID: uid, Status: string(models.DeviceStatusAccepted),
		}))
	}

	// First enrollment, accepted, then removed.
	uid := e.enroll(t, "aa:bb:cc:dd:ee:80", plaintextFor(0x80))
	cur = base.Add(time.Hour)
	accept(uid)
	firstDecidedAt := cur
	require.NoError(t, e.svc.DeleteDevice(context.Background(), models.UID(uid), e.tenantID))

	// Same identity re-enrolls (re-registration → same device row, second event), accepted later.
	cur = base.Add(3 * time.Hour)
	uid2 := e.enroll(t, "aa:bb:cc:dd:ee:80", plaintextFor(0x80))
	require.Equal(t, uid, uid2, "the same identity re-enrolls as the same device")
	cur = base.Add(4 * time.Hour)
	accept(uid2)
	secondDecidedAt := cur

	evs := e.events(t, "man2")
	require.Len(t, evs, 2, "re-enrollment adds a second event for the same device")

	newer, older := evs[0], evs[1]
	if older.Timestamp.After(newer.Timestamp) {
		newer, older = older, newer
	}
	require.True(t, newer.IsCurrent, "the newest event owns the live accept/reject action")
	require.False(t, older.IsCurrent, "the older event is historical, not current")

	// Each event keeps its OWN frozen decision — the accept doesn't leak onto both rows.
	require.Equal(t, models.DeviceStatusAccepted, older.DecidedStatus)
	require.Equal(t, models.DeviceStatusAccepted, newer.DecidedStatus)
	require.NotNil(t, older.DecidedAt)
	require.NotNil(t, newer.DecidedAt)
	require.WithinDuration(t, firstDecidedAt, *older.DecidedAt, time.Second, "older event keeps the first accept time")
	require.WithinDuration(t, secondDecidedAt, *newer.DecidedAt, time.Second, "newer event keeps the second accept time")
}
