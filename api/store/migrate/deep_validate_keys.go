package migrate

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/uptrace/bun"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *Migrator) deepValidatePublicKeys(ctx context.Context, r *ValidationReport) error {
	cursor, err := m.mongo.Collection("public_keys").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	batch := make([]mongoPublicKey, 0, batchSize)

	for cursor.Next(ctx) {
		var doc mongoPublicKey
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		batch = append(batch, doc)
		if len(batch) >= batchSize {
			if err := m.comparePublicKeyBatch(ctx, r, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		return m.comparePublicKeyBatch(ctx, r, batch)
	}

	return nil
}

func (m *Migrator) comparePublicKeyBatch(ctx context.Context, r *ValidationReport, batch []mongoPublicKey) error {
	// Use composite key fingerprint:namespace_id since fingerprint alone is not unique
	type pkKey struct {
		Fingerprint string
		NamespaceID string
	}

	keys := make([]pkKey, len(batch))
	expected := make(map[pkKey]*entity.PublicKey, len(batch))
	fps := make([]string, len(batch))
	for i, doc := range batch {
		e := convertPublicKey(doc)
		k := pkKey{Fingerprint: e.Fingerprint, NamespaceID: e.NamespaceID}
		keys[i] = k
		fps[i] = e.Fingerprint
		expected[k] = e
	}

	var actual []entity.PublicKey
	if err := m.pg.NewSelect().Model(&actual).Where("fingerprint IN (?)", bun.List(fps)).Scan(ctx); err != nil {
		return err
	}

	pgMap := make(map[pkKey]*entity.PublicKey, len(actual))
	for i := range actual {
		pgMap[pkKey{Fingerprint: actual[i].Fingerprint, NamespaceID: actual[i].NamespaceID}] = &actual[i]
	}

	r.AddCompared("public_keys", int64(len(batch)))

	for _, k := range keys {
		exp := expected[k]
		recordID := k.Fingerprint + ":" + k.NamespaceID
		act, ok := pgMap[k]
		if !ok {
			r.AddMissing("public_keys", recordID)

			continue
		}

		t := "public_keys"
		r.CheckField(t, recordID, "NamespaceID", exp.NamespaceID, act.NamespaceID)
		r.CheckTime(t, recordID, "CreatedAt", exp.CreatedAt, act.CreatedAt)
		r.CheckField(t, recordID, "Name", exp.Name, act.Name)
		r.CheckField(t, recordID, "Username", exp.Username, act.Username)
		r.CheckField(t, recordID, "FilterHostname", exp.FilterHostname, act.FilterHostname)

		if !bytes.Equal(exp.Data, act.Data) {
			r.AddMismatch(FieldMismatch{
				Table: t, RecordID: recordID, Field: "Data",
				Expected: fmt.Sprintf("[]byte(len=%d)", len(exp.Data)),
				Actual:   fmt.Sprintf("[]byte(len=%d)", len(act.Data)),
			})
		}
	}

	return nil
}

func (m *Migrator) deepValidatePublicKeyTags(ctx context.Context, r *ValidationReport) error {
	cursor, err := m.mongo.Collection("public_keys").Find(ctx, bson.M{
		"filter.tag_ids": bson.M{"$exists": true, "$ne": bson.A{}},
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	for cursor.Next(ctx) {
		var doc mongoPublicKey
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		expectedTags := make([]string, len(doc.Filter.TagIDs))
		for i, tagID := range doc.Filter.TagIDs {
			expectedTags[i] = ObjectIDToUUID(tagID)
		}
		sort.Strings(expectedTags)

		var actualTags []entity.PublicKeyTag
		if err := m.pg.NewSelect().Model(&actualTags).
			Where("public_key_fingerprint = ?", doc.Fingerprint).
			Where("public_key_namespace_id = ?", doc.TenantID).
			Scan(ctx); err != nil {
			return err
		}

		actualTagIDs := make([]string, len(actualTags))
		for i, pkt := range actualTags {
			actualTagIDs[i] = pkt.TagID
		}
		sort.Strings(actualTagIDs)

		recordID := doc.Fingerprint + ":" + doc.TenantID
		r.AddCompared("public_key_tags", 1)
		r.CheckStrings("public_key_tags", recordID, "TagIDs", expectedTags, actualTagIDs)
	}

	return cursor.Err()
}

func (m *Migrator) deepValidateAPIKeys(ctx context.Context, r *ValidationReport) error {
	cursor, err := m.mongo.Collection("api_keys").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	batch := make([]mongoAPIKey, 0, batchSize)

	for cursor.Next(ctx) {
		var doc mongoAPIKey
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		batch = append(batch, doc)
		if len(batch) >= batchSize {
			if err := m.compareAPIKeyBatch(ctx, r, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		return m.compareAPIKeyBatch(ctx, r, batch)
	}

	return nil
}

func (m *Migrator) compareAPIKeyBatch(ctx context.Context, r *ValidationReport, batch []mongoAPIKey) error {
	digests := make([]string, len(batch))
	expected := make(map[string]*entity.APIKey, len(batch))
	for i, doc := range batch {
		e := convertAPIKey(doc)
		digests[i] = e.KeyDigest
		expected[e.KeyDigest] = e
	}

	var actual []entity.APIKey
	if err := m.pg.NewSelect().Model(&actual).Where("key_digest IN (?)", bun.List(digests)).Scan(ctx); err != nil {
		return err
	}

	pgMap := make(map[string]*entity.APIKey, len(actual))
	for i := range actual {
		pgMap[actual[i].KeyDigest] = &actual[i]
	}

	r.AddCompared("api_keys", int64(len(batch)))

	for _, digest := range digests {
		exp := expected[digest]
		act, ok := pgMap[digest]
		if !ok {
			r.AddMissing("api_keys", digest)

			continue
		}

		t := "api_keys"
		r.CheckField(t, digest, "NamespaceID", exp.NamespaceID, act.NamespaceID)
		r.CheckField(t, digest, "Name", exp.Name, act.Name)
		r.CheckField(t, digest, "Role", exp.Role, act.Role)
		r.CheckField(t, digest, "UserID", exp.UserID, act.UserID)
		r.CheckTime(t, digest, "CreatedAt", exp.CreatedAt, act.CreatedAt)
		r.CheckTime(t, digest, "UpdatedAt", exp.UpdatedAt, act.UpdatedAt)
		r.CheckField(t, digest, "ExpiresIn", exp.ExpiresIn, act.ExpiresIn)
	}

	return nil
}
