package services

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/telemetry"
	"github.com/shellhub-io/shellhub/server/api/pkg/responses"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const telemetryInstanceID = "55555555-5555-4555-8555-555555555555"

var telemetryNow = time.Date(2026, 10, 21, 5, 30, 0, 0, time.UTC)

type recordingSender struct {
	sent []telemetry.Event
	err  error
}

func (r *recordingSender) Endpoint() string { return "http://collector.test" }

func (r *recordingSender) Send(events ...telemetry.Event) error {
	r.sent = append(r.sent, events...)

	return r.err
}

func telemetrySystem(enabled bool) *models.System {
	return &models.System{
		Setup:          true,
		InstanceID:     telemetryInstanceID,
		Telemetry:      &models.SystemTelemetry{Enabled: enabled},
		Authentication: &models.SystemAuthentication{Local: &models.SystemAuthenticationLocal{Enabled: true}},
	}
}

func newTelemetryService(t *testing.T, storeMock *storemock.MockStore, sender TelemetrySender, disabledByEnvironment bool) *APIService {
	t.Helper()

	return newTelemetryServiceWithCache(t, storeMock, sender, disabledByEnvironment, cache.NewNullCache())
}

func newTelemetryServiceWithCache(t *testing.T, storeMock *storemock.MockStore, sender TelemetrySender, disabledByEnvironment bool, c cache.Cache) *APIService {
	t.Helper()

	options := []Option{
		WithClock(clock.NewFixed(telemetryNow)),
		WithEdition(envs.Enterprise),
		WithSystemInfo("v0.27.0", "22"),
		WithTelemetry(sender),
	}
	if disabledByEnvironment {
		options = append(options, WithTelemetryDisabledByEnvironment())
	}

	return NewService(storeMock, privateKey, publicKey, c, options...)
}

func expectedInstanceHeartbeat() telemetry.Event {
	return telemetry.Event{
		Name:       "instance heartbeat",
		DistinctID: telemetryInstanceID,
		Properties: map[string]any{
			"edition":          "enterprise",
			"version":          "v0.27.0",
			"os":               runtime.GOOS,
			"arch":             runtime.GOARCH,
			"database":         "postgres",
			"namespaces":       int64(3),
			"users":            int64(4),
			"devices_accepted": int64(10),
			"devices_pending":  int64(2),
			"devices_rejected": int64(1),
			"sessions":         int64(25),

			"sessions_interactive":       int64(11),
			"sessions_shell_without_pty": int64(12),
			"sessions_exec":              int64(13),
			"sessions_scp":               int64(14),
			"sessions_sftp":              int64(15),
			"sessions_port_forwarding":   int64(16),
			"sessions_agent_forwarding":  int64(17),
			"sessions_web":               int64(18),
			"sessions_recorded":          int64(19),
			"sessions_api_key":           int64(20),

			"sessions_unauthenticated":         int64(51),
			"sessions_active":                  int64(52),
			"sessions_devices":                 int64(53),
			"sessions_events":                  int64(54),
			"sessions_duration_seconds":        int64(55),
			"sessions_duration_median_seconds": int64(56),
			"sessions_duration_p95_seconds":    int64(57),

			"namespaces_legacy":                      int64(21),
			"namespaces_personal":                    int64(22),
			"namespaces_team":                        int64(23),
			"namespaces_recording":                   int64(24),
			"namespaces_require_verified_device_key": int64(25),

			"devices_online":        int64(31),
			"devices_ephemeral":     int64(32),
			"devices_owned":         int64(33),
			"devices_custom_fields": int64(34),
			"devices_tagged":        int64(35),
			"devices_key_verified":  int64(36),
			"devices_platforms":     map[string]int64{"docker": 7, "connector": 2, "other": 1},
			"devices_versions":      map[string]int64{"v0.27": 7, "other": 2, "unknown": 1},
			"devices_architectures": map[string]int64{"amd64": 9, "other": 1},
			"devices_oses":          map[string]int64{"ubuntu": 7, "unknown": 1, "other": 2},
			"devices_enrolled_with": map[string]int64{},

			"access_policies":     int64(41),
			"public_keys":         int64(42),
			"api_keys":            int64(43),
			"instance_api_keys":   int64(44),
			"tags":                int64(45),
			"ssh_identities":      map[string]int64{"browser": 46},
			"provisioning_keys":   map[string]int64{"webhook": 47},
			"enrollments":         int64(48),
			"pending_invitations": int64(49),

			"authentication_local": true,
			"authentication_saml":  false,
			"users_active":         int64(2),
			"users_by_origin":      map[string]int64{"local": 4},
		},
	}
}

