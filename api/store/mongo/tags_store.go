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
	session, err := s.db.Client().StartSession()
	if err != nil {
		return nil, 0, err
	}
	defer session.EndSession(ctx)

	tagsSet := hashset.New()
	_, err = session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		tagsDevice, err := s.db.Collection("devices").Distinct(sessCtx, "tags", bson.M{"tenant_id": tenant})
		if err != nil {
			return nil, err
		}

		tagsKey, err := s.db.Collection("public_keys").Distinct(sessCtx, "filter.tags", bson.M{"tenant_id": tenant})
		if err != nil {
			return nil, err
		}

		tagsSet.Add(tagsDevice...)
		tagsSet.Add(tagsKey...)

		return nil, nil
	})

	tags := make([]string, tagsSet.Size())
	for i, v := range tagsSet.Values() {
		tags[i] = fmt.Sprint(v)
	}

	return tags, len(tags), err
}

func (s *Store) TagRename(ctx context.Context, tenantID string, tag string, newTag string) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		_, err := s.db.Collection("devices").UpdateMany(sessCtx, bson.M{"tags": tag, "tenant_id": tenantID}, bson.M{"$set": bson.M{"tags.$": newTag}})
		if err != nil {
			return nil, fromMongoError(err)
		}

		if err := s.PublicKeyRenameTag(sessCtx, tenantID, tag, newTag); err != store.ErrNoDocuments {
			return nil, err
		}

		return nil, nil
	})

	return err
}

func (s *Store) TagDelete(ctx context.Context, tenantID string, tag string) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		_, err := s.db.Collection("devices").UpdateMany(sessCtx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"tags": tag}})
		if err != nil {
			return nil, err
		}

		if err := s.PublicKeyDeleteTag(sessCtx, tenantID, tag); err != store.ErrNoDocuments {
			return nil, err
		}

		return nil, nil
	})

	return err
}
