package services

import (
	"context"
	"runtime"
	"time"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/telemetry"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/shellhub-io/shellhub/server/api/pkg/responses"
	log "github.com/sirupsen/logrus"
)

// CronTelemetryHeartbeat is when the instance sends its daily telemetry heartbeat.
const CronTelemetryHeartbeat = worker.CronSpec("30 5 * * *")

const telemetryHeartbeatPeriod = 24 * time.Hour

const telemetryHeartbeatClaimPrefix = "telemetry:heartbeat:"

// TelemetrySender delivers telemetry events to the collector at Endpoint. [telemetry.Client]
// implements it.
type TelemetrySender interface {
	Send(events ...telemetry.Event) error
	Endpoint() string
}

// TelemetryService reports the instance's usage to the ShellHub telemetry collector.
type TelemetryService interface {
	// TelemetryStatus returns the administrator's setting, whether the environment keeps
	// telemetry off (SHELLHUB_TELEMETRY is false, or the server has no client to send with), and
	// the events the next heartbeat would send, built whether or not it is on. It returns the store's error when the system row or the counts cannot be read, and the
	// error of an [OnTelemetryHeartbeat] hook that fails.
	TelemetryStatus(ctx context.Context) (*responses.TelemetryStatus, error)

	// TelemetryInstanceStarted sends the instance started event when telemetry is on, and logs
	// whether it is. It returns the store's error when the system row cannot be read, or the
	// sender's when the event cannot be queued.
	TelemetryInstanceStarted(ctx context.Context) error

	// TelemetryHeartbeat sends the instance heartbeat, and the events the
	// [OnTelemetryHeartbeat] hooks add, when telemetry is on. Server replicas sharing a cache send
	// it once per UTC day: the first to claim the day sends, the others return nil. Its handler
	// returns the store's error, the cache's error when the day cannot be claimed, a failing
	// hook's error, or the sender's error when the events cannot be queued; it sends nothing in
	// the first three cases, and releases the claim on a store, hook or sender error so a retry
	// can send.
	TelemetryHeartbeat() worker.CronHandler
}

func (s *service) TelemetryStatus(ctx context.Context) (*responses.TelemetryStatus, error) {
	system, err := s.systemGet(ctx)
	if err != nil {
		return nil, err
	}

	events, err := s.telemetryHeartbeatEvents(ctx, system)
	if err != nil {
		return nil, err
	}

	preview := make([]telemetry.Preview, 0, len(events))
	for _, event := range events {
		preview = append(preview, event.Preview())
	}

	return &responses.TelemetryStatus{
		Enabled:               system.Telemetry != nil && system.Telemetry.Enabled,
		DisabledByEnvironment: s.telemetryDisabledByEnvironment || s.telemetry == nil,
		Preview:               preview,
	}, nil
}

func (s *service) TelemetryInstanceStarted(ctx context.Context) error {
	if s.telemetryDisabledByEnvironment {
		log.Info("telemetry is off: SHELLHUB_TELEMETRY is false")

		return nil
	}

	system, err := s.systemGet(ctx)
	if err != nil {
		return err
	}

	if !s.telemetryEnabled(system) {
		log.Info("telemetry is off")

		return nil
	}

	log.Info("telemetry is on: the instance sends anonymous usage counts to " + s.telemetry.Endpoint() +
		". Set SHELLHUB_TELEMETRY=false to turn it off. See https://docs.shellhub.io/selfhosted/telemetry")

	return s.telemetry.Send(telemetry.Event{
		Name:       "instance started",
		DistinctID: system.InstanceID,
		Properties: s.telemetryInstanceProperties(),
	})
}

func (s *service) TelemetryHeartbeat() worker.CronHandler {
	return func(ctx context.Context) error {
		if s.telemetryDisabledByEnvironment {
			return nil
		}

		system, err := s.systemGet(ctx)
		if err != nil {
			return err
		}

		if !s.telemetryEnabled(system) {
			return nil
		}

		claimKey := telemetryHeartbeatClaimPrefix + s.clock.Now().UTC().Format(time.DateOnly)
		claim := uuid.Generate()

		claimed, err := s.cache.SetNX(ctx, claimKey, claim, telemetryHeartbeatPeriod+time.Hour)
		if err != nil || !claimed {
			return err
		}

		if err := s.sendTelemetryHeartbeat(ctx, system); err != nil {
			_, _ = s.cache.CompareAndDelete(context.WithoutCancel(ctx), claimKey, claim)

			return err
		}

		return nil
	}
}

func (s *service) sendTelemetryHeartbeat(ctx context.Context, system *models.System) error {
	events, err := s.telemetryHeartbeatEvents(ctx, system)
	if err != nil {
		return err
	}

	return s.telemetry.Send(events...)
}

func (s *service) telemetryEnabled(system *models.System) bool {
	return s.telemetry != nil &&
		system.InstanceID != "" &&
		system.Telemetry != nil && system.Telemetry.Enabled
}

func (s *service) telemetryInstanceProperties() map[string]any {
	edition := s.edition
	if edition == "" {
		edition = envs.Community
	}

	return map[string]any{
		"edition":  string(edition),
		"version":  s.version,
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
		"database": "postgres",
	}
}

func (s *service) telemetryHeartbeatEvents(ctx context.Context, system *models.System) ([]telemetry.Event, error) {
	since := s.clock.Now().Add(-telemetryHeartbeatPeriod)

	usage, err := s.store.InstanceUsage(ctx, since)
	if err != nil {
		return nil, err
	}

	properties := s.telemetryInstanceProperties()
	properties["namespaces"] = usage.Namespaces
	properties["users"] = usage.Users
	properties["devices_accepted"] = usage.Devices.Accepted
	properties["devices_pending"] = usage.Devices.Pending
	properties["devices_rejected"] = usage.Devices.Rejected
	properties["sessions"] = usage.Sessions

	hooked, err := fireTelemetryHeartbeat(ctx, since)
	if err != nil {
		return nil, err
	}

	events := make([]telemetry.Event, 0, len(hooked)+1)
	events = append(events, telemetry.Event{Name: "instance heartbeat", DistinctID: system.InstanceID, Properties: properties})

	for _, event := range hooked {
		event.DistinctID = system.InstanceID
		events = append(events, event)
	}

	return events, nil
}
