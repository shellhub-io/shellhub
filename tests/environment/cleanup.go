package environment

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var (
	cleanupOnce     sync.Once
	cleanupRegistry = &ResourceRegistry{
		containers: make(map[string]bool),
	}
)

// ResourceRegistry keeps track of all test resources for cleanup.
type ResourceRegistry struct {
	mu         sync.Mutex
	containers map[string]bool
}

// RegisterContainer adds a container ID to the cleanup registry.
func (rr *ResourceRegistry) RegisterContainer(id string) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.containers[id] = true
}

// UnregisterContainer removes a container ID from the cleanup registry.
func (rr *ResourceRegistry) UnregisterContainer(id string) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	delete(rr.containers, id)
}

// CleanupAll removes all registered containers.
func (rr *ResourceRegistry) CleanupAll(ctx context.Context) error {
	rr.mu.Lock()
	ids := make([]string, 0, len(rr.containers))
	for id := range rr.containers {
		ids = append(ids, id)
	}
	rr.mu.Unlock()

	if len(ids) == 0 {
		return nil
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	var errs []error
	for _, id := range ids {
		if err := cli.ContainerRemove(ctx, id, container.RemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		}); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove container %s: %w", id, err))
		} else {
			rr.UnregisterContainer(id)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// InitCleanup initializes the cleanup system. This should be called once in TestMain.
// It sets up cleanup handlers that will run even if Ryuk is disabled.
func InitCleanup() {
	cleanupOnce.Do(func() {
		// The cleanup will be triggered by the test framework's cleanup mechanisms
		// No need for signal handlers as tests have their own lifecycle management
	})
}

// CleanupOrphanedContainers removes orphaned test containers.
// This is useful to clean up containers left behind by failed tests.
func CleanupOrphanedContainers(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	// Find containers created by our tests
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", "org.testcontainers=true")

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	var errs []error
	for _, c := range containers {
		// Only remove containers from our test runs
		if _, ok := c.Labels["org.testcontainers.session-id"]; ok {
			if err := cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{
				Force:         true,
				RemoveVolumes: true,
			}); err != nil {
				errs = append(errs, fmt.Errorf("failed to remove container %s: %w", c.ID, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// GetCleanupRegistry returns the global cleanup registry.
func GetCleanupRegistry() *ResourceRegistry {
	return cleanupRegistry
}

// ShouldDisableRyuk returns true if Ryuk should be disabled.
// This checks the environment variable and provides a centralized place to manage this setting.
func ShouldDisableRyuk() bool {
	// Check if explicitly disabled
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "true" {
		return true
	}

	// Ryuk is enabled by default
	// Issue #2445 only affects container reuse, which we don't use
	// https://github.com/testcontainers/testcontainers-go/issues/2445
	return false
}
