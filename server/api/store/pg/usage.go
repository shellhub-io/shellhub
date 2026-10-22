package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
	"github.com/uptrace/bun"
)

// InstanceUsage implements [store.StatsStore].
func (pg *Pg) InstanceUsage(ctx context.Context, since time.Time) (*models.InstanceUsage, error) {
	db := pg.GetConnection(ctx)

	usage := new(models.InstanceUsage)
	if err := db.NewRaw(`
		SELECT
			(SELECT count(*) FROM namespaces),
			(SELECT count(*) FROM users),
			count(*) FILTER (WHERE devices.status = 'accepted'),
			count(*) FILTER (WHERE devices.status = 'pending'),
			count(*) FILTER (WHERE devices.status = 'rejected'),
			(SELECT count(*) FROM sessions WHERE started_at >= ?)
		FROM devices`, since).Scan(ctx,
		&usage.Namespaces,
		&usage.Users,
		&usage.Devices.Accepted,
		&usage.Devices.Pending,
		&usage.Devices.Rejected,
		&usage.Sessions,
	); err != nil {
		return nil, fromSQLError(err)
	}

	for _, count := range []func(context.Context, bun.IDB, time.Time, *models.InstanceUsage) error{
		pg.countSessionKinds,
		pg.countSessionActivity,
		pg.countNamespaceKinds,
		pg.countDeviceKinds,
		pg.countAccess,
		pg.countAuthentication,
	} {
		if err := count(ctx, db, since, usage); err != nil {
			return nil, fromSQLError(err)
		}
	}

	return usage, nil
}

func (pg *Pg) countSessionKinds(ctx context.Context, db bun.IDB, since time.Time, usage *models.InstanceUsage) error {
	kinds := &usage.SessionKinds

	if err := db.NewRaw(`
		SELECT
			count(*) FILTER (WHERE web),
			count(*) FILTER (WHERE recorded),
			count(*) FILTER (WHERE api_key_id IS NOT NULL)
		FROM sessions WHERE started_at >= ?`, since).Scan(ctx, &kinds.Web, &kinds.Recorded, &kinds.APIKey); err != nil {
		return err
	}

	scp, err := json.Marshal(models.SSHCommand{Command: "scp "})
	if err != nil {
		return err
	}

	sftp, err := json.Marshal(models.SSHCommand{Command: "sftp"})
	if err != nil {
		return err
	}

	return db.NewRaw(`
		WITH events AS (
			SELECT session_events.session_id, session_events.seat, session_events.type, session_events.data
			FROM session_events
			JOIN sessions ON sessions.id = session_events.session_id
			WHERE sessions.started_at >= ?0 AND session_events.type IN (?1)
		), seats AS (
			SELECT
				session_id,
				bool_or(type = ?2) AS pty,
				bool_or(type = ?3) AS shell,
				bool_or(type = ?4 AND data NOT LIKE ?7) AS exec,
				bool_or(type = ?4 AND data LIKE ?7) AS scp,
				bool_or(type = ?5 AND data = ?8) AS sftp,
				bool_or(type = ?6) AS port_forwarding,
				bool_or(type = ?9) AS agent_forwarding
			FROM events
			GROUP BY session_id, seat
		), kinds AS (
			SELECT
				session_id,
				bool_or(shell AND pty) AS interactive,
				bool_or(shell AND NOT pty) AS shell_without_pty,
				bool_or(exec) AS exec,
				bool_or(scp) AS scp,
				bool_or(sftp) AS sftp,
				bool_or(port_forwarding) AS port_forwarding,
				bool_or(agent_forwarding) AS agent_forwarding
			FROM seats
			GROUP BY session_id
		)
		SELECT
			count(*) FILTER (WHERE interactive),
			count(*) FILTER (WHERE shell_without_pty),
			count(*) FILTER (WHERE exec),
			count(*) FILTER (WHERE scp),
			count(*) FILTER (WHERE sftp),
			count(*) FILTER (WHERE port_forwarding),
			count(*) FILTER (WHERE agent_forwarding)
		FROM kinds`,
		since,
		bun.List(sessionKindEvents),
		models.SessionEventTypePtyRequest,
		models.SessionEventTypeShell,
		models.SessionEventTypeExec,
		models.SessionEventTypeSubsystem,
		models.SessionEventTypeDirectTCPIP,
		strings.TrimSuffix(string(scp), `"}`)+"%",
		string(sftp),
		models.SessionEventTypeAuthAgentRequest,
	).Scan(ctx,
		&kinds.Interactive,
		&kinds.ShellWithoutPty,
		&kinds.Exec,
		&kinds.SCP,
		&kinds.SFTP,
		&kinds.PortForwarding,
		&kinds.AgentForwarding,
	)
}

