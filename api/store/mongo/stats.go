package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) GetStats(ctx context.Context) (*models.Stats, error) {
	query := []bson.M{
		{"$group": bson.M{"_id": bson.M{"uid": "$uid"}, "count": bson.M{"$sum": 1}}},
		{"$group": bson.M{"_id": bson.M{"uid": "$uid"}, "count": bson.M{"$sum": 1}}},
	}

	// Only match for the respective tenant if requested
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		}}, query...)
	}

	query = append([]bson.M{{
		"$match": bson.M{
			"status": "accepted",
		},
	}}, query...)

	query = []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
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

	// Only match for the respective tenant if requested
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
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

	// Only match for the respective tenant if requested
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
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

	// Only match for the respective tenant if requested
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
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

	onlineDevices := 0 // TODO: new online implementation

	return &models.Stats{
		RegisteredDevices: registeredDevices,
		OnlineDevices:     onlineDevices,
		PendingDevices:    pendingDevices,
		RejectedDevices:   rejectedDevices,
		ActiveSessions:    activeSessions,
	}, nil
}
