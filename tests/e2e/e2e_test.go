package e2e

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestE2E(t *testing.T) {
	ctx := context.Background()

	// Start ShellHub environment with UI (using e2e compose)
	compose := environment.NewE2E(t).Up(ctx)
	t.Cleanup(compose.Down)

	// Get the base URL for Playwright - use gateway hostname inside Docker network
	baseURL := "http://gateway:80"

	// Run Playwright tests in container
	t.Run("homepage", func(t *testing.T) {
		runPlaywrightTest(t, ctx, compose, baseURL, "tests/homepage.spec.ts")
	})
}

func runPlaywrightTest(t *testing.T, ctx context.Context, compose *environment.DockerCompose, baseURL, testFile string) {
	// Get absolute path to e2e directory
	absPath, err := filepath.Abs(".")
	require.NoError(t, err)

	// Get the ShellHub network name from compose environment
	networkName := compose.Env("SHELLHUB_NETWORK")

	// Create Playwright container that installs deps and runs tests
	req := tc.ContainerRequest{
		Image: "mcr.microsoft.com/playwright:v1.57.0-jammy",
		Cmd: []string{
			"sh", "-c",
			"npm install && npx playwright test " + testFile,
		},
		Env: map[string]string{
			"BASE_URL": baseURL,
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Mounts = []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: absPath,
					Target: "/app",
				},
			}
		},
		WorkingDir: "/app",
		Networks:   []string{networkName},
		WaitingFor: wait.ForExit().WithExitTimeout(5 * time.Minute),
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Wait for completion
	_, err = container.State(ctx)
	require.NoError(t, err)

	// Get and print logs
	logs, err := container.Logs(ctx)
	require.NoError(t, err)

	buf := new(strings.Builder)
	_, _ = io.Copy(buf, logs)
	t.Logf("Playwright output:\n%s", buf.String())

	// Cleanup
	exitCode, err := container.State(ctx)
	require.NoError(t, err)

	_ = container.Terminate(ctx)

	// Check exit code
	require.Equal(t, 0, exitCode.ExitCode, "Playwright tests should pass")
}
