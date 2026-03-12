package migrate

import (
	"context"
	"strings"
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

// usernameOrFallback returns the username if non-empty, otherwise falls back
// to the email address. This handles legacy users created via SSO/API that
// have NULL usernames in MongoDB.
func usernameOrFallback(username, email string) string {
	if username != "" {
		return username
	}

	if email != "" {
		return email
	}

	return "unknown"
}

func convertUser(doc mongoUser) *entity.User {
	origin := strings.ToLower(doc.Origin)
	if origin == "" {
		origin = "local"
	}

	status := strings.ToLower(doc.Status)
	if status == "" {
		status = "confirmed"
	}

	authMethods := make([]string, len(doc.Preferences.AuthMethods))
	for i, m := range doc.Preferences.AuthMethods {
		authMethods[i] = strings.ToLower(m)
	}
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
		Username:       usernameOrFallback(doc.Username, doc.Email),
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
	// Load valid namespace IDs so we can clear dangling
	// preferred_namespace references that would violate the FK.
	validNS, err := m.loadValidNamespaces(ctx)
	if err != nil {
		return err
	}

	cursor, err := m.mongo.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.User, 0, batchSize)
	total := 0
	cleared := 0

	for cursor.Next(ctx) {
		var doc mongoUser
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		u := convertUser(doc)
		if u.Preferences.PreferredNamespace != "" {
			if _, ok := validNS[u.Preferences.PreferredNamespace]; !ok {
				log.WithFields(log.Fields{
					"scope":     "core",
					"user":      u.Username,
					"namespace": u.Preferences.PreferredNamespace,
				}).Warn("Clearing dangling preferred_namespace_id")
				u.Preferences.PreferredNamespace = ""
				cleared++
			}
		}

		batch = append(batch, u)
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

	log.WithFields(log.Fields{
		"scope":   "core",
		"count":   total,
		"cleared": cleared,
	}).Info("Migrated users")

	return nil
}
