package requests

// GetSystemInfo is the request behind the unauthenticated info endpoint. The forwarded headers are
// how the server learns the address clients reach it on, which is what it echoes back as its
// endpoints.
type GetSystemInfo struct {
	Host string `header:"X-Forwarded-Host"`
	Port int    `header:"X-Forwarded-Port"`
}

// SystemInstallScript is the request behind the install script an operator pipes into a shell. The
// forwarded headers become the server address baked into the script, and the query parameters
// become the agent's flags. A query parameter outside the shape its name implies is rejected
// rather than rendered, so the caller gets a 400 instead of a script.
type SystemInstallScript struct {
	Host              string `header:"X-Forwarded-Host"`
	Scheme            string `header:"X-Forwarded-Proto"`
	ForwardedPort     string `header:"X-Forwarded-Port"`
	TenantID          string `query:"tenant_id" validate:"omitempty,uuid"`
	PreferredHostname string `query:"preferred_hostname" validate:"omitempty,device_name"`
	PreferredIdentity string `query:"preferred_identity" validate:"omitempty,device_identity"`
}
