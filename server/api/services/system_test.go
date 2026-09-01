package services

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/cache"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBuildInstallOverrides(t *testing.T) {
	cases := []struct {
		description string
		req         *requests.SystemInstallScript
		contains    []string
		excludes    []string
	}{
		{
			description: "injects SERVER_ADDRESS from host and forwarded proto",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", Scheme: "https"},
			contains:    []string{"[ -n \"${SERVER_ADDRESS:-}\" ] || SERVER_ADDRESS='https://cloud.example.com'"},
		},
		{
			description: "defaults the scheme to https when not forwarded",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com"},
			contains:    []string{"[ -n \"${SERVER_ADDRESS:-}\" ] || SERVER_ADDRESS='https://cloud.example.com'"},
		},
		{
			description: "appends a non-standard forwarded port",
			req:         &requests.SystemInstallScript{Host: "localhost", ForwardedPort: "8443", Scheme: "https"},
			contains:    []string{"[ -n \"${SERVER_ADDRESS:-}\" ] || SERVER_ADDRESS='https://localhost:8443'"},
		},
		{
			description: "omits the default port for the scheme",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", ForwardedPort: "443", Scheme: "https"},
			contains:    []string{"[ -n \"${SERVER_ADDRESS:-}\" ] || SERVER_ADDRESS='https://cloud.example.com'"},
			excludes:    []string{":443"},
		},
		{
			description: "keeps the http scheme when forwarded",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", Scheme: "http"},
			contains:    []string{"[ -n \"${SERVER_ADDRESS:-}\" ] || SERVER_ADDRESS='http://cloud.example.com'"},
		},
		{
			description: "omits the default http port 80",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", ForwardedPort: "80", Scheme: "http"},
			contains:    []string{"[ -n \"${SERVER_ADDRESS:-}\" ] || SERVER_ADDRESS='http://cloud.example.com'"},
			excludes:    []string{":80"},
		},
		{
			description: "appends a non-standard http port",
			req:         &requests.SystemInstallScript{Host: "localhost", ForwardedPort: "8080", Scheme: "http"},
			contains:    []string{"[ -n \"${SERVER_ADDRESS:-}\" ] || SERVER_ADDRESS='http://localhost:8080'"},
		},
		{
			description: "keeps a port carried inline on the host (direct access, no forwarded port)",
			req:         &requests.SystemInstallScript{Host: "localhost:8080"},
			contains:    []string{"[ -n \"${SERVER_ADDRESS:-}\" ] || SERVER_ADDRESS='https://localhost:8080'"},
		},
		{
			description: "prefers the forwarded port over an inline host port",
			req:         &requests.SystemInstallScript{Host: "localhost:8080", ForwardedPort: "9000", Scheme: "https"},
			contains:    []string{"[ -n \"${SERVER_ADDRESS:-}\" ] || SERVER_ADDRESS='https://localhost:9000'"},
		},
		{
			description: "appends port 443 when scheme is http (non-default for http)",
			req:         &requests.SystemInstallScript{Host: "localhost", ForwardedPort: "443", Scheme: "http"},
			contains:    []string{"[ -n \"${SERVER_ADDRESS:-}\" ] || SERVER_ADDRESS='http://localhost:443'"},
		},
		{
			description: "injects the optional query overrides when present",
			req: &requests.SystemInstallScript{
				Host:              "cloud.example.com",
				Scheme:            "https",
				TenantID:          "00000000-0000-4000-0000-000000000000",
				PreferredHostname: "my-host",
			},
			contains: []string{
				"[ -n \"${TENANT_ID:-}\" ] || TENANT_ID='00000000-0000-4000-0000-000000000000'",
				"[ -n \"${PREFERRED_HOSTNAME:-}\" ] || PREFERRED_HOSTNAME='my-host'",
			},
		},
		{
			description: "omits optional overrides that are absent",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", Scheme: "https"},
			excludes:    []string{"TENANT_ID=", "PREFERRED_HOSTNAME=", "PREFERRED_IDENTITY="},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			out := buildInstallOverrides(tc.req)

			assert.Equal(tt, "\n", out[:1], "the block is spliced onto install.sh's comment line, so it must open with a newline")

			for _, want := range tc.contains {
				assert.Contains(tt, out, want)
			}

			for _, unwanted := range tc.excludes {
				assert.NotContains(tt, out, unwanted)
			}
		})
	}
}

