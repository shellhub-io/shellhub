package models

type System struct {
	Setup bool `json:"setup"`
	// Authentication manages the settings for available authentication methods.
	Authentication *SystemAuthentication `json:"authentication" bson:"authentication"`
}

type SystemAuthentication struct {
	Local *SystemAuthenticationLocal `json:"local" bson:"local"`
}

type SystemAuthenticationLocal struct {
	// Enabled indicates whether manual authentication using a username and password is enabled or
	// not.
	Enabled bool `json:"enabled" bool:"enabled"`
}
