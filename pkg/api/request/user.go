// Package requests defines structures for the requests' bodies.
package request

// UserDataUpdate is the structure for the request body for the update user data endpoint.
type UserDataUpdate struct {
	// ID is the user's ID.
	ID string `param:"id" validate:"required"`
	// TODO: add validation rules.
	// Name is the user's name.
	Name string `json:"name" validate:"required"`
	// Username is the user's username.
	Username string `json:"username" validate:"required"`
	// Email is the user's email.
	Email string `json:"email" validate:"required"`
}

// UserPasswordUpdate is the structure for the request body for the update user password endpoint.
type UserPasswordUpdate struct {
	// ID is the user's ID.
	ID string `param:"id" validate:"required"`
	// CurrentPassword is the user's current password.
	CurrentPassword string `json:"current_password" validate:"required,min=5,max=30,nefield=NewPassword"`
	// NewPassword is the user's new password.
	NewPassword string `json:"new_password" validate:"required,min=5,max=30,nefield=CurrentPassword"`
}
