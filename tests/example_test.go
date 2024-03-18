package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
)

// TestExample is a test function that demonstrates how to use the testing integration framework.
func TestExample(t *testing.T) {
	if testing.Short() {
		t.Skip("Skiping \"TestExample\".")
	}

	t.Parallel()

	cases := []struct {
		description string
		version     string
		compose     *environment.DockerCompose
		expected    func(gatewayPort, sshPort string) map[string]interface{}
	}{
		{
			description: "succeeds to retrieve API info",
			version:     "v0.10.0",
			expected: func(httpPort, sshPort string) map[string]interface{} {
				return map[string]interface{}{
					"version": "v0.10.0",
					"endpoints": map[string]interface{}{
						"api": "localhost:" + httpPort,
						"ssh": "localhost:" + sshPort,
					},
				}
			},
		},
	}

	// Create a new [environment.DockerComposeConfigurator]. It can be used as a base
	// environment between cases by cloning it to avoid boilerplate code. You can configure
	// this as you want. For example, if you want all clones to have the
	// "SHELLHUB_ENVIRONMENT" variable set to "development", you can call
	// `WithEnv("SHELLHUB_ENVIRONMENT", "development")` after `New(t)`.
	cfg := environment.New(t)

	for _, tt := range cases {
		tc := tt

		t.Run(tc.description, func(t *testing.T) {
			// Each test case must also run in parallel.
			t.Parallel()
			ctx := context.Background()

			// Clone the base configuration and bring up the environment. Pass the [Down] function to the
			// test cleanup, ensuring that the environment will be cleaned up whether it succeeds or
			// fails.
			dc := cfg.Clone(t).WithEnv("SHELLHUB_VERSION", tc.version).Up(ctx)
			t.Cleanup(dc.Down)

			// Use the [R] helper to make an API call. If you need a user, you can use the [NewUser] method.
			// [NewNamespace] creates a new namespace, and [AuthUser] can be used to retrieve a login token.
			res, err := dc.R(ctx).Get("/info")
			if !assert.NoError(t, err) {
				assert.FailNow(t, err.Error())
			}

			expected := tc.expected(dc.Env("SHELLHUB_HTTP_PORT"), dc.Env("SHELLHUB_SSH_PORT"))
			actual := make(map[string]interface{})
			if !assert.NoError(t, json.Unmarshal(res.Body(), &actual)) {
				assert.FailNow(t, err.Error())
			}

			// Assert that the expected body matches the actual response.
			assert.Equal(t, expected, actual)
		})
	}
}
