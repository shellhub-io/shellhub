package models

import (
	jwt "github.com/dgrijalva/jwt-go"
)

type User struct {
	ID            string `json:"id" bson:"_id"`
	Name          string `json:"name"`
	Email         string `json:"email" bson:",omitempty" validate:"email"`
	Username      string `json:"username" bson:",omitempty"`
	Password      string `json:"password" bson:",omitempty"`
	TenantID      string `json:"tenant_id" bson:"tenant_id"`
	Devices       int    `json:"devices" bson:",omitempty"`
	Sessions      int    `json:"sessions" bson:",omitempty"`
	SessionRecord bool   `json:"session_record" bson:"session_record,omitempty"`
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
