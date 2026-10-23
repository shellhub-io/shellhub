package services

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/shellhub-io/shellhub/pkg/telemetry"
)

// TelemetryHeartbeatHookFn returns extra events for the telemetry heartbeat, counting from since.
// The heartbeat sets each event's DistinctID to the instance ID, and sends the events only when
// telemetry is on. An error fails the whole heartbeat, so nothing is sent.
type TelemetryHeartbeatHookFn func(ctx context.Context, since time.Time) ([]telemetry.Event, error)

var telemetryHeartbeatHooks []TelemetryHeartbeatHookFn

// OnTelemetryHeartbeat registers a hook that adds events to the telemetry heartbeat and to its
// preview. It must be called during package init, before the server starts, and panics when fn is
// nil. Cloud uses it to report per-namespace usage.
func OnTelemetryHeartbeat(fn TelemetryHeartbeatHookFn) {
	if fn == nil {
		panic("services: OnTelemetryHeartbeat called with nil hook")
	}

	telemetryHeartbeatHooks = append(telemetryHeartbeatHooks, fn)
}

func fireTelemetryHeartbeat(ctx context.Context, since time.Time) ([]telemetry.Event, error) {
	var events []telemetry.Event

	for _, fn := range telemetryHeartbeatHooks {
		hooked, err := fn(ctx, since)
		if err != nil {
			return nil, fmt.Errorf("telemetry heartbeat hook failed: %w", err)
		}

		events = append(events, hooked...)
	}

	return events, nil
}

// TelemetryInstancePropertiesHookFn returns extra properties for the instance heartbeat, counting
// from since. A name the core heartbeat already sends is ignored, so a hook cannot change one. An
// error fails the whole heartbeat, so nothing is sent.
type TelemetryInstancePropertiesHookFn func(ctx context.Context, since time.Time) (map[string]any, error)

var telemetryInstancePropertiesHooks []TelemetryInstancePropertiesHookFn

// OnTelemetryInstanceProperties registers a hook that adds properties to the instance heartbeat
// and to its preview. It must be called during package init, before the server starts, and
// panics when fn is nil. Enterprise and Cloud use it to count the features only they store.
func OnTelemetryInstanceProperties(fn TelemetryInstancePropertiesHookFn) {
	if fn == nil {
		panic("services: OnTelemetryInstanceProperties called with nil hook")
	}

	telemetryInstancePropertiesHooks = append(telemetryInstancePropertiesHooks, fn)
}

func fireTelemetryInstanceProperties(ctx context.Context, since time.Time) (map[string]any, error) {
	properties := map[string]any{}

	for _, fn := range telemetryInstancePropertiesHooks {
		hooked, err := fn(ctx, since)
		if err != nil {
			return nil, fmt.Errorf("telemetry instance properties hook failed: %w", err)
		}

		maps.Copy(properties, hooked)
	}

	return properties, nil
}
