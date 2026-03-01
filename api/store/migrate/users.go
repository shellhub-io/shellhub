package migrate

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mongoUser struct {
	ID             primitive.ObjectID `bson:"_id"`
	Origin         string             `bson:"origin"`
	ExternalID     string             `bson:"external_id"`
	Status         string             `bson:"status"`
	MaxNamespaces  int                `bson:"max_namespaces"`
	CreatedAt      time.Time          `bson:"created_at"`
	LastLogin      time.Time          `bson:"last_login"`
	EmailMarketing bool               `bson:"email_marketing"`
	Name           string             `bson:"name"`
	Username       string             `bson:"username"`
	Email          string             `bson:"email"`
	RecoveryEmail  string             `bson:"recovery_email"`
	Password       string             `bson:"password"`
	MFA            mongoUserMFA       `bson:"mfa"`
	Preferences    mongoUserPrefs     `bson:"preferences"`
	Admin          bool               `bson:"admin"`
}

type mongoUserMFA struct {
	Enabled       bool     `bson:"enabled"`
	Secret        string   `bson:"secret"`
	RecoveryCodes []string `bson:"recovery_codes"`
}

type mongoUserPrefs struct {
	PreferredNamespace string   `bson:"preferred_namespace"`
	AuthMethods        []string `bson:"auth_methods"`
}

func convertUser(doc mongoUser) *entity.User {
	origin := doc.Origin
	if origin == "" {
		origin = "local"
	}

	status := doc.Status
	if status == "" {
		status = "confirmed"
	}

	authMethods := doc.Preferences.AuthMethods
	if len(authMethods) == 0 {
		authMethods = []string{"local"}
	}

	return &entity.User{
		ID:             ObjectIDToUUID(doc.ID.Hex()),
		CreatedAt:      doc.CreatedAt,
		UpdatedAt:      time.Time{},
		LastLogin:      doc.LastLogin,
		Origin:         origin,
		ExternalID:     doc.ExternalID,
		Status:         status,
		Name:           doc.Name,
		Username:       doc.Username,
		Email:          doc.Email,
		PasswordDigest: doc.Password,
		Admin:          doc.Admin,
		Preferences: entity.UserPreferences{
			PreferredNamespace: doc.Preferences.PreferredNamespace,
			AuthMethods:        authMethods,
			SecurityEmail:      doc.RecoveryEmail,
			MaxNamespaces:      doc.MaxNamespaces,
			EmailMarketing:     doc.EmailMarketing,
		},
	}
}

func (m *Migrator) migrateUsers(ctx context.Context) error {
	cursor, err := m.mongo.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.User, 0, batchSize)
	total := 0

	for cursor.Next(ctx) {
		var doc mongoUser
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		batch = append(batch, convertUser(doc))
		if len(batch) >= batchSize {
			if _, err := m.pg.NewInsert().Model(&batch).Exec(ctx); err != nil {
				return err
			}
			total += len(batch)
			batch = batch[:0]
		}
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		if _, err := m.pg.NewInsert().Model(&batch).Exec(ctx); err != nil {
			return err
		}
		total += len(batch)
	}

	log.WithField("count", total).Info("Migrated users")

	return nil
}
