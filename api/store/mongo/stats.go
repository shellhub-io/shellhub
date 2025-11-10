package mongo

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Store) GetStats(ctx context.Context) (*models.Stats, error) {
	var tenantID string
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		tenantID = tenant.ID
	}

	onlineDevicesQuery := buildOnlineDevicesQuery(tenantID)
	onlineDevices, err := CountAllMatchingDocuments(ctx, s.db.Collection("devices"), onlineDevicesQuery)
	if err != nil {
		return nil, err
	}

	registeredDevicesQuery := buildRegisteredDevicesQuery(tenantID)
	registeredDevices, err := CountAllMatchingDocuments(ctx, s.db.Collection("devices"), registeredDevicesQuery)
	if err != nil {
		return nil, err
	}

	pendingDevicesQuery := buildPendingDevicesQuery(tenantID)
	pendingDevices, err := CountAllMatchingDocuments(ctx, s.db.Collection("devices"), pendingDevicesQuery)
	if err != nil {
		return nil, err
	}

	rejectedDevicesQuery := buildRejectedDevicesQuery(tenantID)
	rejectedDevices, err := CountAllMatchingDocuments(ctx, s.db.Collection("devices"), rejectedDevicesQuery)
	if err != nil {
		return nil, err
	}

	activeSessionsQuery := buildActiveSessionsQuery(tenantID)
	activeSessions, err := CountAllMatchingDocuments(ctx, s.db.Collection("active_sessions"), activeSessionsQuery)
	if err != nil {
		return nil, err
	}

	stats := &models.Stats{
		RegisteredDevices: registeredDevices,
		OnlineDevices:     onlineDevices,
		PendingDevices:    pendingDevices,
		RejectedDevices:   rejectedDevices,
		ActiveSessions:    activeSessions,
	}

	return stats, nil
}

func buildOnlineDevicesQuery(tenantID string) []bson.M {
	match := bson.M{
		"disconnected_at": nil,
		"last_seen":       bson.M{"$gt": primitive.NewDateTimeFromTime(time.Now().Add(-2 * time.Minute))},
		"status":          models.DeviceStatusAccepted,
	}

	if tenantID != "" {
		match["tenant_id"] = tenantID
	}

	return []bson.M{{"$match": match}}
}

func buildRegisteredDevicesQuery(tenantID string) []bson.M {
	match := bson.M{"status": models.DeviceStatusAccepted}
	if tenantID != "" {
		match["tenant_id"] = tenantID
	}

	return []bson.M{{"$match": match}}
}

func buildPendingDevicesQuery(tenantID string) []bson.M {
	match := bson.M{"status": models.DeviceStatusPending}
	if tenantID != "" {
		match["tenant_id"] = tenantID
	}

	return []bson.M{{"$match": match}}
}

func buildRejectedDevicesQuery(tenantID string) []bson.M {
	match := bson.M{"status": models.DeviceStatusRejected}
	if tenantID != "" {
		match["tenant_id"] = tenantID
	}

	return []bson.M{{"$match": match}}
}

func buildActiveSessionsQuery(tenantID string) []bson.M {
	match := bson.M{}
	if tenantID != "" {
		match["tenant_id"] = tenantID
	}

	return []bson.M{{"$match": match}}
}
