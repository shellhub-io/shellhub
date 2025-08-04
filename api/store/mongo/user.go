package mongo

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Store) UserList(ctx context.Context, opts ...store.QueryOption) ([]models.User, int, error) {
	query := []bson.M{}
	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, 0, err
		}
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

	count, err := CountAllMatchingDocuments(ctx, s.db.Collection("users"), query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

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

func (s *Store) UserCreate(ctx context.Context, user *models.User) (string, error) {
	user.CreatedAt = time.Now()
	user.LastLogin = time.Time{}

	r, err := s.db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		return "", FromMongoError(err)
	}

	return r.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (s *Store) UserCreateInvited(ctx context.Context, email string) (string, error) {
	user := structToBson(models.User{CreatedAt: clock.Now(), Status: models.UserStatusInvited, UserData: models.UserData{Email: email}})
	sanitizeBson(user)

	r, err := s.db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		return "", FromMongoError(err)
	}

	return r.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (s *Store) UserResolve(ctx context.Context, resolver store.UserResolver, value string, opts ...store.QueryOption) (*models.User, error) {
	matchStage := bson.M{}
	switch resolver {
	case store.UserIDResolver:
		objID, err := primitive.ObjectIDFromHex(value)
		if err != nil {
			return nil, err
		}

		matchStage["_id"] = objID
	case store.UserEmailResolver:
		matchStage["email"] = value
	case store.UserUsernameResolver:
		matchStage["username"] = value
	}

	query := []bson.M{{"$match": matchStage}}
	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, err
		}
	}

	cursor, err := s.db.Collection("users").Aggregate(ctx, query)
	if err != nil {
		return nil, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	cursor.Next(ctx)

	user := new(models.User)
	if err := cursor.Decode(&user); err != nil {
		return nil, FromMongoError(err)
	}

	return user, nil
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

		if ns.Owner == id {
			userInfo.OwnedNamespaces = append(userInfo.OwnedNamespaces, *ns)
		} else {
			userInfo.AssociatedNamespaces = append(userInfo.AssociatedNamespaces, *ns)
		}
	}

	return userInfo, nil
}
