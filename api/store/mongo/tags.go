package mongo

import (
	"context"
	"fmt"

	"github.com/emirpasic/gods/sets/hashset"
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

		tagsRule, err := s.db.Collection("firewall_rules").Distinct(sessCtx, "filter.tags", bson.M{"tenant_id": tenant})
		if err != nil {
			return nil, err
		}

		tagsSet.Add(tagsDevice...)
		tagsSet.Add(tagsKey...)
		tagsSet.Add(tagsRule...)

		return nil, nil
	})

	tags := make([]string, tagsSet.Size())
	for i, v := range tagsSet.Values() {
		tags[i] = fmt.Sprint(v)
	}

	return tags, len(tags), err
}

func (s *Store) TagRename(ctx context.Context, tenantID string, oldTag string, newTag string) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return FromMongoError(err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		if _, err := s.DeviceRenameTag(sessCtx, tenantID, oldTag, newTag); err != nil {
			return nil, err
		}

		if _, err := s.PublicKeyRenameTag(sessCtx, tenantID, oldTag, newTag); err != nil {
			return nil, err
		}

		if _, err := s.FirewallRuleRenameTag(sessCtx, tenantID, oldTag, newTag); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

func (s *Store) TagDelete(ctx context.Context, tenantID string, tag string) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return FromMongoError(err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		if _, err := s.DeviceDeleteTag(sessCtx, tenantID, tag); err != nil {
			return nil, err
		}

		if _, err := s.PublicKeyDeleteTag(sessCtx, tenantID, tag); err != nil {
			return nil, err
		}

		if _, err := s.FirewallRuleDeleteTag(sessCtx, tenantID, tag); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return FromMongoError(err)
}
