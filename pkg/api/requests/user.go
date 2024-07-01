// Package requests defines structures to represent requests' bodies from API.
package requests

import "github.com/shellhub-io/shellhub/pkg/models"

type UserParam struct {
	ID string `param:"id" validate:"required"`
}

// UpdateUser is the structure to represent the request body of the update user data endpoint.
type UpdateUser struct {
	Name          string `json:"name" validate:"omitempty,name"`
	Username      string `json:"username" validate:"omitempty,username"`
	Email         string `json:"email" validate:"omitempty,email"`
	RecoveryEmail string `json:"recovery_email" validate:"omitempty,email"`
	// Password is the new password. If not empty, [UserDataUpdate.CurrentPassword] must be the current user's password.
	Password        string `json:"password" validate:"omitempty,password"`
	CurrentPassword string `json:"current_password"`
}

// UserPasswordUpdate is the structure to represent the request body for the update user password endpoint.
type UserPasswordUpdate struct {
	UserParam
	CurrentPassword string `json:"current_password" validate:"required,min=5,max=32,nefield=NewPassword"`
	NewPassword     string `json:"new_password" validate:"required,password,nefield=CurrentPassword"`
}

// UserAuth is the structure to represent the request body for the user auth endpoint.
type UserAuth struct {
	// Identifier represents an username or email.
	//
	// TODO: change json tag from username to identifier and update the OpenAPI.
	Identifier models.UserAuthIdentifier `json:"username" validate:"required"`
	Password   string                    `json:"password" validate:"required"`
}
