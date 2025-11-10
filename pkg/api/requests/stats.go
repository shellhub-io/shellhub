package requests

type GetStats struct {
	TenantID string `header:"X-Tenant-ID"`
}
