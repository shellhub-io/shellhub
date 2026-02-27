package migrate

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// collectionTable maps MongoDB collection names to PostgreSQL table names.
var collectionTable = []struct {
	collection string
	table      string
}{
	{"system", "systems"},
	{"namespaces", "namespaces"},
	{"users", "users"},
	{"tags", "tags"},
	{"devices", "devices"},
	{"sessions", "sessions"},
	{"sessions_events", "session_events"},
	{"api_keys", "api_keys"},
	{"public_keys", "public_keys"},
}

func (m *Migrator) validate(ctx context.Context) error {
	log.Info("Validating row counts")

	for _, ct := range collectionTable {
		mongoCount, err := m.mongo.Collection(ct.collection).CountDocuments(ctx, bson.M{})
		if err != nil {
			return fmt.Errorf("failed to count %s in MongoDB: %w", ct.collection, err)
		}

		var pgCount int
		if err := m.pg.NewSelect().TableExpr(ct.table).ColumnExpr("count(*)").Scan(ctx, &pgCount); err != nil {
			return fmt.Errorf("failed to count %s in PostgreSQL: %w", ct.table, err)
		}

		if err := setStateCounts(ctx, m.pg, ct.table, mongoCount, int64(pgCount)); err != nil {
			log.WithError(err).WithField("table", ct.table).Warn("Failed to update state counts")
		}

		l := log.WithFields(log.Fields{
			"collection": ct.collection,
			"table":      ct.table,
			"mongo":      mongoCount,
			"postgres":   pgCount,
		})

		if mongoCount != int64(pgCount) {
			l.Error("Row count mismatch")

			return fmt.Errorf("count mismatch for %s: mongo=%d postgres=%d", ct.table, mongoCount, pgCount)
		}

		l.Info("Row count matches")
	}

	// Validate relationship counts.
	if err := m.validateMemberships(ctx); err != nil {
		return err
	}
	if err := m.validateDeviceTags(ctx); err != nil {
		return err
	}
	if err := m.validatePublicKeyTags(ctx); err != nil {
		return err
	}

	// Spot-check samples.
	if err := m.spotCheckUsers(ctx); err != nil {
		return err
	}
	if err := m.spotCheckNamespaces(ctx); err != nil {
		return err
	}
	if err := m.spotCheckDevices(ctx); err != nil {
		return err
	}

	log.Info("All validations passed")

	return nil
}

