package requests

// Setup is the request that initialises a fresh instance: it creates the first user and their
// namespace in one step. It is refused once an instance is set up, so it runs at most once.
type Setup struct {
	Email     string `json:"email" validate:"required,email"`
	Name      string `json:"name" validate:"required,name"`
	Username  string `json:"username" validate:"required,username"`
	Password  string `json:"password" validate:"required,password"`
	Namespace string `json:"namespace" validate:"required,hostname_rfc1123,excludes=."`
}
