package mongo

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Store) UserList(ctx context.Context, pagination paginator.Query, filters []models.Filter) ([]models.User, int, error) {
	query := []bson.M{}

	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
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

	queryMatch, err := buildFilterQuery(filters)
	if err != nil {
		return nil, 0, err
	}

	if len(queryMatch) > 0 {
		query = append(query, queryMatch...)
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("users"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	if pagination.Page > 0 && pagination.PerPage > 0 {
		query = append(query, buildPaginationQuery(pagination)...)
	}

	users := make([]models.User, 0)
	cursor, err := s.db.Collection("users").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		user := new(models.User)
		err = cursor.Decode(&user)
		if err != nil {
			return nil, 0, err
		}

		users = append(users, *user)
	}

	return users, count, err
}

func (s *Store) UserCreate(ctx context.Context, user *models.User) error {
	_, err := s.db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error") {
			return store.ErrDuplicateEmail
		}
	}

	return err
}

func (s *Store) UserGetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := new(models.User)

	if err := s.db.Collection("users").FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) UserGetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)

	if err := s.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) UserGetByID(ctx context.Context, ID string) (*models.User, error) {
	user := new(models.User)
	objID, _ := primitive.ObjectIDFromHex(ID)
	if err := s.db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) UserUpdate(ctx context.Context, name, username, email, currentPassword, newPassword, ID string) error {
	user, err := s.UserGetByID(ctx, ID)
	objID, _ := primitive.ObjectIDFromHex(ID)

	if err != nil {
		return err
	}

	if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"name": name}}); err != nil {
		return err
	}

	if username != "" && username != user.Username {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"username": username}}); err != nil {
			return err
		}
	}

	if email != "" && email != user.Email {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"email": email}}); err != nil {
			return err
		}
	}

	if newPassword != "" && newPassword != currentPassword {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"password": newPassword}}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) UserUpdateFromAdmin(ctx context.Context, name, username, email, password, ID string) error {
	user, err := s.UserGetByID(ctx, ID)
	objID, _ := primitive.ObjectIDFromHex(ID)

	if err != nil {
		return err
	}

	if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"name": name}}); err != nil {
		return err
	}

	if username != "" && username != user.Username {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"username": username}}); err != nil {
			return err
		}
	}

	if email != "" && email != user.Email {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"email": email}}); err != nil {
			return err
		}
	}

	if password != "" {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"password": password}}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) UserDelete(ctx context.Context, ID string) error {
	objID, _ := primitive.ObjectIDFromHex(ID)
	_, err := s.db.Collection("users").DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	findOptions := options.Find()

	cursor, err := s.db.Collection("namespaces").Find(ctx, bson.M{"owner": ID}, findOptions)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		namespace := new(models.Namespace)
		err = cursor.Decode(&namespace)
		if err != nil {
			return err
		}

		err = s.NamespaceDelete(ctx, namespace.TenantID)
		if err != nil {
			return err
		}
	}
	return err
}
