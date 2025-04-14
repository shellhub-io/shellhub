package mongo

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	query = append([]bson.M{
		{
			"$match": bson.M{
				"disconnected_at": nil,
				"last_seen":       bson.M{"$gt": primitive.NewDateTimeFromTime(time.Now().Add(-2 * time.Minute))},
				"status":          "accepted",
			},
		},
	}, query...)

	onlineDevices, err := AggregateCount(ctx, s.db.Collection("devices"), query)
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

	return &models.Stats{
		RegisteredDevices: registeredDevices,
		OnlineDevices:     onlineDevices,
		PendingDevices:    pendingDevices,
		RejectedDevices:   rejectedDevices,
		ActiveSessions:    activeSessions,
	}, nil
}
