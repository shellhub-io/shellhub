package main

import (
	"context"
	"fmt"
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

	baseURL := fmt.Sprintf("http://%s:%s", host, port.Port())

	t.Logf("gateway container listening at %s", baseURL)

	healthURL := baseURL + "/healthcheck"

	client := http.Client{Timeout: 5 * time.Second}

	const maxRetries = 10

	var resp *http.Response

	for attempt := 1; attempt <= maxRetries; attempt++ {
		t.Logf("healthcheck attempt %d/%d: GET %s", attempt, maxRetries, healthURL)

		var err error

		resp, err = client.Get(healthURL)
		if err == nil {
			break
		}

		t.Logf("healthcheck attempt %d/%d failed: %v", attempt, maxRetries, err)

		if attempt == maxRetries {
			t.Fatalf("healthcheck: all %d attempts exhausted, last error: %v", maxRetries, err)
		}

		delay := time.Duration(attempt) * time.Second
		if delay > 5*time.Second {
			delay = 5 * time.Second
		}

		time.Sleep(delay)
	}

	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
