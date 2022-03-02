package mongo

import (
	"context"
	"fmt"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/shellhub-io/shellhub/api/store"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) TagsGet(ctx context.Context, tenant string) ([]string, int, error) {
	tagsDevice, err := s.db.Collection("devices").Distinct(ctx, "tags", bson.M{"tenant_id": tenant})
	if err != nil {
		return nil, 0, err
	}

	tagsKey, err := s.db.Collection("public_keys").Distinct(ctx, "filter.tags", bson.M{"tenant_id": tenant})
	if err != nil {
		return nil, 0, err
	}

	tagsSet := hashset.New()
	tagsSet.Add(tagsDevice...)
	tagsSet.Add(tagsKey...)

	tags := make([]string, tagsSet.Size())
	for i, v := range tagsSet.Values() {
		tags[i] = fmt.Sprint(v)
	}

	return tags, len(tags), err
}

func (s *Store) TagRename(ctx context.Context, tenantID string, tag string, newTag string) error {
	// Create a session to run the transaction.
	session, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Rename all devices tags inside a transaction.
	_, err = session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		_, err := s.db.Collection("devices").UpdateMany(ctx, bson.M{"tags": tag, "tenant_id": tenantID}, bson.M{"$set": bson.M{"tags.$": newTag}})
		if err != nil {
			return nil, fromMongoError(err)
		}

		if err := s.PublicKeyRenameTag(ctx, tenantID, tag, newTag); err != store.ErrNoDocuments {
			return nil, err
		}

		return nil, nil
	})

	return err
}

func (s *Store) TagDelete(ctx context.Context, tenantID string, tag string) error {
	_, err := s.db.Collection("devices").UpdateMany(ctx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"tags": tag}})

	if err := s.PublicKeyDeleteTag(ctx, tenantID, tag); err != store.ErrNoDocuments {
		return err
	}

	return err
}
