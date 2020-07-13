package models

import (
	jwt "github.com/dgrijalva/jwt-go"
)

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email" bson:",omitempty" validate:"email"`
	Username string `json:"username" bson:",omitempty"`
	Password string `json:"password" bson:",omitempty"`
	TenantID string `json:"tenant_id" bson:"tenant_id"`
}

type UserAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserAuthResponse struct {
	Token  string `json:"token"`
	User   string `json:"user"`
	Name   string `json:"name"`
	Tenant string `json:"tenant"`
	Email  string `json:"email"`
}

type UserAuthClaims struct {
	Username string `json:"name"`
	Admin    bool   `json:"admin"`
	Tenant   string `json:"tenant"`

	AuthClaims         `json:",squash"`
	jwt.StandardClaims `json:",squash"`
}
