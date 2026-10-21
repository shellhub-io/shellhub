package pg

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
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

	return usage, nil
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
