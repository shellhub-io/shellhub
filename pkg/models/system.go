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
	// InstanceID identifies the instance to the telemetry collector. The store generates it
	// once and never rewrites it, so SystemSet ignores whatever value the caller passes.
	InstanceID string `json:"instance_id"`
	// Telemetry is whether the administrator lets the instance send telemetry. SystemSet
	// ignores it; it changes only through the telemetry setting.
	Telemetry *SystemTelemetry `json:"telemetry"`
}

// SystemTelemetry is the administrator's telemetry setting. The SHELLHUB_TELEMETRY environment
// variable can still force telemetry off when Enabled is true.
type SystemTelemetry struct {
	Enabled bool `json:"enabled"`
}

// SystemAuthentication groups the authentication methods an instance offers. A nil Local is
// disabled.
type SystemAuthentication struct {
	Local *SystemAuthenticationLocal `json:"local"`
}

// SystemAuthenticationLocal configures username-and-password authentication.
type SystemAuthenticationLocal struct {
	// Enabled indicates whether manual authentication using a username and password is enabled or
	// not.
	Enabled bool `json:"enabled" bool:"enabled"`
}
