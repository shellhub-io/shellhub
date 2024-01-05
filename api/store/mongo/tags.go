package mongo

import (
	"context"

	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) TagsGet(ctx context.Context, tenant string) ([]string, int, error) {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return nil, 0, err
	}
	defer session.EndSession(ctx)

	tags, err := session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		deviceTags, _, err := s.DeviceGetTags(sessCtx, tenant)
		if err != nil {
			return nil, err
		}

		keyTags, _, err := s.PublicKeyGetTags(sessCtx, tenant)
		if err != nil {
			return nil, err
		}

		ruleTags, _, err := s.FirewallRuleGetTags(sessCtx, tenant)
		if err != nil {
			return nil, err
		}

		tags := []string{}
		tags = append(tags, deviceTags...)
		tags = append(tags, keyTags...)
		tags = append(tags, ruleTags...)

		return removeDuplicate[string](tags), nil
	})
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	return tags.([]string), len(tags.([]string)), nil
}

func (s *Store) TagsRename(ctx context.Context, tenantID string, oldTag string, newTag string) (int64, error) {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return int64(0), FromMongoError(err)
	}
	defer session.EndSession(ctx)

	count, err := session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		devCount, err := s.DeviceBulkRenameTag(sessCtx, tenantID, oldTag, newTag)
		if err != nil {
			return int64(0), err
		}

		keyCount, err := s.PublicKeyBulkRenameTag(sessCtx, tenantID, oldTag, newTag)
		if err != nil {
			return int64(0), err
		}

		rulCount, err := s.FirewallRuleBulkRenameTag(sessCtx, tenantID, oldTag, newTag)
		if err != nil {
			return int64(0), err
		}

		return devCount + keyCount + rulCount, nil
	})
	if err != nil {
		return int64(0), FromMongoError(err)
	}

	return count.(int64), nil
}

func (s *Store) TagsDelete(ctx context.Context, tenantID string, tag string) (int64, error) {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return int64(0), FromMongoError(err)
	}
	defer session.EndSession(ctx)

	count, err := session.WithTransaction(ctx, func(sessCtx mongodriver.SessionContext) (interface{}, error) {
		devCount, err := s.DeviceBulkDeleteTag(sessCtx, tenantID, tag)
		if err != nil {
			return int64(0), err
		}

		keyCount, err := s.PublicKeyBulkDeleteTag(sessCtx, tenantID, tag)
		if err != nil {
			return int64(0), err
		}

		rulCount, err := s.FirewallRuleBulkDeleteTag(sessCtx, tenantID, tag)
		if err != nil {
			return int64(0), err
		}

		return devCount + keyCount + rulCount, nil
	})
	if err != nil {
		return int64(0), FromMongoError(err)
	}

	return count.(int64), nil
}
