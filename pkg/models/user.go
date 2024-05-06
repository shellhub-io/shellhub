package models

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/password"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type User struct {
	ID             string    `json:"id,omitempty" bson:"_id,omitempty"`
	Namespaces     int       `json:"namespaces" bson:"namespaces,omitempty"`
	MaxNamespaces  int       `json:"max_namespaces" bson:"max_namespaces"`
	Confirmed      bool      `json:"confirmed"`
	CreatedAt      time.Time `json:"created_at" bson:"created_at"`
	LastLogin      time.Time `json:"last_login" bson:"last_login"`
	EmailMarketing bool      `json:"email_marketing" bson:"email_marketing"`
	MFA            bool      `json:"status_mfa" bson:"status_mfa"`
	Secret         string    `json:"secret" bson:"secret"`
	Codes          []string  `json:"codes" bson:"codes"`
	UserData       `bson:",inline"`
	Password       UserPassword `bson:",inline"`
}

type UserData struct {
	Name     string `json:"name" validate:"required,name"`
	Email    string `json:"email" bson:",omitempty" validate:"required,email"`
	Username string `json:"username" bson:",omitempty" validate:"required,username"`
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
	p.Hash, err = password.Hash(p.Plain)

	return p, err
}

// Compare reports whether a plain password matches with hash.
//
// For compatibility purposes, it can compare using both SHA256 and bcrypt algorithms.
// Hashes starting with "$" are assumed to be a bcrypt hash; otherwise, they are treated as
// SHA256 hashes.
func (p *UserPassword) Compare(plain string) bool {
	return password.Compare(plain, p.Hash)
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
	Token  string `json:"token"`
	User   string `json:"user"`
	Name   string `json:"name"`
	ID     string `json:"id"`
	Tenant string `json:"tenant"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	MFA    MFA    `json:"mfa" bson:"mfa"`
}

type UserAuthClaims struct {
	Username             string `json:"name"`
	Admin                bool   `json:"admin"`
	Tenant               string `json:"tenant"`
	ID                   string `json:"id"`
	Role                 string `json:"role"`
	AuthClaims           `mapstruct:",squash"`
	MFA                  MFA `json:"mfa"`
	jwt.RegisteredClaims `mapstruct:",squash"`
}

// WithDefaults fill itself with default JWT attributes. Returns itself.
func (u *UserAuthClaims) WithDefaults() *UserAuthClaims {
	now := clock.Now()

	u.RegisteredClaims.ID = uuid.Generate()
	// u.RegisteredClaims.Issuer = "" // TODO: how can we get the correct issuer?
	u.RegisteredClaims.IssuedAt = jwt.NewNumericDate(now)
	u.RegisteredClaims.NotBefore = jwt.NewNumericDate(now)
	u.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Hour * 72))

	return u
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
	LastLogin time.Time `bson:"last_login,omitempty"`
	Name      string    `bson:"name,omitempty"`
	Email     string    `bson:"email,omitempty"`
	Username  string    `bson:"username,omitempty"`
	Password  string    `bson:"password,omitempty"`
	Confirmed *bool     `bson:"confirmed,omitempty"`
}

// UserConflicts holds user attributes that must be unique for each itam and can be utilized in queries
// to identify conflicts.
type UserConflicts struct {
	Email    string
	Username string
}
