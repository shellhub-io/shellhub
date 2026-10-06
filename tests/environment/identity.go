package environment

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ExpireSSHIdentity moves the expiry of the SSH identity whose key has fingerprint one minute into
// the past, failing t unless exactly that identity changed. The API sets an expiry only in whole
// days ahead, so it writes the row directly, standing in for the day a real key waits to expire.
func (dc *DockerCompose) ExpireSSHIdentity(t *testing.T, fingerprint string) {
	t.Helper()

	dc.updateOne(t,
		"UPDATE ssh_identities SET expires_at = now() - interval '1 minute' WHERE fingerprint = :'fingerprint'",
		map[string]string{"fingerprint": fingerprint})
}

// AgeSSHIdentityLastUse sets the last use of the SSH identity whose key has fingerprint to age ago,
// failing t unless exactly that identity changed. A connection stamps it with the server's clock,
// so a test that wants to see the next stamp move moves the previous one out of the way first.
func (dc *DockerCompose) AgeSSHIdentityLastUse(t *testing.T, fingerprint string, age time.Duration) {
	t.Helper()

	dc.updateOne(t,
		"UPDATE ssh_identities SET last_used_at = now() - :'age'::interval WHERE fingerprint = :'fingerprint'",
		map[string]string{"fingerprint": fingerprint, "age": interval(age)})
}

// DeleteAccessPolicy deletes the access policy id from the namespace the client is authenticated
// against, failing t unless the server answers 200.
func (dc *DockerCompose) DeleteAccessPolicy(t *testing.T, id string) {
	t.Helper()

	resp, err := dc.R(context.WithoutCancel(t.Context())).Delete("/api/access-policies/" + id)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())
}

func (dc *DockerCompose) updateOne(t *testing.T, statement string, vars map[string]string) {
	t.Helper()

	output, err := dc.stack.SQL(t.Context(), statement, vars)
	require.NoError(t, err)
	require.Regexp(t, `(?m)^UPDATE 1$`, output)
}

func interval(d time.Duration) string {
	return strconv.Itoa(int(d.Seconds())) + " seconds"
}
