package environment

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// SetUserPasswordDigest replaces the stored password digest of the user username with digest,
// failing t unless exactly that user changed. Every path that sets a password hashes it with
// bcrypt, so it writes the row directly, standing in for an account created before that.
func (dc *DockerCompose) SetUserPasswordDigest(t *testing.T, username, digest string) {
	t.Helper()

	dc.updateOne(t,
		"UPDATE users SET password_digest = :'digest' WHERE username = :'username'",
		map[string]string{"username": username, "digest": digest})
}

var passwordDigestPattern = regexp.MustCompile(`digest=(\S+)`)

// UserPasswordDigest returns the stored password digest of the user username, reading the row
// directly because no route returns it. It fails t unless psql runs the query and prints a
// non-empty digest, so a missing user and a user with no digest both fail it.
func (dc *DockerCompose) UserPasswordDigest(t *testing.T, username string) string {
	t.Helper()

	output, err := dc.stack.SQL(t.Context(),
		"SELECT 'digest=' || password_digest FROM users WHERE username = :'username'",
		map[string]string{"username": username})
	require.NoError(t, err)

	match := passwordDigestPattern.FindStringSubmatch(output)
	require.NotNil(t, match, "psql printed no digest for %s: %s", username, output)

	return match[1]
}