func (pg *Pg) countSessionActivity(ctx context.Context, db bun.IDB, since time.Time, usage *models.InstanceUsage) error {
	activity := &usage.SessionActivity

	return db.NewRaw(`
		WITH started AS (
			SELECT id, device_id, authenticated, greatest(extract(epoch FROM seen_at - started_at), 0) AS duration
			FROM sessions WHERE started_at >= ?0
		)
		SELECT
			count(*) FILTER (WHERE NOT authenticated),
			(SELECT count(*) FROM active_sessions),
			count(DISTINCT device_id) FILTER (WHERE authenticated),
			(SELECT count(*) FROM session_events JOIN started ON started.id = session_events.session_id),
			coalesce(sum(duration) FILTER (WHERE authenticated), 0)::bigint,
			coalesce(percentile_cont(0.5) WITHIN GROUP (ORDER BY duration) FILTER (WHERE authenticated), 0)::bigint,
			coalesce(percentile_cont(0.95) WITHIN GROUP (ORDER BY duration) FILTER (WHERE authenticated), 0)::bigint
		FROM started`, since).Scan(ctx,
		&activity.Unauthenticated,
		&activity.Active,
		&activity.Devices,
		&activity.Events,
		&activity.DurationSeconds,
		&activity.DurationMedianSeconds,
		&activity.DurationP95Seconds,
	)
}

var sessionKindEvents = []models.SessionEventType{
	models.SessionEventTypePtyRequest,
	models.SessionEventTypeShell,
	models.SessionEventTypeExec,
	models.SessionEventTypeSubsystem,
	models.SessionEventTypeDirectTCPIP,
	models.SessionEventTypeAuthAgentRequest,
}

func (pg *Pg) countNamespaceKinds(ctx context.Context, db bun.IDB, _ time.Time, usage *models.InstanceUsage) error {
	kinds := &usage.NamespaceKinds

	return db.NewRaw(`
		SELECT
			count(*) FILTER (WHERE legacy),
			count(*) FILTER (WHERE scope = ?),
			count(*) FILTER (WHERE scope = ?),
			count(*) FILTER (WHERE record_sessions),
			count(*) FILTER (WHERE require_verified_device_key)
		FROM namespaces`, models.TypePersonal, models.TypeTeam).Scan(ctx,
		&kinds.Legacy,
		&kinds.Personal,
		&kinds.Team,
		&kinds.Recording,
		&kinds.RequireVerifiedDeviceKey,
	)
}

func (pg *Pg) countDeviceKinds(ctx context.Context, db bun.IDB, _ time.Time, usage *models.InstanceUsage) error {
	kinds := &usage.DeviceKinds

	var err error
	if kinds.Online, err = buildOnlineDevicesQuery(db, OnlineCutoff(pg.clock)).Count(ctx); err != nil {
		return err
	}

	if err := db.NewRaw(`
		SELECT
			count(*) FILTER (WHERE ephemeral),
			count(*) FILTER (WHERE owner_id IS NOT NULL),
			count(*) FILTER (WHERE custom_fields IS NOT NULL AND custom_fields <> '{}'::jsonb),
			count(*) FILTER (WHERE EXISTS (SELECT 1 FROM device_tags WHERE device_tags.device_id = devices.id)),
			count(*) FILTER (WHERE key_verified_at IS NOT NULL)
		FROM devices WHERE status = 'accepted'`).Scan(ctx,
		&kinds.Ephemeral,
		&kinds.Owned,
		&kinds.CustomFields,
		&kinds.Tagged,
		&kinds.KeyVerified,
	); err != nil {
		return err
	}

	for column, counts := range map[string]*map[string]int64{
		"platform":   &kinds.Platforms,
		"version":    &kinds.Versions,
		"arch":       &kinds.Architectures,
		"identifier": &kinds.OSes,
	} {
		if *counts, err = countBy(ctx, db, `
			SELECT ? AS key, count(*) AS count FROM devices WHERE status = 'accepted' GROUP BY 1`, bun.Ident(column)); err != nil {
			return err
		}
	}

	kinds.EnrolledWith, err = countBy(ctx, db, `
		SELECT provisioning_keys.type AS key, count(*) AS count
		FROM devices
		JOIN provisioning_keys
			ON provisioning_keys.key_digest = devices.provisioning_key_id
			AND provisioning_keys.namespace_id = devices.namespace_id
		WHERE devices.status = 'accepted'
		GROUP BY 1`)

	return err
}

