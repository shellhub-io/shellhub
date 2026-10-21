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
		},
	}
}

func instanceUsage() *models.InstanceUsage {
	return &models.InstanceUsage{
		Namespaces: 3,
		Users:      4,
		Devices:    models.DeviceUsage{Accepted: 10, Pending: 2, Rejected: 1},
		Sessions:   25,
	}
}

func TestService_TelemetryHeartbeat(t *testing.T) {
	override(t, &telemetryHeartbeatHooks, nil)

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
