package environment

import (
	"context"
	"io"
	"log"

	"github.com/moby/moby/client"
	"github.com/shellhub-io/shellhub/pkg/testimage"
	tc "github.com/testcontainers/testcontainers-go"
)

const (
	defaultAgentRepository = "shellhub-e2e/agent"
	defaultAgentDockerfile = "agent/Dockerfile.test"

	// AgentUsername is the user the agent built by [BuildAgentImage] accepts logins for.
	AgentUsername = "root"
	// AgentPassword is the password of [AgentUsername] in the agent [BuildAgentImage] builds.
	AgentPassword = "password"
)

// AgentBuild describes the agent image [BuildAgentImage] builds. Context is the build context
// and is required. Repository defaults to shellhub-e2e/agent and Dockerfile, relative to
// Context, to agent/Dockerfile.test. Version, when set, is the version the agent reports instead
// of latest. Output receives the build log; nil discards it.
type AgentBuild struct {
	Repository string
	Context    string
	Dockerfile string
	Version    string
	Output     io.Writer
}

// BuildAgentImage builds the agent the e2e tests start, accepting [AgentUsername] and
// [AgentPassword], labelled with run and tagged with the run's ID in build's repository, followed
// by -<version> when build sets a Version, so a versioned agent never replaces the default one. It
// talks to the daemon through its API, so it needs no docker CLI. It returns the tag, an error
// when the Docker host cannot be resolved, the provider cannot be created, the working directory
// cannot be resolved, no versions.env is found at or above it, the file cannot be read or has a
// line that is not KEY=VALUE, or the error the daemon reports for the build; cancelling ctx aborts
// the build.
func BuildAgentImage(ctx context.Context, run *Run, build AgentBuild) (string, error) {
	if build.Repository == "" {
		build.Repository = defaultAgentRepository
	}

	if build.Dockerfile == "" {
		build.Dockerfile = defaultAgentDockerfile
	}

	if err := pinDockerHost(); err != nil {
		return "", err
	}

	provider, err := tc.NewDockerProvider(tc.WithLogger(log.New(io.Discard, "", log.LstdFlags)))
	if err != nil {
		return "", err
	}

	defer provider.Close() //nolint:errcheck // the image is built; closing the provider's client changes nothing

	root, err := testimage.Root()
	if err != nil {
		return "", err
	}

	buildArgs, err := testimage.BuildArgs(root)
	if err != nil {
		return "", err
	}

	username, password := AgentUsername, AgentPassword
	tag := run.ID()
	buildArgs["USERNAME"] = &username
	buildArgs["PASSWORD"] = &password

	if build.Version != "" {
		tag += "-" + build.Version
		buildArgs["SHELLHUB_VERSION"] = &build.Version
	}

	return provider.BuildImage(ctx, &tc.ContainerRequest{
		FromDockerfile: tc.FromDockerfile{
			Context:        build.Context,
			Dockerfile:     build.Dockerfile,
			Repo:           build.Repository,
			Tag:            tag,
			BuildLogWriter: build.Output,
			BuildArgs:      buildArgs,
			BuildOptionsModifier: func(options *client.ImageBuildOptions) {
				options.Labels = map[string]string{runLabel: run.ID()}
			},
		},
	})
}
