package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type usageFixture struct {
	since   time.Time
	tenantA string
	tenantB string
}

func (s *Suite) seedUsage(t *testing.T) usageFixture {
	t.Helper()

	require.NoError(t, s.provider.CleanDatabase(t))

	now := s.Clock().Now()

	tenantA := s.CreateNamespace(t)
	tenantB := s.CreateNamespace(t)
	s.CreateMembership(t, tenantB, s.CreateUser(t), "observer")

	accepted := s.CreateDevice(t, WithTenantID(tenantA), WithDeviceStatus(models.DeviceStatusAccepted))
	s.CreateDevice(t, WithTenantID(tenantA), WithDeviceStatus(models.DeviceStatusAccepted))
	s.CreateDevice(t, WithTenantID(tenantA), WithDeviceStatus(models.DeviceStatusPending))
	s.CreateDevice(t, WithTenantID(tenantB), WithDeviceStatus(models.DeviceStatusRejected))

	clk := s.pinClock(t, now.Add(-48*time.Hour))
	s.CreateSession(t, WithSessionDevice(accepted), WithSessionActive(false))
	clk.Set(now)
	s.CreateSession(t, WithSessionDevice(accepted), WithSessionActive(false))

	return usageFixture{since: now.Add(-24 * time.Hour), tenantA: tenantA, tenantB: tenantB}
}

