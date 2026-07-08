package requests

type Setup struct {
	Email     string `json:"email" validate:"required,email"`
	Name      string `json:"name" validate:"required,name"`
	Username  string `json:"username" validate:"required,username"`
	Password  string `json:"password" validate:"required,password"`
	Namespace string `json:"namespace" validate:"required,hostname_rfc1123,excludes=."`
}
