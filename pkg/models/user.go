package models

import (
	jwt "github.com/dgrijalva/jwt-go"
)

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email" validate:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
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
}

type UserAuthClaims struct {
	Username string `json:"name"`
	Admin    bool   `json:"admin"`
	Tenant   string `json:"tenant"`

	AuthClaims         `json:",squash"`
	jwt.StandardClaims `json:",squash"`
}
