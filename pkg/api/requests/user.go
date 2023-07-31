// Package requests defines structures to represent requests' bodies from API.
package requests

//go:generate structsnapshot UserParam
type UserParam struct {
	ID string `param:"id" validate:"required"`
}

// UserDataUpdate is the structure to represent the request body of the update user data endpoint.
//
//go:generate structsnapshot UserDataUpdate
type UserDataUpdate struct {
	UserParam
	Name     string `json:"name" validate:"required"`
	Username string `json:"username" validate:"required,username"`
	Email    string `json:"email" validate:"required"`
}

// UserPasswordUpdate is the structure to represent the request body for the update user password endpoint.
//
//go:generate structsnapshot UserPasswordUpdate
type UserPasswordUpdate struct {
	UserParam
	CurrentPassword string `json:"current_password" validate:"required,min=5,max=30,nefield=NewPassword"`
	NewPassword     string `json:"new_password" validate:"required,min=5,max=30,nefield=CurrentPassword"`
}

// UserAuth is the structure to represent the request body for the user auth endpoint.
//
//go:generate structsnapshot UserAuth
type UserAuth struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
