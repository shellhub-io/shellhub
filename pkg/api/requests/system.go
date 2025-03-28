package requests

type GetSystemInfo struct {
	Host string `header:"X-Forwarded-Host"`
	Port int    `header:"X-Forwarded-Port"`
}

type SystemInstallScript struct {
	Host                string `header:"X-Forwarded-Host"`
	Scheme              string `header:"X-Forwarded-Proto"`
	ForwardedPort       string `header:"X-Forwarded-Port"`
	TenantID            string `query:"tenant_id"`
	KeepAliveInternavel string `query:"keepalive_interval"`
	PreferredHostname   string `query:"preferred_hostname"`
	PreferredIdentity   string `query:"preferred_identity"`
}
