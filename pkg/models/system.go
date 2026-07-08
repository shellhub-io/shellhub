package models

type System struct {
	Setup bool `json:"setup"`
	// InstanceTenantID binds the instance to its namespace in single-namespace (Community)
	// deployments. When set, the store refuses any further namespace creation. Enterprise/Cloud
	// leave it empty (the store wrapper strips it) to keep multi-tenant behavior.
	InstanceTenantID string `json:"instance_tenant_id"`
	// Authentication manages the settings for available authentication methods.
	Authentication *SystemAuthentication `json:"authentication"`
}

type SystemAuthentication struct {
	Local *SystemAuthenticationLocal `json:"local"`
}

type SystemAuthenticationLocal struct {
	// Enabled indicates whether manual authentication using a username and password is enabled or
	// not.
	Enabled bool `json:"enabled" bool:"enabled"`
}
