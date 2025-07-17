package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo/queries"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Store) TagCreate(ctx context.Context, tag *models.Tag) (string, error) {
	id := primitive.NewObjectID()

	upsert := bson.M{
		"$setOnInsert": bson.M{"_id": id},
		"$set": bson.M{
			"name":       tag.Name,
			"tenant_id":  tag.TenantID,
			"created_at": clock.Now(),
			"updated_at": clock.Now(),
		},
	}

	_, err := s.db.
		Collection("tags").
		UpdateOne(ctx, bson.M{"tenant_id": tag.TenantID, "name": tag.Name}, upsert, options.Update().SetUpsert(true))
	if err != nil {
		return "", FromMongoError(err)
	}

	return id.Hex(), nil
}

func (s *Store) TagConflicts(ctx context.Context, tenantID string, target *models.TagConflicts) ([]string, bool, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"tenant_id": tenantID,
				"$or":       []bson.M{{"name": target.Name}},
			},
		},
	}

	cursor, err := s.db.Collection("tags").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, false, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	tag := new(models.Tag)
	conflicts := make([]string, 0)

	for cursor.Next(ctx) {
		if err := cursor.Decode(&tag); err != nil {
			return nil, false, FromMongoError(err)
		}

		if tag.Name == target.Name {
			conflicts = append(conflicts, "name")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (s *Store) TagList(ctx context.Context, tenantID string, paginator query.Paginator, filters query.Filters, sorter query.Sorter) ([]models.Tag, int, error) {
	query := []bson.M{}
	if tenantID != "" {
		query = append(query, bson.M{"$match": bson.M{"tenant_id": tenantID}})
	}

	queryMatch, err := queries.FromFilters(&filters)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	query = append(query, queryMatch...)

	countQuery := append(query, bson.M{"$count": "count"})
	count, err := AggregateCount(ctx, s.db.Collection("tags"), countQuery)
	if err != nil {
		return []models.Tag{}, 0, err
	}

	if sorter.Order == "" {
		sorter.Order = "desc"
	}

	if sorter.By == "" {
		sorter.By = "created_at"
	}

	query = append(query, queries.FromSorter(&sorter)...)
	query = append(query, queries.FromPaginator(&paginator)...)

	cursor, err := s.db.Collection("tags").Aggregate(ctx, query)
	if err != nil {
		return []models.Tag{}, 0, err
	}
	defer cursor.Close(ctx)

	tags := make([]models.Tag, 0)
	for cursor.Next(ctx) {
		tag := new(models.Tag)
		if err := cursor.Decode(tag); err != nil {
			return []models.Tag{}, 0, err
		}

		tags = append(tags, *tag)
	}

	return tags, count, err
}

func (s *Store) TagResolve(ctx context.Context, resolver store.TagResolver, value string, opts ...store.QueryOption) (*models.Tag, error) {
	tag := new(models.Tag)
	if err := s.cache.Get(ctx, "tag={"+value+"}", tag); err == nil && tag.ID != "" {
		return tag, nil
	}

	matchStage := bson.M{}
	switch resolver {
	case store.TagIDResolver:
		objID, err := primitive.ObjectIDFromHex(value)
		if err != nil {
			return nil, err
		}

		matchStage["_id"] = objID
	case store.TagNameResolver:
		matchStage["name"] = value
	}

	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &matchStage)); err != nil {
			return nil, err
		}
	}

	cursor, err := s.db.Collection("tags").Aggregate(ctx, []bson.M{{"$match": matchStage}})
	if err != nil {
		return nil, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	cursor.Next(ctx)

	tag = nil
	if err := cursor.Decode(&tag); err != nil {
		return nil, FromMongoError(err)
	}

	if err := s.cache.Set(ctx, "tag={"+value+"}", tag, time.Hour); err != nil {
		log.WithError(err).Error("failed to store tag in cache")
	}

	return tag, nil
}

