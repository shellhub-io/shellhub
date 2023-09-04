package models

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type UserData struct {
	Name     string `json:"name" validate:"required,name"`
	Email    string `json:"email" bson:",omitempty" validate:"required,email"`
	Username string `json:"username" bson:",omitempty" validate:"required,username"`
}

type UserPassword struct {
	// PlainPassword contains the plain text password.
	PlainPassword string `json:"password" bson:"-" validate:"required,password"`
	// HashedPassword contains the hashed pasword from plain text.
	HashedPassword string `json:"-" bson:"password"`
}

// NewUserPassword creates a new [UserPassword] and hashes it.
func NewUserPassword(password string) UserPassword {
	model := UserPassword{
		PlainPassword: password,
	}

	model.Hash()

	return model
}

func (p *UserPassword) hash(string) string {
	sum := sha256.Sum256([]byte(p.PlainPassword))

	return hex.EncodeToString(sum[:])
}

// Hash hashes the plain password.
func (p *UserPassword) Hash() string {
	p.HashedPassword = p.hash(p.PlainPassword)

	return p.HashedPassword
}

// Compare the hashed password with the parameter.
//
// The compared password must be hashed.
func (p *UserPassword) Compare(password UserPassword) bool {
	return password.HashedPassword == p.HashedPassword
}

func (p *UserPassword) String() string {
	return p.HashedPassword
}

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
	UserData       `bson:",inline"`
	UserPassword   `bson:",inline"`
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

type UserAuthRequest struct {
	// Identifier represents an username or email.
	//
	// TODO: change json tag from username to identifier and update the OpenAPI.
	Identifier UserAuthIdentifier `json:"username"`
	Password   string             `json:"password"`
}

type UserAuthResponse struct {
	Token  string `json:"token"`
	User   string `json:"user"`
	Name   string `json:"name"`
	ID     string `json:"id"`
	Tenant string `json:"tenant"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	MFA    bool   `json:"mfa" bson:"mfa"`
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

type UserTokenRecover struct {
	Token     string    `json:"uid"`
	User      string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