func instanceUsage() *models.InstanceUsage {
	return &models.InstanceUsage{
		Namespaces: 3,
		Users:      4,
		Devices:    models.DeviceUsage{Accepted: 10, Pending: 2, Rejected: 1},
		Sessions:   25,
		SessionKinds: models.SessionKindUsage{
			Interactive: 11, ShellWithoutPty: 12, Exec: 13, SCP: 14, SFTP: 15,
			PortForwarding: 16, AgentForwarding: 17, Web: 18, Recorded: 19, APIKey: 20,
		},
		SessionActivity: models.SessionActivityUsage{
			Unauthenticated: 51, Active: 52, Devices: 53, Events: 54,
			DurationSeconds: 55, DurationMedianSeconds: 56, DurationP95Seconds: 57,
		},
		NamespaceKinds: models.NamespaceKindUsage{
			Legacy: 21, Personal: 22, Team: 23, Recording: 24, RequireVerifiedDeviceKey: 25,
		},
		DeviceKinds: models.DeviceKindUsage{
			Online: 31, Ephemeral: 32, Owned: 33, CustomFields: 34, Tagged: 35, KeyVerified: 36,
			Platforms:     map[string]int64{"docker": 7, "connector": 2, "acme-appliance": 1},
			Versions:      map[string]int64{"v0.27.0": 5, "0.27.1-rc.1": 2, "acme-build-7": 2, "unknown": 1},
			Architectures: map[string]int64{"amd64": 9, "acme-cpu": 1},
			OSes:          map[string]int64{"ubuntu": 7, "unknown": 1, "acmecorp-prod": 1, "Ubuntu": 1},
		},
		Access: models.AccessUsage{
			AccessPolicies: 41, PublicKeys: 42, APIKeys: 43, InstanceAPIKeys: 44, Tags: 45,
			SSHIdentities:    map[string]int64{"browser": 46},
			ProvisioningKeys: map[string]int64{"webhook": 47},
			Enrollments:      48, PendingInvitations: 49,
		},
		Authentication: models.AuthenticationUsage{
			LocalEnabled: true, ActiveUsers: 2, UsersByOrigin: map[string]int64{"local": 4},
		},
	}
}

