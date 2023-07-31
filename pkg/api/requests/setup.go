package requests

//go:generate structsnapshot Setup
type Setup struct {
	Email     string `json:"email" validate:"required,email"`
	Name      string `json:"name" validate:"required"`
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required,min=5,max=30"`
	Namespace string `json:"namespace" validate:"required,hostname_rfc1123,excludes=."`
}
