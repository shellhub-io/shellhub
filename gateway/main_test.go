package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMain_smoke(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "..",
			Dockerfile: "gateway/Dockerfile",
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
	assert.NoError(t, err)

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}()

	host, err := container.Host(ctx)
	assert.NoError(t, err)

	port, err := container.MappedPort(ctx, "80")
	assert.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", host, port.Port())

	t.Logf("gateway container listening at %s", baseURL)

	healthURL := baseURL + "/healthcheck"

	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(healthURL)
	assert.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
