package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/hash"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type UserStatus string

const (
	// UserStatusInvited applies to cloud-only instances. This status is assigned to a user who has been invited to a
	// namespace but has not yet completed the registration process.
	UserStatusInvited UserStatus = "invited"

	// UserStatusNotConfirmed applies to cloud-only instances. This status is assigned to a user who has registered
	// but has not yet confirmed their email address.
	UserStatusNotConfirmed UserStatus = "not-confirmed"

	// UserStatusConfirmed indicates that the user has completed the registration process and confirmed their email address.
	// Users in community and enterprise instances will always be created with this status.
	UserStatusConfirmed UserStatus = "confirmed"
)

func (s UserStatus) String() string {
	return string(s)
}

type UserOrigin string

const (
	// UserOriginLocal indicates that the user was created through the standard signup process, without
	// using third-party integrations like SSO IdPs.
	UserOriginLocal UserOrigin = "local"
)

func (o UserOrigin) String() string {
	return string(o)
}

type User struct {
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
	LastLogin      time.Time `json:"last_login" bson:"last_login"`
	UserData       `bson:",inline"`
	Password       UserPassword    `bson:",inline"`
	ID             string          `json:"id,omitempty" bson:"_id,omitempty"`
	Origin         UserOrigin      `json:"-" bson:"origin"`
	Status         UserStatus      `json:"status" bson:"status"`
	Preferences    UserPreferences `json:"-" bson:"preferences"`
	MFA            UserMFA         `json:"mfa" bson:"mfa"`
	MaxNamespaces  int             `json:"max_namespaces" bson:"max_namespaces"`
	EmailMarketing bool            `json:"email_marketing" bson:"email_marketing"`
}

type UserData struct {
	Name     string `json:"name" validate:"required,name"`
	Username string `json:"username" bson:"username" validate:"required,username"`
	Email    string `json:"email" bson:"email" validate:"required,email"`
	// RecoveryEmail is a custom, non-unique email address that a user can use to recover their account
	// when they lose access to all other methods. It must never be equal to [UserData.Email].
	//
	// NOTE: Recovery email is available as a cloud-only feature and must be ignored in community.
	RecoveryEmail string `json:"recovery_email" bson:"recovery_email" validate:"omitempty,email"`
}

// UserMFA represents the attributes related to MFA for a user.
type UserMFA struct {
	Secret        string   `json:"-" bson:"secret"`
	RecoveryCodes []string `json:"-" bson:"recovery_codes"`
	Enabled       bool     `json:"enabled" bson:"enabled"`
}

type UserPreferences struct {
	// PreferredNamespace represents the namespace the user most recently authenticated with.
	PreferredNamespace string `json:"-" bson:"preferred_namespace"`
}

type UserPassword struct {
	// Plain contains the plain text password.
	Plain string `json:"password" bson:"-" validate:"required,password"`
	// Hash contains the hashed pasword from plain text.
	Hash string `json:"-" bson:"password"`
}

// HashUserPassword receives a plain password and hash it, returning
// a [UserPassword].
func HashUserPassword(plain string) (UserPassword, error) {
	p := UserPassword{
		Plain: plain,
	}

	var err error
	p.Hash, err = hash.Do(p.Plain)

	return p, err
}

// Compare reports whether a plain password matches with hash.
//
// For compatibility purposes, it can compare using both SHA256 and bcrypt algorithms.
// Hashes starting with "$" are assumed to be a bcrypt hash; otherwise, they are treated as
// SHA256 hashes.
func (p *UserPassword) Compare(plain string) bool {
	return hash.CompareWith(plain, p.Hash)
}

// UserAuthIdentifier is an username or email used to authenticate.
type UserAuthIdentifier string

// IsEmail checks if the identifier is an email.
func (i *UserAuthIdentifier) IsEmail() bool {
	if ok, err := validator.New().Var(i, "required,email"); !ok || err != nil {
		return false
	}

	return true
}

type UserAuthResponse struct {
	Token         string `json:"token"`
	User          string `json:"user"`
	Origin        string `json:"string"`
	Name          string `json:"name"`
	ID            string `json:"id"`
	Tenant        string `json:"tenant"`
	Email         string `json:"email"`
	RecoveryEmail string `json:"recovery_email"`
	Role          string `json:"role"`
	MFA           bool   `json:"mfa"`
	MaxNamespaces int    `json:"max_namespaces"`
}

// NOTE: This struct has been moved to the cloud repo as it is only used in a cloud context;
// however, it is also utilized by migrations. For this reason, we must maintain the struct
// here ensure everything continues to function as expected.
// TODO: Remove this struct when it is no longer needed for migrations.
type UserTokenRecover struct {
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	Token     string    `json:"uid"`
	User      string    `json:"user_id"`
}

// UserChanges specifies the attributes that can be updated for a user. Any zero values in this
// struct must be ignored. If an attribute is a pointer type, its zero value is represented as `nil`.
type UserChanges struct {
	LastLogin          time.Time  `bson:"last_login,omitempty"`
	PreferredNamespace *string    `bson:"preferences.preferred_namespace,omitempty"`
	MaxNamespaces      *int       `bson:"max_namespaces,omitempty"`
	EmailMarketing     *bool      `bson:"email_marketing,omitempty"`
	Name               string     `bson:"name,omitempty"`
	Username           string     `bson:"username,omitempty"`
	Email              string     `bson:"email,omitempty"`
	RecoveryEmail      string     `bson:"recovery_email,omitempty"`
	Password           string     `bson:"password,omitempty"`
	Status             UserStatus `bson:"status,omitempty"`
}

// UserConflicts holds user attributes that must be unique for each itam and can be utilized in queries
// to identify conflicts.
type UserConflicts struct {
	Email    string
	Username string
}

// Distinct removes the c attributes whether it's equal to the user attribute.
func (c *UserConflicts) Distinct(user *User) {
	if c.Email == user.Email {
		c.Email = ""
	}

	if c.Username == user.Username {
		c.Username = ""
	}
}

type UserInfo struct {
	// OwnedNamespaces are the namespaces where the user is the owner.
	OwnedNamespaces []Namespace
	// AssociatedNamespaces are the namespaces where the user is a member.
	AssociatedNamespaces []Namespace
}
