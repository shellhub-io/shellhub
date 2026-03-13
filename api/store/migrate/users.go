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

		u := convertUser(doc)

		// Clear preferred_namespace on insert; namespaces haven't been
		// migrated yet so the FK would fail. A follow-up pass restores
		// valid references after namespaces are in place.
		u.Preferences.PreferredNamespace = ""

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
		"scope": "core",
		"count": total,
	}).Info("Migrated users")

	return nil
}

// restorePreferredNamespaces updates users' preferred_namespace_id from MongoDB
// after both users and namespaces have been migrated to PG.
func (m *Migrator) restorePreferredNamespaces(ctx context.Context) error {
	validNS, err := m.loadValidNamespaces(ctx)
	if err != nil {
		return err
	}

	cursor, err := m.mongo.Collection("users").Find(ctx, bson.M{
		"preferences.preferred_namespace": bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	total := 0
	skipped := 0

	for cursor.Next(ctx) {
		var doc mongoUser
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		ns := doc.Preferences.PreferredNamespace
		if _, ok := validNS[ns]; !ok {
			skipped++

			continue
		}

		userID := ObjectIDToUUID(doc.ID.Hex())
		if _, err := m.pg.NewUpdate().
			TableExpr("users").
			Set("preferred_namespace_id = ?", ns).
			Where("id = ?", userID).
			Exec(ctx); err != nil {
			return err
		}

		total++
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"scope":   "core",
		"count":   total,
		"skipped": skipped,
	}).Info("Restored preferred_namespace references")

	return nil
}
