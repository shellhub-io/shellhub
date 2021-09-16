package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Store) GetStats(ctx context.Context) (*models.Stats, error) {
	query := []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		}}, query...)
	}

	query = append([]bson.M{{
		"$match": bson.M{
			"status": "accepted",
			"online": true,
		},
	}}, query...)

	onlineDevices, err := aggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
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

	registeredDevices, err := aggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
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

	pendingDevices, err := aggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
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

	rejectedDevices, err := aggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{
			"$lookup": bson.M{
				"from":         "active_sessions",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "active",
			},
		},
		{
			"$addFields": bson.M{
				"active": bson.M{"$anyElementTrue": []interface{}{"$active"}},
			},
		},
		{
			"$match": bson.M{
				"active": true,
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	query = append(query, bson.M{
		"$count": "count",
	})

	activeSessions, err := aggregateCount(ctx, s.db.Collection("sessions"), query)
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
