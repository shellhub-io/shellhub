package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo/queries"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Store) UserList(ctx context.Context, paginator query.Paginator, filters query.Filters) ([]models.User, int, error) {
	query := []bson.M{}

	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	query = append(query, []bson.M{
		{
			"$addFields": bson.M{
				"user_id": bson.M{"$toString": "$_id"},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "namespaces",
				"localField":   "user_id",
				"foreignField": "owner",
				"as":           "namespaces",
			},
		},
		{
			"$addFields": bson.M{
				"namespaces": bson.M{"$size": "$namespaces"},
			},
		},
	}...)

	queryMatch, err := queries.FromFilters(&filters)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}
	query = append(query, queryMatch...)

	queryCount := query
	queryCount = append(queryCount, bson.M{"$count": "count"})
	count, err := AggregateCount(ctx, s.db.Collection("users"), queryCount)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	query = append(query, queries.FromPaginator(&paginator)...)

	users := make([]models.User, 0)
	cursor, err := s.db.Collection("users").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		user := new(models.User)
		err = cursor.Decode(&user)
		if err != nil {
			return nil, 0, FromMongoError(err)
		}

		users = append(users, *user)
	}

	return users, count, FromMongoError(err)
}

func (s *Store) UserCreate(ctx context.Context, user *models.User) error {
	_, err := s.db.Collection("users").InsertOne(ctx, user)

	return FromMongoError(err)
}

func (s *Store) UserGetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := new(models.User)

	if err := s.db.Collection("users").FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
		return nil, FromMongoError(err)
	}

	return user, nil
}

func (s *Store) UserGetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)

	if err := s.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		return nil, FromMongoError(err)
	}

	return user, nil
}

func (s *Store) UserGetByID(ctx context.Context, id string, ns bool) (*models.User, int, error) {
	user := new(models.User)
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, 0, err
	}

	if err := s.db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
		return nil, 0, FromMongoError(err)
	}

	if !ns {
		return user, 0, nil
	}

	nss := struct {
		NamespacesOwned int `bson:"namespacesOwned"`
	}{}

	query := []bson.M{
		{
			"$match": bson.M{
				"_id": objID,
			},
		},
		{
			"$addFields": bson.M{
				"_id": bson.M{
					"$toString": "$_id",
				},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "namespaces",
				"localField":   "_id",
				"foreignField": "owner",
				"as":           "ns",
			},
		},
		{
			"$addFields": bson.M{
				"namespacesOwned": bson.M{
					"$size": "$ns",
				},
			},
		},
		{
			"$project": bson.M{
				"namespacesOwned": 1,
				"_id":             0,
			},
		},
	}

	cursor, err := s.db.Collection("users").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, 0, FromMongoError(err)
	}

	if err = cursor.Decode(&nss); err != nil {
		return nil, 0, FromMongoError(err)
	}

	return user, nss.NamespacesOwned, nil
}

func (s *Store) UserConflicts(ctx context.Context, target *models.UserConflicts) ([]string, bool, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"email": target.Email},
					{"username": target.Username},
				},
			},
		},
	}

	cursor, err := s.db.Collection("users").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, false, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	user := new(models.UserConflicts)
	conflicts := make([]string, 0)
	for cursor.Next(ctx) {
		if err := cursor.Decode(&user); err != nil {
			return nil, false, FromMongoError(err)
		}

		if user.Username == target.Username {
			conflicts = append(conflicts, "username")
		}

		if user.Email == target.Email {
			conflicts = append(conflicts, "email")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (s *Store) UserUpdate(ctx context.Context, id string, changes *models.UserChanges) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return FromMongoError(err)
	}

	r, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": changes})
	if err != nil {
		return FromMongoError(err)
	}

	if r.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) UserDelete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return FromMongoError(err)
	}

	user, err := s.db.Collection("users").DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return FromMongoError(err)
	}

	if user.DeletedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) UserGetInfo(ctx context.Context, id string) (*models.UserInfo, error) {
	cursor, err := s.db.Collection("namespaces").Find(ctx, bson.M{"members": bson.M{"$elemMatch": bson.M{"id": id}}})
	if err != nil {
		return nil, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	userInfo := &models.UserInfo{}

	for cursor.Next(ctx) {
		ns := new(models.Namespace)
		if err := cursor.Decode(ns); err != nil {
			return nil, FromMongoError(err)
		}

		if ns.Owner != id {
			userInfo.OwnedNamespaces = append(userInfo.OwnedNamespaces, *ns)
		} else {
			userInfo.AssociatedNamespaces = append(userInfo.AssociatedNamespaces, *ns)
		}
	}

	return userInfo, nil
}