func (m *Migrator) validateMemberships(ctx context.Context) error {
	// Count total members across all namespaces in MongoDB.
	pipeline := []bson.M{
		{"$project": bson.M{"count": bson.M{"$size": bson.M{"$ifNull": bson.A{"$members", bson.A{}}}}}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$count"}}},
	}

	cursor, err := m.mongo.Collection("namespaces").Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("failed to count memberships in MongoDB: %w", err)
	}
	defer cursor.Close(ctx)

	var mongoCount int64
	if cursor.Next(ctx) {
		var result struct {
			Total int64 `bson:"total"`
		}
		if err := cursor.Decode(&result); err != nil {
			return err
		}
		mongoCount = result.Total
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	var pgCount int
	if err := m.pg.NewSelect().TableExpr("memberships").ColumnExpr("count(*)").Scan(ctx, &pgCount); err != nil {
		return fmt.Errorf("failed to count memberships in PostgreSQL: %w", err)
	}

	if mongoCount != int64(pgCount) {
		return fmt.Errorf("membership count mismatch: mongo=%d postgres=%d", mongoCount, pgCount)
	}

	log.WithFields(log.Fields{"mongo": mongoCount, "postgres": pgCount}).Info("Membership count matches")

	return nil
}

func (m *Migrator) validateDeviceTags(ctx context.Context) error {
	pipeline := []bson.M{
		{"$project": bson.M{"count": bson.M{"$size": bson.M{"$ifNull": bson.A{"$tag_ids", bson.A{}}}}}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$count"}}},
	}

	cursor, err := m.mongo.Collection("devices").Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("failed to count device tags in MongoDB: %w", err)
	}
	defer cursor.Close(ctx)

	var mongoCount int64
	if cursor.Next(ctx) {
		var result struct {
			Total int64 `bson:"total"`
		}
		if err := cursor.Decode(&result); err != nil {
			return err
		}
		mongoCount = result.Total
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	var pgCount int
	if err := m.pg.NewSelect().TableExpr("device_tags").ColumnExpr("count(*)").Scan(ctx, &pgCount); err != nil {
		return fmt.Errorf("failed to count device_tags in PostgreSQL: %w", err)
	}

	if mongoCount != int64(pgCount) {
		return fmt.Errorf("device_tags count mismatch: mongo=%d postgres=%d", mongoCount, pgCount)
	}

	log.WithFields(log.Fields{"mongo": mongoCount, "postgres": pgCount}).Info("Device tags count matches")

	return nil
}

func (m *Migrator) validatePublicKeyTags(ctx context.Context) error {
	pipeline := []bson.M{
		{"$project": bson.M{"count": bson.M{"$size": bson.M{"$ifNull": bson.A{"$filter.tag_ids", bson.A{}}}}}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$count"}}},
	}

	cursor, err := m.mongo.Collection("public_keys").Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("failed to count public key tags in MongoDB: %w", err)
	}
	defer cursor.Close(ctx)

	var mongoCount int64
	if cursor.Next(ctx) {
		var result struct {
			Total int64 `bson:"total"`
		}
		if err := cursor.Decode(&result); err != nil {
			return err
		}
		mongoCount = result.Total
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	var pgCount int
	if err := m.pg.NewSelect().TableExpr("public_key_tags").ColumnExpr("count(*)").Scan(ctx, &pgCount); err != nil {
		return fmt.Errorf("failed to count public_key_tags in PostgreSQL: %w", err)
	}

	if mongoCount != int64(pgCount) {
		return fmt.Errorf("public_key_tags count mismatch: mongo=%d postgres=%d", mongoCount, pgCount)
	}

	log.WithFields(log.Fields{"mongo": mongoCount, "postgres": pgCount}).Info("Public key tags count matches")

	return nil
}

func (m *Migrator) spotCheckUsers(ctx context.Context) error {
	cursor, err := m.mongo.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	checked := 0
	for cursor.Next(ctx) && checked < 10 {
		var doc struct {
			ID       primitive.ObjectID `bson:"_id"`
			Email    string             `bson:"email"`
			Username string             `bson:"username"`
			Password string             `bson:"password"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		pgID := ObjectIDToUUID(doc.ID.Hex())

		var pgUser struct {
			Email          string `bun:"email"`
			Username       string `bun:"username"`
			PasswordDigest string `bun:"password_digest"`
		}
		err := m.pg.NewSelect().
			TableExpr("users").
			Column("email", "username", "password_digest").
			Where("id = ?", pgID).
			Scan(ctx, &pgUser)
		if err != nil {
			return fmt.Errorf("spot-check user %s not found in PG: %w", pgID, err)
		}

		if pgUser.Email != doc.Email {
			return fmt.Errorf("spot-check user %s email mismatch: %q vs %q", pgID, doc.Email, pgUser.Email)
		}
		if pgUser.Username != doc.Username {
			return fmt.Errorf("spot-check user %s username mismatch: %q vs %q", pgID, doc.Username, pgUser.Username)
		}
		if pgUser.PasswordDigest != doc.Password {
			return fmt.Errorf("spot-check user %s password_digest mismatch", pgID)
		}

		checked++
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	log.WithField("checked", checked).Info("User spot-check passed")

	return nil
}

func (m *Migrator) spotCheckNamespaces(ctx context.Context) error {
	cursor, err := m.mongo.Collection("namespaces").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	checked := 0
	for cursor.Next(ctx) && checked < 10 {
		var doc struct {
			TenantID string `bson:"tenant_id"`
			Name     string `bson:"name"`
			Owner    string `bson:"owner"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		var pgNS struct {
			Name    string `bun:"name"`
			OwnerID string `bun:"owner_id"`
		}
		err := m.pg.NewSelect().
			TableExpr("namespaces").
			Column("name", "owner_id").
			Where("id = ?", doc.TenantID).
			Scan(ctx, &pgNS)
		if err != nil {
			return fmt.Errorf("spot-check namespace %s not found in PG: %w", doc.TenantID, err)
		}

		if pgNS.Name != doc.Name {
			return fmt.Errorf("spot-check namespace %s name mismatch: %q vs %q", doc.TenantID, doc.Name, pgNS.Name)
		}
		if pgNS.OwnerID != ObjectIDToUUID(doc.Owner) {
			return fmt.Errorf("spot-check namespace %s owner_id mismatch: %q vs %q", doc.TenantID, ObjectIDToUUID(doc.Owner), pgNS.OwnerID)
		}

		checked++
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	log.WithField("checked", checked).Info("Namespace spot-check passed")

	return nil
}

func (m *Migrator) spotCheckDevices(ctx context.Context) error {
	cursor, err := m.mongo.Collection("devices").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	checked := 0
	for cursor.Next(ctx) && checked < 10 {
		var doc struct {
			UID      string         `bson:"uid"`
			Name     string         `bson:"name"`
			Identity *mongoDeviceID `bson:"identity"`
			Status   string         `bson:"status"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		var pgDev struct {
			Name   string `bun:"name"`
			MAC    string `bun:"mac"`
			Status string `bun:"status"`
		}
		err := m.pg.NewSelect().
			TableExpr("devices").
			Column("name", "mac", "status").
			Where("id = ?", doc.UID).
			Scan(ctx, &pgDev)
		if err != nil {
			return fmt.Errorf("spot-check device %s not found in PG: %w", doc.UID, err)
		}

		if pgDev.Name != doc.Name {
			return fmt.Errorf("spot-check device %s name mismatch: %q vs %q", doc.UID, doc.Name, pgDev.Name)
		}

		mac := ""
		if doc.Identity != nil {
			mac = doc.Identity.MAC
		}
		if pgDev.MAC != mac {
			return fmt.Errorf("spot-check device %s mac mismatch: %q vs %q", doc.UID, mac, pgDev.MAC)
		}

		status := doc.Status
		if status == "" {
			status = "pending"
		}
		if pgDev.Status != status {
			return fmt.Errorf("spot-check device %s status mismatch: %q vs %q", doc.UID, status, pgDev.Status)
		}

		checked++
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	log.WithField("checked", checked).Info("Device spot-check passed")

	return nil
}