func TestService_TelemetryHeartbeat(t *testing.T) {
	override(t, &telemetryHeartbeatHooks, nil)
	override(t, &telemetryInstancePropertiesHooks, nil)

	ctx := context.Background()
	since := telemetryNow.Add(-24 * time.Hour)

	t.Run("sends the instance heartbeat with exact counts", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		storeMock.On("SystemGet", ctx).Return(telemetrySystem(true), nil).Once()
		storeMock.On("InstanceUsage", ctx, since).Return(instanceUsage(), nil).Once()

		sender := new(recordingSender)
		require.NoError(t, newTelemetryService(t, storeMock, sender, false).TelemetryHeartbeat()(ctx))

		assert.Equal(t, []telemetry.Event{expectedInstanceHeartbeat()}, sender.sent)
	})

	t.Run("sends nothing when the administrator turned telemetry off", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		storeMock.On("SystemGet", ctx).Return(telemetrySystem(false), nil).Once()

		sender := new(recordingSender)
		require.NoError(t, newTelemetryService(t, storeMock, sender, false).TelemetryHeartbeat()(ctx))

		assert.Empty(t, sender.sent)
	})

	t.Run("sends nothing when the environment turned telemetry off", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)

		sender := new(recordingSender)
		require.NoError(t, newTelemetryService(t, storeMock, sender, true).TelemetryHeartbeat()(ctx))

		assert.Empty(t, sender.sent)
	})

	t.Run("sends nothing before the instance has an ID", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		system := telemetrySystem(true)
		system.InstanceID = ""
		storeMock.On("SystemGet", ctx).Return(system, nil).Once()

		sender := new(recordingSender)
		require.NoError(t, newTelemetryService(t, storeMock, sender, false).TelemetryHeartbeat()(ctx))

		assert.Empty(t, sender.sent)
	})

	t.Run("appends the events a hook adds, attributed to the instance", func(t *testing.T) {
		override(t, &telemetryHeartbeatHooks, nil)

		OnTelemetryHeartbeat(func(_ context.Context, gotSince time.Time) ([]telemetry.Event, error) {
			assert.Equal(t, since, gotSince)

			return []telemetry.Event{{Name: "namespace heartbeat", Groups: map[string]string{"namespace": "hashed"}}}, nil
		})

		storeMock := storemock.NewMockStore(t)
		storeMock.On("SystemGet", ctx).Return(telemetrySystem(true), nil).Once()
		storeMock.On("InstanceUsage", ctx, since).Return(instanceUsage(), nil).Once()

		sender := new(recordingSender)
		require.NoError(t, newTelemetryService(t, storeMock, sender, false).TelemetryHeartbeat()(ctx))

		assert.Equal(t, []telemetry.Event{
			expectedInstanceHeartbeat(),
			{Name: "namespace heartbeat", DistinctID: telemetryInstanceID, Groups: map[string]string{"namespace": "hashed"}},
		}, sender.sent)
	})

	t.Run("adds the properties a hook adds, never replacing the core ones", func(t *testing.T) {
		override(t, &telemetryInstancePropertiesHooks, nil)

		OnTelemetryInstanceProperties(func(_ context.Context, gotSince time.Time) (map[string]any, error) {
			assert.Equal(t, since, gotSince)

			return map[string]any{"firewall_rules": int64(5), "users": int64(999)}, nil
		})

		storeMock := storemock.NewMockStore(t)
		storeMock.On("SystemGet", ctx).Return(telemetrySystem(true), nil).Once()
		storeMock.On("InstanceUsage", ctx, since).Return(instanceUsage(), nil).Once()

		sender := new(recordingSender)
		require.NoError(t, newTelemetryService(t, storeMock, sender, false).TelemetryHeartbeat()(ctx))

		expected := expectedInstanceHeartbeat()
		expected.Properties["firewall_rules"] = int64(5)
		assert.Equal(t, []telemetry.Event{expected}, sender.sent)
	})

	t.Run("a failing properties hook sends nothing", func(t *testing.T) {
		override(t, &telemetryInstancePropertiesHooks, nil)

		hookErr := errors.New("cloud store down")
		OnTelemetryInstanceProperties(func(context.Context, time.Time) (map[string]any, error) {
			return nil, hookErr
		})

		storeMock := storemock.NewMockStore(t)
		storeMock.On("SystemGet", ctx).Return(telemetrySystem(true), nil).Once()
		storeMock.On("InstanceUsage", ctx, since).Return(instanceUsage(), nil).Once()

		sender := new(recordingSender)
		require.ErrorIs(t, newTelemetryService(t, storeMock, sender, false).TelemetryHeartbeat()(ctx), hookErr)
		assert.Empty(t, sender.sent)
	})

	t.Run("a failing hook sends nothing and releases the day for the retry", func(t *testing.T) {
		override(t, &telemetryHeartbeatHooks, nil)

		hookErr := errors.New("store down")
		OnTelemetryHeartbeat(func(context.Context, time.Time) ([]telemetry.Event, error) {
			return nil, hookErr
		})

		storeMock := storemock.NewMockStore(t)
		storeMock.On("SystemGet", ctx).Return(telemetrySystem(true), nil)
		storeMock.On("InstanceUsage", ctx, since).Return(instanceUsage(), nil).Twice()

		shared := newMemoryCache()
		sender := new(recordingSender)
		require.ErrorIs(t, newTelemetryServiceWithCache(t, storeMock, sender, false, shared).TelemetryHeartbeat()(ctx), hookErr)
		assert.Empty(t, sender.sent)

		hookErr = nil
		require.NoError(t, newTelemetryServiceWithCache(t, storeMock, sender, false, shared).TelemetryHeartbeat()(ctx))
		assert.Equal(t, []telemetry.Event{expectedInstanceHeartbeat()}, sender.sent)
	})

	t.Run("a failing usage count sends nothing and releases the day for the retry", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		storeMock.On("SystemGet", ctx).Return(telemetrySystem(true), nil)
		storeMock.On("InstanceUsage", ctx, since).Return(nil, errors.New("database down")).Once()
		storeMock.On("InstanceUsage", ctx, since).Return(instanceUsage(), nil).Once()

		shared := newMemoryCache()
		sender := new(recordingSender)
		require.Error(t, newTelemetryServiceWithCache(t, storeMock, sender, false, shared).TelemetryHeartbeat()(ctx))
		assert.Empty(t, sender.sent)

		require.NoError(t, newTelemetryServiceWithCache(t, storeMock, sender, false, shared).TelemetryHeartbeat()(ctx))
		assert.Equal(t, []telemetry.Event{expectedInstanceHeartbeat()}, sender.sent)
	})

	t.Run("replicas sharing a cache send one heartbeat a day", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		storeMock.On("SystemGet", ctx).Return(telemetrySystem(true), nil)
		storeMock.On("InstanceUsage", ctx, since).Return(instanceUsage(), nil).Once()

		shared := newMemoryCache()
		sender := new(recordingSender)
		for range 3 {
			require.NoError(t, newTelemetryServiceWithCache(t, storeMock, sender, false, shared).TelemetryHeartbeat()(ctx))
		}

		assert.Equal(t, []telemetry.Event{expectedInstanceHeartbeat()}, sender.sent)
	})

	t.Run("a heartbeat that fails to send releases the day for the retry", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		storeMock.On("SystemGet", ctx).Return(telemetrySystem(true), nil)
		storeMock.On("InstanceUsage", ctx, since).Return(instanceUsage(), nil).Twice()

		shared := newMemoryCache()
		failing := &recordingSender{err: errors.New("queue full")}
		require.Error(t, newTelemetryServiceWithCache(t, storeMock, failing, false, shared).TelemetryHeartbeat()(ctx))

		retry := new(recordingSender)
		require.NoError(t, newTelemetryServiceWithCache(t, storeMock, retry, false, shared).TelemetryHeartbeat()(ctx))

		assert.Equal(t, []telemetry.Event{expectedInstanceHeartbeat()}, retry.sent)
	})
}

