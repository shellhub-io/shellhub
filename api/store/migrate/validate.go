package migrate

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// collectionTable maps MongoDB collection names to PostgreSQL table names.
// allowSkip indicates the migration filters orphaned records, so PG < Mongo is expected.
var collectionTable = []struct {
	collection string
	table      string
	allowSkip  bool
}{
	{"system", "systems", false},
	{"namespaces", "namespaces", false},
	{"users", "users", false},
	{"tags", "tags", true},
	{"devices", "devices", true},
	{"sessions", "sessions", true},
	{"sessions_events", "session_events", true},
	{"api_keys", "api_keys", true},
	{"public_keys", "public_keys", true},
}

func (m *Migrator) validateCounts(ctx context.Context) error {
	log.WithField("scope", "core").Info("Validating row counts")

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
			log.WithError(err).WithFields(log.Fields{"scope": "core", "table": ct.table}).Warn("Failed to update state counts")
		}

		l := log.WithFields(log.Fields{
			"scope":      "core",
			"collection": ct.collection,
			"table":      ct.table,
			"mongo":      mongoCount,
			"postgres":   pgCount,
		})

		if mongoCount != int64(pgCount) {
			if int64(pgCount) < mongoCount && ct.allowSkip {
				l.Warn("Count: PG < Mongo (orphaned records skipped)")
			} else {
				l.Error("Count mismatch")

				return fmt.Errorf("count mismatch for %s: mongo=%d postgres=%d", ct.table, mongoCount, pgCount)
			}
		} else {
			l.Info("Count matches")
		}
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

	log.WithField("scope", "core").Info("Count validation passed")

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

	l := log.WithFields(log.Fields{"scope": "core", "table": "memberships", "mongo": mongoCount, "postgres": pgCount})
	if mongoCount != int64(pgCount) {
		if int64(pgCount) < mongoCount {
			l.Warn("Count: PG < Mongo (orphaned records skipped)")
		} else {
			return fmt.Errorf("membership count mismatch: mongo=%d postgres=%d", mongoCount, pgCount)
		}
	} else {
		l.Info("Count matches")
	}

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

	l := log.WithFields(log.Fields{"scope": "core", "table": "device_tags", "mongo": mongoCount, "postgres": pgCount})
	if mongoCount != int64(pgCount) {
		if int64(pgCount) < mongoCount {
			l.Warn("Count: PG < Mongo (orphaned records skipped)")
		} else {
			return fmt.Errorf("device_tags count mismatch: mongo=%d postgres=%d", mongoCount, pgCount)
		}
	} else {
		l.Info("Count matches")
	}

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

	l := log.WithFields(log.Fields{"scope": "core", "table": "public_key_tags", "mongo": mongoCount, "postgres": pgCount})
	if mongoCount != int64(pgCount) {
		if int64(pgCount) < mongoCount {
			l.Warn("Count: PG < Mongo (orphaned records skipped)")
		} else {
			return fmt.Errorf("public_key_tags count mismatch: mongo=%d postgres=%d", mongoCount, pgCount)
		}
	} else {
		l.Info("Count matches")
	}

	return nil
}
