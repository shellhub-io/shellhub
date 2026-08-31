package requests

// GetStats is the request for one namespace's dashboard counters.
type GetStats struct {
	TenantID string `header:"X-Tenant-ID"`
}