func TestService_TelemetryInstanceStarted(t *testing.T) {
	ctx := context.Background()

	t.Run("sends instance started when telemetry is on", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		storeMock.On("SystemGet", ctx).Return(telemetrySystem(true), nil).Once()

		hook := logtest.NewGlobal()
		t.Cleanup(hook.Reset)

		sender := new(recordingSender)
		require.NoError(t, newTelemetryService(t, storeMock, sender, false).TelemetryInstanceStarted(ctx))

		require.NotNil(t, hook.LastEntry())
		assert.Contains(t, hook.LastEntry().Message, "http://collector.test")

		assert.Equal(t, []telemetry.Event{{
			Name:       "instance started",
			DistinctID: telemetryInstanceID,
			Properties: map[string]any{
				"edition":  "enterprise",
				"version":  "v0.27.0",
				"os":       runtime.GOOS,
				"arch":     runtime.GOARCH,
				"database": "postgres",
			},
		}}, sender.sent)
	})

	t.Run("sends nothing when telemetry is off", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		storeMock.On("SystemGet", ctx).Return(telemetrySystem(false), nil).Once()

		sender := new(recordingSender)
		require.NoError(t, newTelemetryService(t, storeMock, sender, false).TelemetryInstanceStarted(ctx))

		assert.Empty(t, sender.sent)
	})

	t.Run("sends nothing when the environment turns it off", func(t *testing.T) {
		sender := new(recordingSender)
		require.NoError(t, newTelemetryService(t, storemock.NewMockStore(t), sender, true).TelemetryInstanceStarted(ctx))

		assert.Empty(t, sender.sent)
	})
}

func TestService_TelemetryStatus(t *testing.T) {
	override(t, &telemetryHeartbeatHooks, nil)

	ctx := context.Background()
	since := telemetryNow.Add(-24 * time.Hour)

	cases := []struct {
		description           string
		enabled               bool
		disabledByEnvironment bool
		noSender              bool
		wantLocked            bool
	}{
		{description: "telemetry on", enabled: true, wantLocked: false},
		{description: "turned off by the administrator", enabled: false, wantLocked: false},
		{description: "turned off by the environment", enabled: true, disabledByEnvironment: true, wantLocked: true},
		{description: "without a sender", enabled: true, noSender: true, wantLocked: true},
	}

	for _, tc := range cases {
		t.Run(tc.description+" still previews the next heartbeat", func(t *testing.T) {
			storeMock := storemock.NewMockStore(t)
			storeMock.On("SystemGet", ctx).Return(telemetrySystem(tc.enabled), nil).Once()
			storeMock.On("InstanceUsage", ctx, since).Return(instanceUsage(), nil).Once()

			var sender TelemetrySender = new(recordingSender)
			if tc.noSender {
				sender = nil
			}

			status, err := newTelemetryService(t, storeMock, sender, tc.disabledByEnvironment).TelemetryStatus(ctx)
			require.NoError(t, err)

			assert.Equal(t, &responses.TelemetryStatus{
				Enabled:               tc.enabled,
				DisabledByEnvironment: tc.wantLocked,
				Preview:               []telemetry.Preview{expectedInstanceHeartbeat().Preview()},
			}, status)
		})
	}
}

func TestOnTelemetryInstanceProperties(t *testing.T) {
	override(t, &telemetryInstancePropertiesHooks, nil)

	assert.PanicsWithValue(t, "services: OnTelemetryInstanceProperties called with nil hook", func() { OnTelemetryInstanceProperties(nil) })
}
