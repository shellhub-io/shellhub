package environment

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/shellhub-io/shellhub/pkg/uuid"
	tc "github.com/testcontainers/testcontainers-go"
)

var (
	agentImageOnce sync.Once
	agentImage     string
	agentImageErr  error
)

// AgentImageBuildOptions holds options for building the agent test image.
type AgentImageBuildOptions struct {
	Username string
	Password string
}

// BuildAgentImage builds the agent test image once and caches it for reuse.
// Subsequent calls return the cached image name.
func BuildAgentImage(ctx context.Context, opts AgentImageBuildOptions) (string, error) {
	agentImageOnce.Do(func() {
		repo := "shellhub-agent-test"
		tag := uuid.Generate()

		username := opts.Username
		if username == "" {
			username = DefaultAgentUsername
		}

		password := opts.Password
		if password == "" {
			password = DefaultAgentPassword
		}

		req := tc.ContainerRequest{
			FromDockerfile: tc.FromDockerfile{
				Context:       "../..",
				Dockerfile:    "agent/Dockerfile.test",
				Repo:          repo,
				Tag:           tag,
				KeepImage:     true, // Keep image for reuse
				PrintBuildLog: false,
				BuildArgs: map[string]*string{
					"USERNAME": &username,
					"PASSWORD": &password,
				},
			},
		}

		container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
			ContainerRequest: req,
			Started:          false, // Don't start, just build
			Logger:           log.New(io.Discard, "", log.LstdFlags),
		})
		if err != nil {
			agentImageErr = fmt.Errorf("failed to build agent image: %w", err)

			return
		}

		// Terminate the container immediately since we only needed to build the image
		if container != nil {
			_ = container.Terminate(ctx)
		}

		agentImage = fmt.Sprintf("%s:%s", repo, tag)
	})

	return agentImage, agentImageErr
}

// AgentContainerOptions holds options for creating an agent container.
type AgentContainerOptions struct {
	ServerAddress string
	TenantID      string
	Identity      string
	Username      string
	Password      string
	Networks      []string
	NetworkAlias  string
}

// NewAgentContainer creates a new agent container with the given options.
// It uses a pre-built image from BuildAgentImage to avoid rebuilding on each test.
func NewAgentContainer(ctx context.Context, opts AgentContainerOptions) (tc.Container, error) {
	// Build or get cached image
	image, err := BuildAgentImage(ctx, AgentImageBuildOptions{
		Username: opts.Username,
		Password: opts.Password,
	})
	if err != nil {
		return nil, err
	}

	// Set defaults
	if opts.TenantID == "" {
		opts.TenantID = DefaultNamespace
	}
	if opts.ServerAddress == "" {
		opts.ServerAddress = "http://gateway:80"
	}

	envs := map[string]string{
		"SHELLHUB_SERVER_ADDRESS":     opts.ServerAddress,
		"SHELLHUB_TENANT_ID":          opts.TenantID,
		"SHELLHUB_PRIVATE_KEY":        "/tmp/shellhub.key",
		"SHELLHUB_LOG_FORMAT":         "json",
		"SHELLHUB_KEEPALIVE_INTERVAL": "1",
		"SHELLHUB_LOG_LEVEL":          "trace",
	}

	if opts.Identity != "" {
		envs["SHELLHUB_PREFERRED_IDENTITY"] = opts.Identity
	}

	req := tc.ContainerRequest{
		Image: image,
		Env:   envs,
	}

	// Network configuration
	if len(opts.Networks) > 0 {
		// Use compose network for isolation
		req.Networks = opts.Networks
		if opts.NetworkAlias != "" {
			req.NetworkAliases = map[string][]string{
				opts.Networks[0]: {opts.NetworkAlias},
			}
		}
	} else {
		// Use host network for backward compatibility (less isolated but works with current tests)
		req.NetworkMode = "host"
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          false, // Let caller control when to start
		Logger:           log.New(io.Discard, "", log.LstdFlags),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent container: %w", err)
	}

	return container, nil
}
