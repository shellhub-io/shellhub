package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeviceListRefusesAStatusThePostgresEnumCannotHold drives the real database, which is the
// point: devices.status is an enum, so a value outside it fails the cast rather than matching no
// rows. Before the status parameter declared its values the server answered 500 and reported the
// cast to Sentry, so the assertion that matters is the status code, not the body.
func TestDeviceListRefusesAStatusThePostgresEnumCannotHold(t *testing.T) {
	ctx := context.Background()

	compose := environment.New(t).Up(ctx)
	t.Cleanup(compose.Down)

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace, "")

	auth := compose.AuthUser(t, ShellHubUsername, ShellHubPassword)
	compose.JWT(auth.Token)

	cases := []struct {
		description string
		params      map[string]string
		expected    int
	}{
		{
			description: "refuses a status the enum cannot hold",
			params:      map[string]string{"status": "theprimeagen"},
			expected:    http.StatusBadRequest,
		},
		{
			description: "serves a status the enum holds",
			params:      map[string]string{"status": "accepted"},
			expected:    http.StatusOK,
		},
		{
			description: "serves every status when the parameter is omitted",
			params:      map[string]string{},
			expected:    http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			require.EventuallyWithT(t, func(tt *assert.CollectT) {
				resp, err := compose.R(ctx).SetQueryParams(tc.params).Get("/api/devices")
				assert.NoError(tt, err)
				assert.Equal(tt, tc.expected, resp.StatusCode())
			}, 30*time.Second, 1*time.Second)
		})
	}
}
