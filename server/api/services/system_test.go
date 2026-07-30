package services

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/stretchr/testify/assert"
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
			contains:    []string{"\nSERVER_ADDRESS=\"${SERVER_ADDRESS:-https://cloud.example.com}\"\n"},
		},
		{
			description: "defaults the scheme to https when not forwarded",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-https://cloud.example.com}\""},
		},
		{
			description: "appends a non-standard forwarded port",
			req:         &requests.SystemInstallScript{Host: "localhost", ForwardedPort: "8443", Scheme: "https"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-https://localhost:8443}\""},
		},
		{
			description: "omits the default port for the scheme",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", ForwardedPort: "443", Scheme: "https"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-https://cloud.example.com}\""},
			excludes:    []string{":443"},
		},
		{
			description: "keeps the http scheme when forwarded",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", Scheme: "http"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-http://cloud.example.com}\""},
		},
		{
			description: "omits the default http port 80",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", ForwardedPort: "80", Scheme: "http"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-http://cloud.example.com}\""},
			excludes:    []string{":80"},
		},
		{
			description: "appends a non-standard http port",
			req:         &requests.SystemInstallScript{Host: "localhost", ForwardedPort: "8080", Scheme: "http"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-http://localhost:8080}\""},
		},
		{
			description: "keeps a port carried inline on the host (direct access, no forwarded port)",
			req:         &requests.SystemInstallScript{Host: "localhost:8080"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-https://localhost:8080}\""},
		},
		{
			description: "prefers the forwarded port over an inline host port",
			req:         &requests.SystemInstallScript{Host: "localhost:8080", ForwardedPort: "9000", Scheme: "https"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-https://localhost:9000}\""},
		},
		{
			description: "appends port 443 when scheme is http (non-default for http)",
			req:         &requests.SystemInstallScript{Host: "localhost", ForwardedPort: "443", Scheme: "http"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-http://localhost:443}\""},
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
				"TENANT_ID=\"${TENANT_ID:-00000000-0000-4000-0000-000000000000}\"",
				"PREFERRED_HOSTNAME=\"${PREFERRED_HOSTNAME:-my-host}\"",
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

			// The marker sits on a comment line, so the block must start with a
			// newline to break onto real assignment lines.
			assert.Equal(tt, "\n", out[:1])

			for _, want := range tc.contains {
				assert.Contains(tt, out, want)
			}

			for _, unwanted := range tc.excludes {
				assert.NotContains(tt, out, unwanted)
			}
		})
	}
}
