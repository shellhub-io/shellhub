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

	// The one route nginx answers by itself. Everything else proxies to an
	// upstream that does not exist beside a lone container, so asking for it
	// would assert on how nginx reports a missing backend rather than on
	// whether this image boots and serves the configuration it generated.
	//
	// Location matters as much as the status: absolute_redirect is off, so the
	// redirects this gateway emits are relative, and a client that compares the
	// header would see it change if that ever stopped being true.
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

		resp, err = client.Get(baseURL + "/v1/")
		if err == nil {
			break
		}

		t.Logf("attempt %d/%d failed: %v", attempt, maxRetries, err)

		if attempt == maxRetries {
			t.Fatalf("all %d attempts exhausted, last error: %v", maxRetries, err)
		}

		delay := time.Duration(attempt) * time.Second
		if delay > 5*time.Second {
			delay = 5 * time.Second
		}

		time.Sleep(delay)
	}

	defer resp.Body.Close()

	assert.Equal(t, http.StatusMovedPermanently, resp.StatusCode)
	assert.Equal(t, "/", resp.Header.Get("Location"))
}
