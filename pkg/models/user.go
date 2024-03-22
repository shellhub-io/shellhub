package models

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
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

// HashUserPassword creates a new [UserPassword] and hashes it.
func HashUserPassword(password string) UserPassword {
	model := UserPassword{
		Plain: password,
	}

	model.hash()

	return model
}

// Hash hashes the plain password.
func (p *UserPassword) hash() {
	sum := sha256.Sum256([]byte(p.Plain))
	p.Hash = hex.EncodeToString(sum[:])
}

// Compare the hashed password with the parameter.
//
// The compared password must be hashed.
func (p *UserPassword) Compare(password UserPassword) bool {
	return password.Hash == p.Hash
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

func (u *UserAuthClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	u.RegisteredClaims = claims
}

type UserTokenRecover struct {
	Token     string    `json:"uid"`
	User      string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