func (s *Store) TagUpdate(ctx context.Context, tenantID, name string, changes *models.TagChanges) error {
	tag := new(models.Tag)
	if err := s.db.Collection("tags").FindOneAndUpdate(ctx, bson.M{"tenant_id": tenantID, "name": name}, bson.M{"$set": changes}).Decode(tag); err != nil {
		return FromMongoError(err)
	}

	for _, key := range []string{"tag={" + tag.ID + "}", "tag={" + tag.TenantID + "," + tag.Name + "}"} {
		if err := s.cache.Delete(ctx, key); err != nil {
			log.WithError(err).Error("failed to delete tag from cache")
		}
	}

	return nil
}

func (s *Store) TagPushToTarget(ctx context.Context, tenantID, name string, target models.TagTarget, targetID string) error {
	tag, err := s.TagResolve(ctx, store.TagNameResolver, name, s.options.InNamespace(tenantID))
	if err != nil {
		return err
	}

	collection, filter, attribute, err := getTargetCollectionMetadata(target)
	if err != nil {
		return err
	}

	res, err := s.db.
		Collection(collection).
		UpdateOne(ctx, bson.M{filter: targetID}, bson.M{"$addToSet": bson.M{attribute: tag.ID}})

	if res.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return FromMongoError(err)
}

func (s *Store) TagPullFromTarget(ctx context.Context, tenantID, name string, target models.TagTarget, targetsID ...string) error {
	tag, err := s.TagResolve(ctx, store.TagNameResolver, name, s.options.InNamespace(tenantID))
	if err != nil {
		return err
	}

	collection, filter, attribute, err := getTargetCollectionMetadata(target)
	if err != nil {
		return err
	}

	if len(targetsID) > 0 {
		res, err := s.db.
			Collection(collection).
			UpdateMany(ctx, bson.M{filter: bson.M{"$in": targetsID}}, bson.M{"$pull": bson.M{attribute: tag.ID}})
		if err != nil {
			return FromMongoError(err)
		}

		if res.MatchedCount < 1 {
			return store.ErrNoDocuments
		}

		return nil
	}

	_, err = s.db.Collection(collection).UpdateMany(ctx, bson.M{}, bson.M{"$pull": bson.M{"tags": tag.ID}})

	return FromMongoError(err)
}

func (s *Store) TagDelete(ctx context.Context, tenantID, name string) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	sessionCallback := func(sessCtx mongo.SessionContext) (any, error) {
		tag := new(models.Tag)
		if err := s.db.Collection("tags").FindOneAndDelete(sessCtx, bson.M{"tenant_id": tenantID, "name": name}).Decode(tag); err != nil {
			return nil, FromMongoError(err)
		}

		if _, err := s.db.Collection("devices").UpdateMany(sessCtx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"tags": tag.ID}}); err != nil {
			return nil, FromMongoError(err)
		}

		for _, c := range []string{"public_keys", "firewall_rules"} {
			if _, err := s.db.Collection(c).UpdateMany(sessCtx, bson.M{"tenant_id": tenantID}, bson.M{"$pull": bson.M{"filters.tags": tag.ID}}); err != nil {
				return nil, FromMongoError(err)
			}
		}

		for _, key := range []string{"tag={" + tag.ID + "}", "tag={" + tag.TenantID + "," + tag.Name + "}"} {
			if err := s.cache.Delete(sessCtx, key); err != nil {
				log.WithError(err).Error("failed to delete tag from cache")
			}
		}

		return nil, nil
	}

	_, err = session.WithTransaction(ctx, sessionCallback)

	return err
}

func getTargetCollectionMetadata(target models.TagTarget) (string, string, string, error) {
	switch target {
	case models.TagTargetDevice:
		return "devices", "uid", "tags", nil
	case models.TagTargetPublicKey:
		return "public_keys", "fingerprint", "filter.tags", nil
	case models.TagTargetFirewallRule:
		return "firewall_rules", "_id", "", nil
	default:
		return "", "", "", errors.New("invalid tag target")
	}
}