// TestInstanceUsage locks the instance-wide counts the telemetry heartbeat reports.
func (s *Suite) TestInstanceUsage(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("counts the whole instance, sessions only since the cutoff", func(t *testing.T) {
		fixture := s.seedUsage(t)

		usage, err := st.InstanceUsage(ctx, fixture.since)
		require.NoError(t, err)

		assert.Equal(t, int64(2), usage.Namespaces)
		assert.Equal(t, int64(3), usage.Users)
		assert.Equal(t, models.DeviceUsage{Accepted: 2, Pending: 1, Rejected: 1}, usage.Devices)
		assert.Equal(t, int64(1), usage.Sessions)
	})

	t.Run("an empty instance counts zero, with local login on", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		usage, err := st.InstanceUsage(ctx, s.Clock().Now())
		require.NoError(t, err)

		assert.Equal(t, &models.InstanceUsage{
			DeviceKinds: models.DeviceKindUsage{
				Platforms:     map[string]int64{},
				Versions:      map[string]int64{},
				Architectures: map[string]int64{},
				OSes:          map[string]int64{},
				EnrolledWith:  map[string]int64{},
			},
			Access: models.AccessUsage{
				SSHIdentities:    map[string]int64{},
				ProvisioningKeys: map[string]int64{},
			},
			Authentication: models.AuthenticationUsage{LocalEnabled: true, UsersByOrigin: map[string]int64{}},
		}, usage)
	})

	t.Run("breaks the counts down by kind", func(t *testing.T) {
		provider, ok := s.provider.(dbAccessor)
		if !ok {
			t.Skip("provider does not expose the raw DB; the kinds cannot be seeded")
		}

		db := provider.DB()
		exec := func(query string, args ...any) {
			t.Helper()

			_, err := db.NewRaw(query, args...).Exec(ctx)
			require.NoError(t, err)
		}

		require.NoError(t, s.provider.CleanDatabase(t))

		now := s.Clock().Now()

		active := s.CreateUser(t)
		federated := s.CreateUser(t)
		invited := s.CreateUser(t)

		legacy := s.CreateNamespace(t, WithOwner(active))
		personal := s.CreateNamespace(t, WithOwner(federated))
		const ownerPoliciesSeededByNamespaceCreate = 2
		exec(`UPDATE namespaces SET legacy = true, scope = 'team', record_sessions = true WHERE id = ?`, legacy)
		exec(`UPDATE namespaces SET scope = 'personal', record_sessions = false, require_verified_device_key = true
			WHERE id = ?`, personal)

		exec(`UPDATE users SET last_login = ?`, now.Add(-48*time.Hour))
		exec(`UPDATE users SET last_login = ? WHERE id = ?`, now, active)
		exec(`UPDATE users SET origin = 'saml' WHERE id = ?`, federated)
		require.NoError(t, st.SystemSet(ctx, &models.System{
			Setup:          true,
			Authentication: &models.SystemAuthentication{Local: &models.SystemAuthenticationLocal{Enabled: false}},
		}))
		exec(`UPDATE systems SET authentication_saml_enabled = true`)

		digest := s.createProvisioningKey(t, legacy)

		docker := s.CreateDevice(t, WithTenantID(legacy), WithDeviceOwner(active))
		native := s.CreateDevice(t, WithTenantID(legacy), WithDeviceEphemeral(60), WithDeviceProvisioningKey(digest),
			WithDeviceLastSeen(now.Add(-time.Hour)))
		pending := s.CreateDevice(t, WithTenantID(legacy), WithDeviceStatus(models.DeviceStatusPending))
		exec(`UPDATE devices SET platform = 'docker', version = 'v0.27.0', arch = 'amd64', identifier = 'ubuntu',
			custom_fields = '{"rack": "a1"}' WHERE id = ?`, docker)
		exec(`UPDATE devices SET platform = 'native', version = '', arch = 'arm64', identifier = 'debian',
			key_verified_at = now() WHERE id = ?`, native)

		tag := s.CreateTag(t, WithTagTenant(legacy))
		exec(`INSERT INTO device_tags (device_id, tag_id, created_at) VALUES (?, ?, now())`, docker, tag)
		s.CreatePublicKey(t, WithPublicKeyTenant(legacy))
		s.CreateAPIKey(t, WithAPIKeyTenant(legacy), WithAPIKeyCreatedBy(active))
		s.CreateInstanceAPIKey(t, WithInstanceAPIKeyCreatedBy(active))

		_, err := st.AccessPolicyCreate(ctx, &models.AccessPolicy{
			TenantID: legacy,
			Name:     "ops",
			Subject:  models.PolicySubject{Type: models.PolicySubjectUser, Value: active},
			Filter:   models.AccessPolicyFilter{Taggable: models.Taggable{TagIDs: []string{tag}}},
			Logins:   []string{"root"},
			SourceIP: []string{},
			Action:   models.PolicyActionAllow,
		})
		require.NoError(t, err)

		authorized, fingerprint := newSSHIdentityKey(t)
		_, err = st.SSHIdentityCreate(ctx, &models.SSHIdentity{
			TenantID:    legacy,
			PrincipalID: active,
			Fingerprint: fingerprint,
			Data:        authorized,
			Name:        "laptop",
			Source:      models.SSHIdentitySourceManual,
		})
		require.NoError(t, err)

		expiresAt := now.Add(7 * 24 * time.Hour)
		require.NoError(t, st.MembershipInvitationCreate(ctx, &models.MembershipInvitation{
			TenantID:    legacy,
			UserID:      invited,
			InvitedBy:   active,
			Role:        authorizer.RoleObserver,
			Status:      models.MembershipInvitationStatusPending,
			ExpiresAt:   &expiresAt,
			Invitations: 1,
		}))

		enrollment := func(deviceUID string, refusal models.ProvisioningKeyEventRefusal) *models.ProvisioningKeyEvent {
			return &models.ProvisioningKeyEvent{
				ProvisioningKeyID: digest,
				TenantID:          legacy,
				DeviceUID:         deviceUID,
				Hostname:          "box",
				SourceIP:          "203.0.113.7",
				Refusal:           refusal,
			}
		}
		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, enrollment(string(native), "")))
		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, enrollment("refused", models.ProvisioningKeyEventRefusalUnverifiedDeviceKey)))
		require.NoError(t, st.ProvisioningKeyEventCreate(ctx, enrollment("refused", models.ProvisioningKeyEventRefusalUnverifiedDeviceKey)))

		events := func(session models.UID, seats ...[]models.SessionEvent) {
			for seat, seatEvents := range seats {
				for _, event := range seatEvents {
					event.Session = string(session)
					event.Seat = seat
					event.Timestamp = now
					require.NoError(t, st.SessionEventsCreate(ctx, &event))
				}
			}
		}

		pty := models.SessionEvent{Type: models.SessionEventTypePtyRequest, Data: map[string]any{"term": "xterm"}}
		shell := models.SessionEvent{Type: models.SessionEventTypeShell}
		command := func(eventType models.SessionEventType, line string) models.SessionEvent {
			return models.SessionEvent{Type: eventType, Data: &models.SSHCommand{Command: line}}
		}

		web := s.CreateSession(t, WithSessionDevice(docker), WithSessionActive(false))
		exec(`UPDATE sessions SET web = true, seen_at = started_at + interval '600 seconds' WHERE id = ?`, web)
		events(web,
			[]models.SessionEvent{pty, shell, {Type: models.SessionEventTypeAuthAgentRequest}, {Type: models.SessionEventTypePtyOutput}},
			[]models.SessionEvent{command(models.SessionEventTypeExec, "scp -t /tmp")},
			[]models.SessionEvent{{Type: models.SessionEventTypeDirectTCPIP, Data: map[string]any{"dest_port": 80}}},
		)

		heredoc := s.CreateSession(t, WithSessionDevice(docker), WithSessionActive(false))
		exec(`UPDATE sessions SET recorded = true, api_key_id = (SELECT id FROM api_keys),
			seen_at = started_at + interval '60 seconds' WHERE id = ?`, heredoc)
		events(heredoc,
			[]models.SessionEvent{shell},
			[]models.SessionEvent{pty, command(models.SessionEventTypeExec, "uptime")},
		)

		sftp := s.CreateSession(t, WithSessionDevice(native), WithSessionActive(false))
		exec(`UPDATE sessions SET seen_at = started_at + interval '20 seconds' WHERE id = ?`, sftp)
		events(sftp, []models.SessionEvent{command(models.SessionEventTypeSubsystem, "sftp")})

		nul := s.CreateSession(t, WithSessionDevice(pending), WithSessionActive(false))
		exec(`UPDATE sessions SET authenticated = false, seen_at = started_at + interval '5 seconds' WHERE id = ?`, nul)
		events(nul, []models.SessionEvent{command(models.SessionEventTypeExec, "ls\x00")})

		old := s.CreateSession(t, WithSessionDevice(docker), WithSessionActive(true))
		exec(`UPDATE sessions SET started_at = ? WHERE id = ?`, now.Add(-48*time.Hour), old)
		events(old, []models.SessionEvent{pty, shell})

		usage, err := st.InstanceUsage(ctx, now.Add(-24*time.Hour))
		require.NoError(t, err)

		assert.Equal(t, &models.InstanceUsage{
			Namespaces: 2,
			Users:      3,
			Devices:    models.DeviceUsage{Accepted: 2, Pending: 1},
			Sessions:   4,
			SessionKinds: models.SessionKindUsage{
				Interactive:     1,
				ShellWithoutPty: 1,
				Exec:            2,
				SCP:             1,
				SFTP:            1,
				PortForwarding:  1,
				AgentForwarding: 1,
				Web:             1,
				Recorded:        1,
				APIKey:          1,
			},
			SessionActivity: models.SessionActivityUsage{
				Unauthenticated:       1,
				Active:                1,
				Devices:               2,
				Events:                10,
				DurationSeconds:       600 + 60 + 20,
				DurationMedianSeconds: 60,
				DurationP95Seconds:    546,
			},
			NamespaceKinds: models.NamespaceKindUsage{
				Legacy:                   1,
				Personal:                 1,
				Team:                     1,
				Recording:                1,
				RequireVerifiedDeviceKey: 1,
			},
			DeviceKinds: models.DeviceKindUsage{
				Online:        1,
				Ephemeral:     1,
				Owned:         1,
				CustomFields:  1,
				Tagged:        1,
				KeyVerified:   1,
				Platforms:     map[string]int64{"docker": 1, "native": 1},
				Versions:      map[string]int64{"v0.27.0": 1, "unknown": 1},
				Architectures: map[string]int64{"amd64": 1, "arm64": 1},
				OSes:          map[string]int64{"ubuntu": 1, "debian": 1},
				EnrolledWith:  map[string]int64{"user": 1},
			},
			Access: models.AccessUsage{
				AccessPolicies:     1 + ownerPoliciesSeededByNamespaceCreate,
				PublicKeys:         1,
				APIKeys:            1,
				InstanceAPIKeys:    1,
				Tags:               1,
				SSHIdentities:      map[string]int64{"manual": 1},
				ProvisioningKeys:   map[string]int64{"automatic": 1},
				Enrollments:        3,
				PendingInvitations: 1,
			},
			Authentication: models.AuthenticationUsage{
				LocalEnabled:  false,
				SAMLEnabled:   true,
				ActiveUsers:   1,
				UsersByOrigin: map[string]int64{"local": 2, "saml": 1},
			},
		}, usage, "a pty on another seat does not make a shell interactive, a refused retry is an attempt, and an "+
			"unauthenticated session adds no duration")
	})
}

// TestNamespaceUsages locks the per-namespace counts the cloud telemetry heartbeat reports.
func (s *Suite) TestNamespaceUsages(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("counts each namespace on its own", func(t *testing.T) {
		fixture := s.seedUsage(t)

		usages, err := st.NamespaceUsages(ctx, fixture.since)
		require.NoError(t, err)

		assert.ElementsMatch(t, []models.NamespaceUsage{
			{
				TenantID: fixture.tenantA,
				Members:  1,
				Devices:  models.DeviceUsage{Accepted: 2, Pending: 1},
				Sessions: 1,
			},
			{
				TenantID: fixture.tenantB,
				Members:  2,
				Devices:  models.DeviceUsage{Rejected: 1},
			},
		}, usages)
	})
}
