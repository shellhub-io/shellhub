package environment

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// InstalledLicense is the license GET /admin/api/license reports: the grant, and where it stands
// against the server's clock.
type InstalledLicense struct {
	License
	Expired       bool `json:"expired"`
	AboutToExpire bool `json:"about_to_expire"`
	GracePeriod   bool `json:"grace_period"`
}

// NewAdmin creates a user with the instance administrator's privileges through the server's
// "admin user create --admin" command, failing t if it fails. Logged in, the user reaches
// /admin/api as well as the namespaces it belongs to.
func (dc *DockerCompose) NewAdmin(t *testing.T, username, email, password string) {
	t.Helper()

	require.NoError(t, dc.stack.Admin(t.Context(), "user", "create", username, password, email, "--admin"))
}

// SignLicense signs license with the issuer of the stack's run, failing t when the run issues no
// licenses or the license cannot be signed.
func (dc *DockerCompose) SignLicense(t *testing.T, license License) []byte {
	t.Helper()

	issuer := dc.stack.run.LicenseIssuer()
	require.NotNil(t, issuer, errRunIssuesNoLicense.Error())

	signed, err := issuer.Sign(license)
	require.NoError(t, err)

	return signed
}

// PostLicense uploads contents as the instance's license through POST /admin/api/license and
// returns the answer whatever its status code. The client must be authenticated as an instance
// administrator. ctx bounds the request. It returns the error only for a request that never got an
// answer.
func (dc *DockerCompose) PostLicense(ctx context.Context, contents []byte) (*resty.Response, error) {
	return dc.R(ctx).
		SetFileReader("file", "license.dat", bytes.NewReader(contents)).
		Post("/admin/api/license")
}

// UseLicense signs license with the run's issuer and installs it through the admin API, failing t
// unless the server answers 200. It takes effect on the server's next license check; installing
// another replaces it. It outlives t's context, so a t.Cleanup can restore a license with it.
func (dc *DockerCompose) UseLicense(t *testing.T, license License) {
	t.Helper()

	resp, err := dc.PostLicense(context.WithoutCancel(t.Context()), dc.SignLicense(t, license))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}

// InstalledLicense returns the license the server holds as GET /admin/api/license reports it,
// failing t unless the server answers 200. The client must be authenticated as an instance
// administrator.
func (dc *DockerCompose) InstalledLicense(t *testing.T) InstalledLicense {
	t.Helper()

	installed := InstalledLicense{}

	resp, err := dc.R(t.Context()).SetResult(&installed).Get("/admin/api/license")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return installed
}

// RemoveLicenses deletes every license the server stored, failing t if the rows cannot be deleted.
// The admin API installs a license but removes none, so it writes the table directly, standing in
// for an instance that was never licensed.
func (dc *DockerCompose) RemoveLicenses(t *testing.T) {
	t.Helper()

	_, err := dc.stack.SQL(t.Context(), "DELETE FROM licenses", nil)
	require.NoError(t, err)
}

var licenseRowsPattern = regexp.MustCompile(`licenses=(\d+)`)

// LicenseRows returns how many licenses the server has stored. Installing one adds a row rather
// than replacing one, so the count tells whether a license was written. It fails t when the count
// cannot be read.
func (dc *DockerCompose) LicenseRows(t *testing.T) int {
	t.Helper()

	output, err := dc.stack.SQL(t.Context(), "SELECT 'licenses=' || count(*) FROM licenses", nil)
	require.NoError(t, err)

	match := licenseRowsPattern.FindStringSubmatch(output)
	require.NotNil(t, match, "psql printed no count: %s", output)

	rows, err := strconv.Atoi(match[1])
	require.NoError(t, err)

	return rows
}

// WriteLicenseFile replaces the contents of the license file the server loads on startup with
// contents, failing t when the stack has no license file, when the file is the one
// SHELLHUB_LICENSE_FILE names rather than one the run issued, or when it cannot be written. It rewrites the
// file rather than replacing it, because the server's container mounts the file itself, not its
// directory. The server reads it on its next start, see [DockerCompose.RestartServer].
func (dc *DockerCompose) WriteLicenseFile(t *testing.T, contents []byte) {
	t.Helper()

	require.NotNil(t, dc.stack.run.LicenseIssuer(), "the license file is SHELLHUB_LICENSE_FILE's, not the run's to rewrite")

	path := dc.stack.envs[licenseFileEnv]
	require.NotEmpty(t, path, "the stack runs without a license file")

	require.NoError(t, os.WriteFile(path, contents, 0o644)) //nolint:gosec // the server container reads the file as whatever user it runs as, and a test license is no secret
}

// RestartServer stops the server and starts it again, failing t unless the API answers again
// within three minutes. It outlives t's context, so a t.Cleanup can bring the server back with it.
func (dc *DockerCompose) RestartServer(t *testing.T) {
	t.Helper()

	dc.restartServer(t)
	require.NoError(t, dc.stack.AwaitAPI(context.WithoutCancel(t.Context())))
}

// RestartFailingServer stops the server and starts it again, then waits until the server has
// exited and its container has given up restarting it, failing t if it is still up after two
// minutes. Call it for a start the server is expected to refuse, and [DockerCompose.RestartServer]
// once the cause is gone.
func (dc *DockerCompose) RestartFailingServer(t *testing.T) {
	t.Helper()

	dc.restartServer(t)

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		state, err := dc.Service(ServiceServer).State(t.Context())
		if !assert.NoError(tt, err) {
			return
		}

		assert.False(tt, state.Running, "the server is still running")
		assert.False(tt, state.Restarting, "the server is still restarting")
	}, 2*time.Minute, time.Second)
}

func (dc *DockerCompose) restartServer(t *testing.T) {
	t.Helper()

	ctx := context.WithoutCancel(t.Context())
	server := dc.Service(ServiceServer)

	timeout := 30 * time.Second
	require.NoError(t, server.Stop(ctx, &timeout))
	require.NoError(t, server.Start(ctx))
}
