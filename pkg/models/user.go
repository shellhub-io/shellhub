package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/validator"
)

// UserAuthIdentifier is an username or email used to authenticate.
type UserAuthIdentifier string

// IsEmail checks if the identifier is an email.
func (i *UserAuthIdentifier) IsEmail() bool {
	if ok, err := validator.New().Var(i, "required,email"); !ok || err != nil {
		return false
	}

	return true
}

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

	// UserOriginSAML indicates that the user was created using a SAML method.
	UserOriginSAML UserOrigin = "SAML"
)

func (o UserOrigin) String() string {
	return string(o)
}

type UserAuthMethod string

const (
	// UserAuthMethodLocal indicates that the user can authenticate using an email and password.
	UserAuthMethodLocal UserAuthMethod = "local"

	// UserAuthMethodManual indicates that the user can authenticate using a third-party SAML application.
	UserAuthMethodSAML UserAuthMethod = "saml"
)

func (a UserAuthMethod) String() string {
	return string(a)
}

type User struct {
	// ID is the primary identifier for the user
	ID string `json:"id,omitempty" bun:"id,pk,type:char"`

	// CreatedAt represents the timestamp when the user was created
	CreatedAt time.Time `json:"created_at" bun:"created_at"`
	// UpdatedAt represents the timestamp when the user was last updated
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at"`
	// LastLogin represents the timestamp of the user's most recent login. This field may be zero value for users who have
	// never logged in.
	LastLogin time.Time `json:"last_login" bun:"last_login,nullzero"`

	// Origin specifies the user's signup/registration method (e.g., local, OAuth, SSO).
	Origin UserOrigin `json:"-" bun:"origin"`
	// ExternalID represents the user's identifier in an external authentication system. It is only populated when
	// [User.Origin] is not [UserOriginLocal]
	ExternalID string `json:"-" bun:"external_id,nullzero"`

	// Status indicates the current state of the user account (e.g., active, suspended, pending).
	Status UserStatus `json:"status" bun:"status"`

	Name     string `json:"name" bun:"name"`
	Username string `json:"username" bun:"username"`
	Email    string `json:"email" bun:"email"`
	// PasswordDigest stores the hashed password.
	PasswordDigest string `json:"-" bun:"password_digest"`

	Preferences UserPreferences `json:"preferences" bun:"embed:"`
	MFA         UserMFA         `json:"mfa" bun:"-"`
}

type UserPreferences struct {
	// PreferredNamespace represents the namespace ID the user most recently authenticated with.
	// This field may be null if the user was never logged in or the namespace was deleted.
	PreferredNamespace string `json:"-" bun:"preferred_namespace_id,nullzero"`

	// AuthMethods indicates the authentication methods that the user can use to authenticate.
	AuthMethods []UserAuthMethod `json:"auth_methods" bun:"auth_methods,array"`

	// SecurityEmail is a secondary email address used for account recovery.
	SecurityEmail string `json:"recovery_email" bun:"security_email,nullzero"`

	// MaxNamespaces defines the maximum number of namespaces that a user can own. Values can be 0 for no namespaces, -1
	// for unlimited namespaces or positive integer for specific limit
	MaxNamespaces int `json:"max_namespaces" bun:"namespace_ownership_limit"`

	// EmailMarketing reports whether the user has opted in to receive marketing communications. When true, marketing
	// emails may be sent to the user's primary email address.
	EmailMarketing bool `json:"email_marketing" bun:"email_marketing"`
}

// UserMFA represents the attributes related to MFA for a user.
type UserMFA struct {
	// Enabled reports whether MFA is enabled for the user.
	Enabled bool `json:"enabled" bun:"enabled"`
	// Secret is the key used for authenticating with the OTP server.
	Secret string `json:"-" bun:"secret,nullzero"`
	// RecoveryCodes are recovery tokens that the user can use to regain account access if they lose their MFA device.
	RecoveryCodes []string `json:"-" bun:"recovery_codes,nullzero,array"`
}

type UserAuthResponse struct {
	Token         string           `json:"token"`
	User          string           `json:"user"`
	Origin        string           `json:"origin"`
	AuthMethods   []UserAuthMethod `json:"auth_methods"`
	Name          string           `json:"name"`
	ID            string           `json:"id"`
	Tenant        string           `json:"tenant"`
	Email         string           `json:"email"`
	RecoveryEmail string           `json:"recovery_email"`
	Role          string           `json:"role"`
	MFA           bool             `json:"mfa"`
	MaxNamespaces int              `json:"max_namespaces"`
}

// NOTE: This struct has been moved to the cloud repo as it is only used in a cloud context;
// however, it is also utilized by migrations. For this reason, we must maintain the struct
// here ensure everything continues to function as expected.
// TODO: Remove this struct when it is no longer needed for migrations.
type UserTokenRecover struct {
	Token     string    `json:"uid"`
	User      string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// UserChanges specifies the attributes that can be updated for a user. Any zero values in this
// struct must be ignored. If an attribute is a pointer type, its zero value is represented as `nil`.
type UserChanges struct {
	LastLogin          time.Time        `bson:"last_login,omitempty"`
	Name               string           `bson:"name,omitempty"`
	Username           string           `bson:"username,omitempty"`
	Email              string           `bson:"email,omitempty"`
	RecoveryEmail      string           `bson:"recovery_email,omitempty"`
	Password           string           `bson:"password,omitempty"`
	Status             UserStatus       `bson:"status,omitempty"`
	ExternalID         *string          `bson:"external_id,omitempty"`
	PreferredNamespace *string          `bson:"preferences.preferred_namespace,omitempty"`
	MaxNamespaces      *int             `bson:"max_namespaces,omitempty"`
	EmailMarketing     *bool            `bson:"email_marketing,omitempty"`
	AuthMethods        []UserAuthMethod `bson:"preferences.auth_methods,omitempty"`
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
