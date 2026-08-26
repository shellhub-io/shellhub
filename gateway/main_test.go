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

	// The one route the proxy answers by itself. Everything else proxies to an
	// upstream that does not exist beside a lone container, so asking for it
	// would assert on how a missing backend is reported rather than on whether
	// this image boots and serves the configuration it generated.
	//
	// 404 is what distinguishes it: an unmatched path falls through to the
	// catch-all and reaches for the console, which is not here, so a
	// configuration that failed to render this block would answer 502.
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

		resp, err = client.Get(baseURL + "/internal/whatever")
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

	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
