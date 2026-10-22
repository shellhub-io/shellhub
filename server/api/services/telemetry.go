package services

import (
	"context"
	"regexp"
	"runtime"
	"slices"
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
	// error of an [OnTelemetryHeartbeat] or [OnTelemetryInstanceProperties] hook that fails.
	TelemetryStatus(ctx context.Context) (*responses.TelemetryStatus, error)

	// TelemetryInstanceStarted sends the instance started event when telemetry is on, and logs
	// whether it is. It returns the store's error when the system row cannot be read, or the
	// sender's when the event cannot be queued.
	TelemetryInstanceStarted(ctx context.Context) error

	// TelemetryHeartbeat sends the instance heartbeat, with the properties the
	// [OnTelemetryInstanceProperties] hooks add, and the events the [OnTelemetryHeartbeat] hooks
	// add, when telemetry is on. Server replicas sharing a cache send
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
	addUsageBreakdown(properties, usage)

	extra, err := fireTelemetryInstanceProperties(ctx, since)
	if err != nil {
		return nil, err
	}

	for name, value := range extra {
		if _, taken := properties[name]; !taken {
			properties[name] = value
		}
	}

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

func addUsageBreakdown(properties map[string]any, usage *models.InstanceUsage) {
	sessions := usage.SessionKinds
	properties["sessions_interactive"] = sessions.Interactive
	properties["sessions_shell_without_pty"] = sessions.ShellWithoutPty
	properties["sessions_exec"] = sessions.Exec
	properties["sessions_scp"] = sessions.SCP
	properties["sessions_sftp"] = sessions.SFTP
	properties["sessions_port_forwarding"] = sessions.PortForwarding
	properties["sessions_agent_forwarding"] = sessions.AgentForwarding
	properties["sessions_web"] = sessions.Web
	properties["sessions_recorded"] = sessions.Recorded
	properties["sessions_api_key"] = sessions.APIKey

	activity := usage.SessionActivity
	properties["sessions_unauthenticated"] = activity.Unauthenticated
	properties["sessions_active"] = activity.Active
	properties["sessions_devices"] = activity.Devices
	properties["sessions_events"] = activity.Events
	properties["sessions_duration_seconds"] = activity.DurationSeconds
	properties["sessions_duration_median_seconds"] = activity.DurationMedianSeconds
	properties["sessions_duration_p95_seconds"] = activity.DurationP95Seconds

	namespaces := usage.NamespaceKinds
	properties["namespaces_legacy"] = namespaces.Legacy
	properties["namespaces_personal"] = namespaces.Personal
	properties["namespaces_team"] = namespaces.Team
	properties["namespaces_recording"] = namespaces.Recording
	properties["namespaces_require_verified_device_key"] = namespaces.RequireVerifiedDeviceKey

	devices := usage.DeviceKinds
	properties["devices_online"] = devices.Online
	properties["devices_ephemeral"] = devices.Ephemeral
	properties["devices_owned"] = devices.Owned
	properties["devices_custom_fields"] = devices.CustomFields
	properties["devices_tagged"] = devices.Tagged
	properties["devices_key_verified"] = devices.KeyVerified
	properties["devices_platforms"] = bucketCounts(devices.Platforms, knownLabel(telemetryAgentPlatforms))
	properties["devices_versions"] = bucketCounts(devices.Versions, agentMinorVersion)
	properties["devices_architectures"] = bucketCounts(devices.Architectures, knownLabel(telemetryAgentArchitectures))
	properties["devices_oses"] = bucketCounts(devices.OSes, knownLabel(telemetryAgentOSes))
	properties["devices_enrolled_with"] = nonNilCounts(devices.EnrolledWith)

	access := usage.Access
	properties["access_policies"] = access.AccessPolicies
	properties["public_keys"] = access.PublicKeys
	properties["api_keys"] = access.APIKeys
	properties["instance_api_keys"] = access.InstanceAPIKeys
	properties["tags"] = access.Tags
	properties["ssh_identities"] = nonNilCounts(access.SSHIdentities)
	properties["provisioning_keys"] = nonNilCounts(access.ProvisioningKeys)
	properties["enrollments"] = access.Enrollments
	properties["pending_invitations"] = access.PendingInvitations

	auth := usage.Authentication
	properties["authentication_local"] = auth.LocalEnabled
	properties["authentication_saml"] = auth.SAMLEnabled
	properties["users_active"] = auth.ActiveUsers
	properties["users_by_origin"] = nonNilCounts(auth.UsersByOrigin)
}

const telemetryOtherLabel = "other"

var (
	telemetryAgentPlatforms     = []string{"docker", "bundle", "native", "connector"}
	telemetryAgentArchitectures = []string{
		"386", "amd64", "arm", "arm64", "loong64", "mips", "mips64", "mips64le", "mipsle",
		"ppc64", "ppc64le", "riscv64", "s390x",
	}
	telemetryAgentOSes = []string{
		"alpine", "arch", "centos", "coreos", "debian", "devuan", "elementary", "fedora", "freebsd",
		"gentoo", "linuxmint", "mageia", "manjaro", "mandriva", "nixos", "opensuse", "rhel",
		"sabayon", "slackware", "ubuntu", "ubuntu-core", "ubuntucore", "raspbian", "void", "docker",
	}
	agentReleaseVersion = regexp.MustCompile(`^v?(\d{1,4})\.(\d{1,4})\.\d{1,4}`)
)

func knownLabel(known []string) func(string) string {
	return func(value string) string {
		if value == models.UsageUnknownValue || slices.Contains(known, value) {
			return value
		}

		return telemetryOtherLabel
	}
}

func agentMinorVersion(version string) string {
	if version == models.UsageUnknownValue {
		return version
	}

	match := agentReleaseVersion.FindStringSubmatch(version)
	if match == nil {
		return telemetryOtherLabel
	}

	return "v" + match[1] + "." + match[2]
}

func bucketCounts(counts map[string]int64, label func(string) string) map[string]int64 {
	bucketed := make(map[string]int64, len(counts))
	for value, count := range counts {
		bucketed[label(value)] += count
	}

	return bucketed
}

func nonNilCounts(counts map[string]int64) map[string]int64 {
	if counts == nil {
		return map[string]int64{}
	}

	return counts
}
