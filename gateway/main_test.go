package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMain_smoke(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "..",
			Dockerfile: "gateway/Dockerfile",
			Repo:       "gateway",
			Tag:        "smoke",
		},
		ExposedPorts: []string{"80/tcp"},
		Env: map[string]string{
			"SHELLHUB_DOMAIN": "localhost",
		},
		WaitingFor: wait.ForListeningPort("80/tcp").
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}()

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "80")
	require.NoError(t, err)

	baseURL := "http://" + net.JoinHostPort(host, port.Port())

	t.Logf("gateway container listening at %s", baseURL)

	client := http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	const maxRetries = 10

	var resp *http.Response

	for attempt := 1; attempt <= maxRetries; attempt++ {
		var err error

		req, reqErr := http.NewRequestWithContext(t.Context(), http.MethodGet, baseURL+"/internal/whatever", nil)
		require.NoError(t, reqErr)

		req.Host = "localhost"

		resp, err = client.Do(req)
		if err == nil {
			break
		}

		t.Logf("attempt %d/%d failed: %v", attempt, maxRetries, err)

		if attempt == maxRetries {
			t.Fatalf("all %d attempts exhausted, last error: %v", maxRetries, err)
		}

		delay := min(time.Duration(attempt)*time.Second, 5*time.Second)

		time.Sleep(delay)
	}

	defer resp.Body.Close() //nolint:errcheck

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	unknown, err := http.NewRequestWithContext(t.Context(), http.MethodGet, baseURL+"/info", nil)
	require.NoError(t, err)

	unknown.Host = "a.name.this.gateway.does.not.serve"

	refused, err := client.Do(unknown)
	require.NoError(t, err)

	defer refused.Body.Close() //nolint:errcheck

	assert.Equal(t, http.StatusNotFound, refused.StatusCode,
		"an unserved Host must be refused rather than answered with an empty 200, which an agent reads as ShellHub")
}