func TestSystemGet(t *testing.T) {
	setupSystem := &models.System{
		Setup: true,
		Authentication: &models.SystemAuthentication{
			Local: &models.SystemAuthenticationLocal{Enabled: true},
		},
	}

	t.Run("serves a cached row without touching the store", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		cacheMock := cachemock.NewMockCache(t)

		cacheMock.
			On("Get", mock.Anything, cache.SystemKey, mock.AnythingOfType("*models.System")).
			Run(func(args mock.Arguments) {
				arg, ok := args.Get(2).(*models.System)
				require.True(t, ok)
				*arg = *setupSystem
			}).
			Return(nil).
			Once()

		system, err := NewService(storeMock, privateKey, publicKey, cacheMock).systemGet(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, setupSystem, system)
	})

	t.Run("reads through on a miss and caches the result", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		cacheMock := cachemock.NewMockCache(t)

		cacheMock.On("Get", mock.Anything, cache.SystemKey, mock.AnythingOfType("*models.System")).Return(nil).Once()
		storeMock.On("SystemGet", mock.Anything).Return(setupSystem, nil).Once()
		cacheMock.On("Set", mock.Anything, cache.SystemKey, setupSystem, systemCacheTTL).Return(nil).Once()

		system, err := NewService(storeMock, privateKey, publicKey, cacheMock).systemGet(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, setupSystem, system)
	})

	t.Run("does not cache before setup completes", func(t *testing.T) {
		pending := &models.System{Setup: false, Authentication: setupSystem.Authentication}

		storeMock := storemock.NewMockStore(t)
		cacheMock := cachemock.NewMockCache(t)

		cacheMock.On("Get", mock.Anything, cache.SystemKey, mock.AnythingOfType("*models.System")).Return(nil).Once()
		storeMock.On("SystemGet", mock.Anything).Return(pending, nil).Once()

		system, err := NewService(storeMock, privateKey, publicKey, cacheMock).systemGet(context.TODO())
		require.NoError(t, err)
		assert.False(t, system.Setup)
	})

	t.Run("ignores a cached row with no authentication", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		cacheMock := cachemock.NewMockCache(t)

		cacheMock.
			On("Get", mock.Anything, cache.SystemKey, mock.AnythingOfType("*models.System")).
			Run(func(args mock.Arguments) {
				arg, ok := args.Get(2).(*models.System)
				require.True(t, ok)
				*arg = models.System{Setup: true}
			}).
			Return(nil).
			Once()
		storeMock.On("SystemGet", mock.Anything).Return(setupSystem, nil).Once()
		cacheMock.On("Set", mock.Anything, cache.SystemKey, setupSystem, systemCacheTTL).Return(nil).Once()

		system, err := NewService(storeMock, privateKey, publicKey, cacheMock).systemGet(context.TODO())
		require.NoError(t, err)
		require.NotNil(t, system.Authentication)
		assert.True(t, system.Authentication.Local.Enabled)
	})

	t.Run("serves the row when the cache is unreachable", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		cacheMock := cachemock.NewMockCache(t)

		cacheMock.
			On("Get", mock.Anything, cache.SystemKey, mock.AnythingOfType("*models.System")).
			Return(errors.New("connection refused", "", 0)).
			Once()
		storeMock.On("SystemGet", mock.Anything).Return(setupSystem, nil).Once()
		cacheMock.
			On("Set", mock.Anything, cache.SystemKey, setupSystem, systemCacheTTL).
			Return(errors.New("connection refused", "", 0)).
			Once()

		system, err := NewService(storeMock, privateKey, publicKey, cacheMock).systemGet(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, setupSystem, system)
	})
}

