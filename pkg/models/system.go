package models

type System struct {
	Setup bool `json:"setup"`
	// Authentication manages the settings for available authentication methods, such as manual
	// username/password authentication and SAML authentication. Each authentication method
	// can be individually enabled or disabled.
	Authentication *SystemAuthentication `json:"authentication" bson:"authentication"`
}

type SystemAuthentication struct {
	Local *SystemAuthenticationLocal `json:"manual" bson:"manual"`
}

type SystemAuthenticationLocal struct {
	// Enabled indicates whether manual authentication using a username and password is enabled or
	// not.
	Enabled bool `json:"enabled" bool:"enabled"`
}
