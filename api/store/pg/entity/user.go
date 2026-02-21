package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	ID             string          `bun:"id,pk,type:uuid"`
	CreatedAt      time.Time       `bun:"created_at"`
	UpdatedAt      time.Time       `bun:"updated_at"`
	LastLogin      time.Time       `bun:"last_login,nullzero"`
	Origin         string          `bun:"origin"`
	ExternalID     string          `bun:"external_id,nullzero"`
	Status         string          `bun:"status"`
	Name           string          `bun:"name"`
	Username       string          `bun:"username"`
	Email          string          `bun:"email"`
	PasswordDigest string          `bun:"password_digest"`
	Preferences    UserPreferences `bun:"embed:"`
	MFA            UserMFA         `bun:"-"`
	Admin          bool            `bun:"admin"`
	Namespaces     int             `bun:"namespaces,scanonly"`
}

type UserPreferences struct {
	PreferredNamespace string   `bun:"preferred_namespace_id,nullzero"`
	AuthMethods        []string `bun:"auth_methods,array"`
	SecurityEmail      string   `bun:"security_email,nullzero"`
	MaxNamespaces      int      `bun:"namespace_ownership_limit"`
	EmailMarketing     bool     `bun:"email_marketing"`
}

type UserMFA struct {
	Enabled       bool     `bun:"enabled"`
	Secret        string   `bun:"secret,nullzero"`
	RecoveryCodes []string `bun:"recovery_codes,nullzero,array"`
}

func UserFromModel(model *models.User) *User {
	authMethods := make([]string, len(model.Preferences.AuthMethods))
	for i, method := range model.Preferences.AuthMethods {
		authMethods[i] = method.String()
	}

	// Default to local if Origin is empty (for test cases)
	origin := model.Origin.String()
	if origin == "" {
		origin = string(models.UserOriginLocal)
	}

	// Default to confirmed if Status is empty (for test cases)
	status := model.Status.String()
	if status == "" {
		status = string(models.UserStatusConfirmed)
	}

	return &User{
		ID:             model.ID,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      time.Time{},
		LastLogin:      model.LastLogin,
		Origin:         origin,
		ExternalID:     model.ExternalID,
		Status:         status,
		Name:           model.Name,
		Username:       model.Username,
		Email:          model.Email,
		PasswordDigest: model.Password.Hash,
		Admin:          model.Admin,
		Preferences: UserPreferences{
			PreferredNamespace: model.Preferences.PreferredNamespace,
			AuthMethods:        authMethods,
			SecurityEmail:      model.UserData.RecoveryEmail,
			MaxNamespaces:      model.MaxNamespaces,
			EmailMarketing:     model.EmailMarketing,
		},
		MFA: UserMFA{
			Enabled:       model.MFA.Enabled,
			Secret:        model.MFA.Secret,
			RecoveryCodes: model.MFA.RecoveryCodes,
		},
	}
}

func UserToModel(entity *User) *models.User {
	authMethods := make([]models.UserAuthMethod, len(entity.Preferences.AuthMethods))
	for i, method := range entity.Preferences.AuthMethods {
		authMethods[i] = models.UserAuthMethod(method)
	}

	return &models.User{
		ID:             entity.ID,
		Origin:         models.UserOrigin(entity.Origin),
		ExternalID:     entity.ExternalID,
		Status:         models.UserStatus(entity.Status),
		MaxNamespaces:  entity.Preferences.MaxNamespaces,
		CreatedAt:      entity.CreatedAt,
		LastLogin:      entity.LastLogin,
		EmailMarketing: entity.Preferences.EmailMarketing,
		Admin:          entity.Admin,
		UserData: models.UserData{
			Name:          entity.Name,
			Username:      entity.Username,
			Email:         entity.Email,
			RecoveryEmail: entity.Preferences.SecurityEmail,
		},
		Password: models.UserPassword{
			Hash: entity.PasswordDigest,
		},
		MFA: models.UserMFA{
			Enabled:       entity.MFA.Enabled,
			Secret:        entity.MFA.Secret,
			RecoveryCodes: entity.MFA.RecoveryCodes,
		},
		Preferences: models.UserPreferences{
			PreferredNamespace: entity.Preferences.PreferredNamespace,
			AuthMethods:        authMethods,
		},
	}
}