// TestBuildInstallOverridesIsShellSafe runs the generated prologue through a real shell. The
// overrides are prepended to a script the operator pipes into a root shell, so a value that the
// shell still expands is remote code execution on the machine being enrolled (GHSA-w6jh-83pc-x59g).
func TestBuildInstallOverridesIsShellSafe(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no POSIX shell to run the generated prologue against")
	}

	const probe = `
printf 'SERVER_ADDRESS=%s\036' "${SERVER_ADDRESS:-}"
printf 'TENANT_ID=%s\036' "${TENANT_ID:-}"
printf 'PREFERRED_HOSTNAME=%s\036' "${PREFERRED_HOSTNAME:-}"
printf 'PREFERRED_IDENTITY=%s\036' "${PREFERRED_IDENTITY:-}"
`

	cases := []struct {
		description string
		req         *requests.SystemInstallScript
		want        map[string]string
	}{
		{
			description: "keeps a command substitution in the tenant id unexpanded",
			req: &requests.SystemInstallScript{
				Host: "cloud.example.com", Scheme: "https",
				TenantID: "x$(touch pwned)",
			},
			want: map[string]string{
				"SERVER_ADDRESS": "https://cloud.example.com",
				"TENANT_ID":      "x$(touch pwned)",
			},
		},
		{
			description: "keeps backticks in the preferred hostname unexpanded",
			req: &requests.SystemInstallScript{
				Host: "cloud.example.com", Scheme: "https",
				PreferredHostname: "host`touch pwned`",
			},
			want: map[string]string{
				"SERVER_ADDRESS":     "https://cloud.example.com",
				"PREFERRED_HOSTNAME": "host`touch pwned`",
			},
		},
		{
			description: "keeps a quote break in the preferred identity from opening a statement",
			req: &requests.SystemInstallScript{
				Host: "cloud.example.com", Scheme: "https",
				PreferredIdentity: "a\"\ntouch pwned\n#",
			},
			want: map[string]string{
				"SERVER_ADDRESS":     "https://cloud.example.com",
				"PREFERRED_IDENTITY": "a\"\ntouch pwned\n#",
			},
		},
		{
			description: "keeps a single quote in a value from closing the quoting",
			req: &requests.SystemInstallScript{
				Host: "cloud.example.com", Scheme: "https",
				PreferredHostname: "it's'$(touch pwned)'",
			},
			want: map[string]string{
				"SERVER_ADDRESS":     "https://cloud.example.com",
				"PREFERRED_HOSTNAME": "it's'$(touch pwned)'",
			},
		},
		{
			description: "keeps a command substitution in the forwarded host unexpanded",
			req: &requests.SystemInstallScript{
				Host: "cloud.example.com$(touch pwned)", Scheme: "https",
			},
			want: map[string]string{
				"SERVER_ADDRESS": "https://cloud.example.com$(touch pwned)",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			dir := tt.TempDir()

			//nolint:gosec // G204: running the generated prologue under a shell is the assertion
			cmd := exec.CommandContext(tt.Context(), "sh", "-c", buildInstallOverrides(tc.req)+probe)
			cmd.Dir = dir
			cmd.Env = []string{"PATH=" + os.Getenv("PATH")}

			out, err := cmd.CombinedOutput()
			require.NoError(tt, err, "the generated prologue must be a valid script: %s", out)

			entries, err := os.ReadDir(dir)
			require.NoError(tt, err)
			assert.Empty(tt, entries, "the payload ran: the prologue created files in the working directory")

			assert.Equal(tt, tc.want, parseProbe(string(out)))
		})
	}
}

func parseProbe(out string) map[string]string {
	values := map[string]string{}

	for record := range strings.SplitSeq(strings.TrimSuffix(out, "\x1e"), "\x1e") {
		if name, value, found := strings.Cut(record, "="); found && value != "" {
			values[name] = value
		}
	}

	return values
}

// TestInstallScriptRendersAValidShellScript renders the shipped template with a payload in every
// override and parses the result. The overrides are spliced into a comment line, so a value that
// escapes its quoting does not merely change a variable: it becomes a statement in a script that
// is about to run as root.
func TestInstallScriptRendersAValidShellScript(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no POSIX shell to parse the rendered script with")
	}

	template, err := os.ReadFile("../../../install.sh")
	require.NoError(t, err)

	script := renderInstallScript(string(template), &requests.SystemInstallScript{
		Host:              "cloud.example.com$(touch pwned)",
		Scheme:            "https",
		TenantID:          "x$(touch pwned)",
		PreferredHostname: "host`touch pwned`",
		PreferredIdentity: "a'\"\ntouch pwned\n#",
	})

	require.NotContains(t, script, "{{.Overrides}}", "the marker must be replaced, not left in the served script")

	dir := t.TempDir()
	rendered := filepath.Join(dir, "install.sh")
	//nolint:gosec // G703: rendered is filepath.Join(t.TempDir(), ...), so the path carries no input
	require.NoError(t, os.WriteFile(rendered, []byte(script), 0o600))

	//nolint:gosec // G204: parsing the rendered script under a shell is the assertion
	cmd := exec.CommandContext(t.Context(), "sh", "-n", rendered)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "the rendered script must parse: %s", out)

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "the payload ran while the script was being parsed")
}
