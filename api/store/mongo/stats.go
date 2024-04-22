package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) GetStats(ctx context.Context) (*models.Stats, error) {
	tenant := ""
	if t := gateway.TenantFromContext(ctx); t != nil {
		tenant = t.ID
	}

	query := []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant != "" {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant,
			},
		}}, query...)
	}
	query = append([]bson.M{{
		"$match": bson.M{
			"status": "accepted",
		},
	}}, query...)

	registeredDevices, err := AggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{"$count": "count"},
	}

	if tenant != "" {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant,
			},
		}}, query...)
	}

	query = append([]bson.M{{
		"$match": bson.M{
			"status": "pending",
		},
	}}, query...)

	pendingDevices, err := AggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{"$count": "count"},
	}

	if tenant != "" {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant,
			},
		}}, query...)
	}

	query = append([]bson.M{{
		"$match": bson.M{
			"status": "rejected",
		},
	}}, query...)

	rejectedDevices, err := AggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{}

	if tenant != "" {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant,
			},
		})
	}

	query = append(query, bson.M{
		"$count": "count",
	})

	activeSessions, err := AggregateCount(ctx, s.db.Collection("active_sessions"), query)
	if err != nil {
		return nil, err
	}

	var onlineDevices int64
	if tenant != "" {
		if onlineDevices, err = s.cache.CountConnectedDevices(ctx, tenant, models.DeviceStatusAccepted); err != nil {
			return nil, err
		}
	} else {
		// TODO:
		if onlineDevices, err = s.cache.CountConnectedDevices(ctx, tenant, models.DeviceStatus("*")); err != nil {
			return nil, err
		}
	}

	return &models.Stats{
		RegisteredDevices: registeredDevices,
		OnlineDevices:     int(onlineDevices), // TODO: convert the return type to int64
		PendingDevices:    pendingDevices,
		RejectedDevices:   rejectedDevices,
		ActiveSessions:    activeSessions,
	}, nil
}
