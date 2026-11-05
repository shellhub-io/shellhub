package requests

// GetSystemInfo is the request behind the unauthenticated info endpoint. The forwarded headers are
// how the server learns the address clients reach it on, which is what it echoes back as its
// endpoints.
type GetSystemInfo struct {
	Host string `header:"X-Forwarded-Host" json:"-"`
	Port int    `header:"X-Forwarded-Port" json:"-"`
}

// SystemInstallScript is the request behind the install script an operator pipes into a shell. The
// forwarded headers become the server address baked into the script, and the query parameters
// become the agent's flags. A query parameter outside the shape its name implies is rejected
// rather than rendered, so the caller gets a 400 instead of a script.
type SystemInstallScript struct {
	Host              string `header:"X-Forwarded-Host" json:"-"`
	Scheme            string `header:"X-Forwarded-Proto" json:"-"`
	ForwardedPort     string `header:"X-Forwarded-Port" json:"-"`
	TenantID          string `query:"tenant_id" json:"-" validate:"omitempty,uuid"`
	PreferredHostname string `query:"preferred_hostname" json:"-" validate:"omitempty,device_name"`
	PreferredIdentity string `query:"preferred_identity" json:"-" validate:"omitempty,device_identity"`
}
