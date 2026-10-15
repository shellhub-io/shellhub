package environment

import "testing"

// ExpireMembershipInvitation moves the expiry of the one pending invitation to the namespace tenant
// one minute into the past, failing t unless exactly that invitation changed. An invitation lasts
// seven days, so the row is written directly, standing in for the week a real link waits to expire.
func (dc *DockerCompose) ExpireMembershipInvitation(t *testing.T, tenant string) {
	t.Helper()

	dc.updateOne(t,
		"UPDATE membership_invitations SET expires_at = now() - interval '1 minute' WHERE tenant_id = :'tenant' AND status = 'pending'",
		map[string]string{"tenant": tenant})
}
