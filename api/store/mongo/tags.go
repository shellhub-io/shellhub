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