func (pg *Pg) countAccess(ctx context.Context, db bun.IDB, since time.Time, usage *models.InstanceUsage) error {
	access := &usage.Access

	if err := db.NewRaw(`
		SELECT
			(SELECT count(*) FROM access_policies),
			(SELECT count(*) FROM public_keys),
			(SELECT count(*) FROM api_keys),
			(SELECT count(*) FROM instance_api_keys),
			(SELECT count(*) FROM tags),
			(SELECT coalesce(sum(attempts), 0) FROM provisioning_key_events WHERE created_at >= ?),
			(SELECT count(*) FROM membership_invitations WHERE status = 'pending')`, since).Scan(ctx,
		&access.AccessPolicies,
		&access.PublicKeys,
		&access.APIKeys,
		&access.InstanceAPIKeys,
		&access.Tags,
		&access.Enrollments,
		&access.PendingInvitations,
	); err != nil {
		return err
	}

	var err error
	if access.SSHIdentities, err = countBy(ctx, db, `SELECT source AS key, count(*) AS count FROM ssh_identities GROUP BY 1`); err != nil {
		return err
	}

	access.ProvisioningKeys, err = countBy(ctx, db, `
		SELECT mode AS key, count(*) AS count FROM provisioning_keys WHERE type = ? AND NOT revoked GROUP BY 1`, models.ProvisioningKeyTypeUser)

	return err
}

func (pg *Pg) countAuthentication(ctx context.Context, db bun.IDB, since time.Time, usage *models.InstanceUsage) error {
	auth := &usage.Authentication

	system, err := pg.SystemGet(ctx)
	if err != nil {
		return err
	}

	auth.LocalEnabled = system.Authentication.Local.Enabled

	if err := selectSystem(db, new(entity.System)).
		ColumnExpr("authentication_saml_enabled").
		Scan(ctx, &auth.SAMLEnabled); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if err := db.NewRaw(`SELECT count(*) FROM users WHERE last_login >= ?`, since).Scan(ctx, &auth.ActiveUsers); err != nil {
		return err
	}

	auth.UsersByOrigin, err = countBy(ctx, db, `SELECT origin AS key, count(*) AS count FROM users GROUP BY 1`)

	return err
}

type keyedCount struct {
	Key   sql.NullString `bun:"key"`
	Count int64          `bun:"count"`
}

func countBy(ctx context.Context, db bun.IDB, query string, args ...any) (map[string]int64, error) {
	var rows []keyedCount
	if err := db.NewRaw(query, args...).Scan(ctx, &rows); err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		key := row.Key.String
		if key == "" {
			key = models.UsageUnknownValue
		}

		counts[key] += row.Count
	}

	return counts, nil
}

type namespaceUsageRow struct {
	TenantID        string `bun:"tenant_id"`
	Members         int64  `bun:"members"`
	AcceptedDevices int64  `bun:"accepted_devices"`
	PendingDevices  int64  `bun:"pending_devices"`
	RejectedDevices int64  `bun:"rejected_devices"`
	SessionsStarted int64  `bun:"sessions_started"`
}

// NamespaceUsages implements [store.StatsStore].
func (pg *Pg) NamespaceUsages(ctx context.Context, since time.Time) ([]models.NamespaceUsage, error) {
	db := pg.GetConnection(ctx)

	var rows []namespaceUsageRow
	if err := db.NewRaw(`
		SELECT
			namespaces.id AS tenant_id,
			(SELECT count(*) FROM memberships WHERE memberships.namespace_id = namespaces.id) AS members,
			count(devices.id) FILTER (WHERE devices.status = 'accepted') AS accepted_devices,
			count(devices.id) FILTER (WHERE devices.status = 'pending') AS pending_devices,
			count(devices.id) FILTER (WHERE devices.status = 'rejected') AS rejected_devices,
			(SELECT count(*) FROM sessions WHERE sessions.namespace_id = namespaces.id AND sessions.started_at >= ?) AS sessions_started
		FROM namespaces
		LEFT JOIN devices ON devices.namespace_id = namespaces.id
		GROUP BY namespaces.id`, since).Scan(ctx, &rows); err != nil {
		return nil, fromSQLError(err)
	}

	usages := make([]models.NamespaceUsage, 0, len(rows))
	for _, row := range rows {
		usages = append(usages, models.NamespaceUsage{
			TenantID: row.TenantID,
			Members:  row.Members,
			Devices: models.DeviceUsage{
				Accepted: row.AcceptedDevices,
				Pending:  row.PendingDevices,
				Rejected: row.RejectedDevices,
			},
			Sessions: row.SessionsStarted,
		})
	}

	return usages, nil
}
