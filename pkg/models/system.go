package models

// System is the instance-wide configuration, of which exactly one row exists. It is read on paths
// that must work before any namespace exists, such as setup and login.
type System struct {
	Setup bool `json:"setup"`
	// InstanceTenantID binds the instance to its namespace in single-namespace (Community)
	// deployments. When set, the store refuses any further namespace creation. Enterprise/Cloud
	// leave it empty (the store wrapper strips it) to keep multi-tenant behavior.
	InstanceTenantID string `json:"instance_tenant_id"`
	// Authentication manages the settings for available authentication methods.
	Authentication *SystemAuthentication `json:"authentication"`
}

// SystemAuthentication groups the authentication methods an instance offers. A nil method is
// disabled, which is why each is a pointer.
type SystemAuthentication struct {
	Local *SystemAuthenticationLocal `json:"local"`
}

// SystemAuthenticationLocal configures username-and-password authentication.
type SystemAuthenticationLocal struct {
	// Enabled indicates whether manual authentication using a username and password is enabled or
	// not.
	Enabled bool `json:"enabled" bool:"enabled"`
}
