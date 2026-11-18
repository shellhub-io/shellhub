package services

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/telemetry"
)

// DeviceMergeHookFn is called when two devices are merged. The hook receives
// the tenant ID and both the old and new device models. Hooks run inside the
// same transaction as mergeDevice, so a returned error will roll back the
// entire merge.
type DeviceMergeHookFn func(ctx context.Context, tenantID string, oldDevice, newDevice *models.Device) error

// DeviceRemovedHookFn is called after this process removes a device from its namespace, once the
// removal has committed. It cannot fail the removal, which has already happened. It must be safe to
// call more than once for the same device. A removal another process made only evicts the tunnel
// when the device next beats; it does not reach these hooks.
type DeviceRemovedHookFn func(ctx context.Context, tenantID, uid string)

// MembershipInvitedHookFn is called after a membership invitation is created or refreshed and
// the transaction has committed. It receives the fully-populated notification (invitation
// signature, expiry, recipient email + name), assembled by the intake flow: everything the email
// needs besides the public URL its link starts at. Cloud uses it to deliver the invitation email;
// the hook runs outside the DB transaction, so a delivery failure never rolls back the invite.
type MembershipInvitedHookFn func(ctx context.Context, notification *models.MembershipInvitationNotification) error

// NamespaceDeleteHookFn is called when a namespace is about to be deleted.
// The hook receives the namespace that will be deleted. Hooks run before the
// actual deletion, so a returned error will abort the entire operation.
type NamespaceDeleteHookFn func(ctx context.Context, namespace *models.Namespace) error

// TelemetryHeartbeatHookFn returns extra events for the telemetry heartbeat, counting from since.
// The heartbeat sets each event's DistinctID to the instance ID, and sends the events only when
// telemetry is on. An error fails the whole heartbeat, so nothing is sent.
type TelemetryHeartbeatHookFn func(ctx context.Context, since time.Time) ([]telemetry.Event, error)

// TelemetryInstancePropertiesHookFn returns extra properties for the instance heartbeat, counting
// from since. A name the core heartbeat already sends is ignored, so a hook cannot change one. An
// error fails the whole heartbeat, so nothing is sent.
type TelemetryInstancePropertiesHookFn func(ctx context.Context, since time.Time) (map[string]any, error)

// UserRegisteredHookFn is called after a not-confirmed user is created during registration. Cloud
// uses it to send the confirmation code, whose validity it decides. deviceCode is the canonical
// pairing code of the device the registration started from, or empty. It runs outside any
// transaction and its failure is non-fatal (registration succeeds; the user can request a resend).
type UserRegisteredHookFn func(ctx context.Context, user *models.User, deviceCode string) error

// Hooks are what an edition does when the service layer reaches one of its events. Each list runs
// in order; the first error stops the list and is returned to the event, as each hook type states.
type Hooks struct {
	DeviceMerge                 []DeviceMergeHookFn
	DeviceRemoved               []DeviceRemovedHookFn
	EmailCheck                  []EmailCheckHookFn
	EmailChangeRequested        []EmailChangeRequestedHookFn
	UserEmailsChanged           []UserEmailsChangedHookFn
	MFACodeCheck                []MFACodeCheckHookFn
	MembershipInvited           []MembershipInvitedHookFn
	NamespaceDelete             []NamespaceDeleteHookFn
	UserRegistered              []UserRegisteredHookFn
	TelemetryHeartbeat          []TelemetryHeartbeatHookFn
	TelemetryInstanceProperties []TelemetryInstancePropertiesHookFn
}

func (h *Hooks) add(other Hooks) {
	h.DeviceMerge = append(h.DeviceMerge, other.DeviceMerge...)
	h.DeviceRemoved = append(h.DeviceRemoved, other.DeviceRemoved...)
	h.EmailCheck = append(h.EmailCheck, other.EmailCheck...)
	h.EmailChangeRequested = append(h.EmailChangeRequested, other.EmailChangeRequested...)
	h.UserEmailsChanged = append(h.UserEmailsChanged, other.UserEmailsChanged...)
	h.MFACodeCheck = append(h.MFACodeCheck, other.MFACodeCheck...)
	h.MembershipInvited = append(h.MembershipInvited, other.MembershipInvited...)
	h.NamespaceDelete = append(h.NamespaceDelete, other.NamespaceDelete...)
	h.UserRegistered = append(h.UserRegistered, other.UserRegistered...)
	h.TelemetryHeartbeat = append(h.TelemetryHeartbeat, other.TelemetryHeartbeat...)
	h.TelemetryInstanceProperties = append(h.TelemetryInstanceProperties, other.TelemetryInstanceProperties...)
}

func (h *Hooks) deviceMerge(ctx context.Context, tenantID string, oldDevice, newDevice *models.Device) error {
	for _, fn := range h.DeviceMerge {
		if err := fn(ctx, tenantID, oldDevice, newDevice); err != nil {
			return fmt.Errorf("device merge hook failed: %w", err)
		}
	}

	return nil
}

func (h *Hooks) deviceRemoved(ctx context.Context, tenantID, uid string) {
	for _, fn := range h.DeviceRemoved {
		fn(ctx, tenantID, uid)
	}
}

func (h *Hooks) membershipInvited(ctx context.Context, notification *models.MembershipInvitationNotification) error {
	for _, fn := range h.MembershipInvited {
		if err := fn(ctx, notification); err != nil {
			return fmt.Errorf("membership invited hook failed: %w", err)
		}
	}

	return nil
}

func (h *Hooks) namespaceDelete(ctx context.Context, ns *models.Namespace) error {
	for _, fn := range h.NamespaceDelete {
		if err := fn(ctx, ns); err != nil {
			return fmt.Errorf("namespace delete hook failed: %w", err)
		}
	}

	return nil
}

func (h *Hooks) userRegistered(ctx context.Context, user *models.User, deviceCode string) error {
	for _, fn := range h.UserRegistered {
		if err := fn(ctx, user, deviceCode); err != nil {
			return fmt.Errorf("user registered hook failed: %w", err)
		}
	}

	return nil
}

func (h *Hooks) telemetryHeartbeat(ctx context.Context, since time.Time) ([]telemetry.Event, error) {
	var events []telemetry.Event

	for _, fn := range h.TelemetryHeartbeat {
		hooked, err := fn(ctx, since)
		if err != nil {
			return nil, fmt.Errorf("telemetry heartbeat hook failed: %w", err)
		}

		events = append(events, hooked...)
	}

	return events, nil
}

func (h *Hooks) telemetryInstanceProperties(ctx context.Context, since time.Time) (map[string]any, error) {
	properties := map[string]any{}

	for _, fn := range h.TelemetryInstanceProperties {
		hooked, err := fn(ctx, since)
		if err != nil {
			return nil, fmt.Errorf("telemetry instance properties hook failed: %w", err)
		}

		maps.Copy(properties, hooked)
	}

	return properties, nil
}
