package requests

// SystemInstallScript is the request behind the install script an operator pipes into a shell. The
// query parameters become the agent's flags. A query parameter outside the shape its name implies
// is rejected rather than rendered, so the caller gets a 400 instead of a script.
type SystemInstallScript struct {
	TenantID          string `query:"tenant_id" json:"-" validate:"omitempty,uuid"`
	PreferredHostname string `query:"preferred_hostname" json:"-" validate:"omitempty,device_name"`
	PreferredIdentity string `query:"preferred_identity" json:"-" validate:"omitempty,device_identity"`
}
