package models

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type UserData struct {
	Name     string `json:"name" validate:"required,min=1"`
	Email    string `json:"email" bson:",omitempty" validate:"required,email"`
	Username string `json:"username" bson:",omitempty" validate:"required,min=3,max=30,alphanum,ascii"`
}

type UserPassword struct {
	Password string `json:"password" bson:",omitempty" validate:"required,min=5,max=30"`
}

type User struct {
	ID           string    `json:"id,omitempty" bson:"_id,omitempty"`
	Namespaces   int       `json:"namespaces" bson:"namespaces,omitempty"`
	Confirmed    bool      `json:"confirmed"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	LastLogin    time.Time `json:"last_login" bson:"last_login"`
	UserData     `bson:",inline"`
	UserPassword `bson:",inline"`
}

type UserAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserAuthResponse struct {
	Token  string `json:"token"`
	User   string `json:"user"`
	Name   string `json:"name"`
	ID     string `json:"id"`
	Tenant string `json:"tenant"`
	Role   string `json:"role"`
	Email  string `json:"email"`
}

type UserAuthClaims struct {
	Username string `json:"name"`
	Admin    bool   `json:"admin"`
	Tenant   string `json:"tenant"`
	ID       string `json:"id"`
	Role     string `json:"role"`

	AuthClaims         `mapstruct:",squash"`
	jwt.StandardClaims `mapstruct:",squash"`
}

type UserTokenRecover struct {
	Token     string    `json:"uid"`
	User      string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
