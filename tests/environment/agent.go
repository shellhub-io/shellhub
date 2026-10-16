package environment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	composeapi "github.com/docker/compose/v5/pkg/api"
	composeservice "github.com/docker/compose/v5/pkg/compose"
	"github.com/shellhub-io/shellhub/pkg/testimage"
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
// of latest. Tags are the Go build tags the agent is compiled with, such as docker for an agent
// that runs as a container and replaces itself on update. Output receives BuildKit's plain
// progress; nil discards it.
type AgentBuild struct {
	Repository string
	Context    string
	Dockerfile string
	Version    string
	Tags       []string
	Output     io.Writer
}

// BuildAgentImage builds the agent the e2e tests start, accepting [AgentUsername] and
// [AgentPassword], labelled with run and tagged with the run's ID in build's repository, followed
// by -<version> when build sets a Version and by its build tags joined with dashes when it sets
// Tags, so neither replaces the default agent. It builds through compose, as the stack images are
// built, so it runs on the same builder with the same Go cache mounts, and pulls base images with
// the registry credentials of the Docker CLI config. With SHELLHUB_E2E_IMAGE_REPOSITORY set, an
// agent built from the repository root with the default repository and Dockerfile is first pulled
// from there under a key covering its sources and build, and built only when no image carries
// that key, reading its layer cache from the repository's cache-agent tag; with
// SHELLHUB_E2E_IMAGE_PUSH=true as well, the agent it built and its layer cache are pushed there. A
// pull that fails counts as a miss, and a layer cache it cannot read makes the build slower without
// failing it. It returns the image reference, repository:tag, or an error when the working
// directory or the build context cannot be resolved, no versions.env is found at or above the
// working directory, the file cannot be read or has a line that is not KEY=VALUE, the Docker host
// cannot be resolved, the Docker CLI cannot be set up, the compose service cannot be created, a
// build step fails, carrying the builder's message, git cannot read the agent's sources, or the
// push fails. Cancelling ctx aborts the build and returns an error.
func BuildAgentImage(ctx context.Context, run *Run, build AgentBuild) (string, error) {
	if build.Repository == "" {
		build.Repository = defaultAgentRepository
	}

	if build.Dockerfile == "" {
		build.Dockerfile = defaultAgentDockerfile
	}

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

	if len(build.Tags) > 0 {
		goTags := strings.Join(build.Tags, ",")
		tag += "-" + strings.Join(build.Tags, "-")
		buildArgs["GO_TAGS"] = &goTags
	}

	image := build.Repository + ":" + tag

	cache := imageCacheFromEnv()

	agent := stackImage{service: "agent", local: image}

	var cacheRef string

	reusable, err := reusableAgent(cache, root, build)
	if err != nil {
		return "", err
	}

	if reusable {
		key, err := agentImageKey(ctx, root, build)
		if err != nil {
			return "", fmt.Errorf("computing the agent image's key: %w", err)
		}

		agent.remote = cache.repository + ":agent-" + key
		cacheRef = cache.repository + ":cache-agent"
	}

	if err := cache.pullOrBuild(ctx, run, []stackImage{agent}, func(ctx context.Context, _ []string) error {
		return composeBuildAgent(ctx, run, build, image, buildArgs, cacheRef, cache.push)
	}); err != nil {
		return "", err
	}

	return image, nil
}

func reusableAgent(cache imageCache, root string, build AgentBuild) (bool, error) {
	if cache.repository == "" || build.Repository != defaultAgentRepository || build.Dockerfile != defaultAgentDockerfile {
		return false, nil
	}

	contextDir, err := filepath.Abs(build.Context)
	if err != nil {
		return false, err
	}

	return contextDir == root, nil
}

func agentImageKey(ctx context.Context, repoRoot string, build AgentBuild) (string, error) {
	inputs, err := sourceTrees(ctx, repoRoot, "agent")
	if err != nil {
		return "", err
	}

	inputs.WriteString("version " + build.Version + "\ntags " + strings.Join(build.Tags, ",") + "\n")
	inputs.WriteString("user " + AgentUsername + "\npassword " + AgentPassword + "\n")

	sum := sha256.Sum256([]byte(inputs.String()))

	return hex.EncodeToString(sum[:]), nil
}

func composeBuildAgent(ctx context.Context, run *Run, build AgentBuild, image string, args map[string]*string, cacheRef string, push bool) error {
	contextDir, err := filepath.Abs(build.Context)
	if err != nil {
		return err
	}

	if err := pinDockerHost(); err != nil {
		return err
	}

	dockerCli, err := newDockerCli()
	if err != nil {
		return err
	}

	defer dockerCli.Client().Close() //nolint:errcheck // the agent is built or the build failed; closing the client changes neither

	output := build.Output
	if output == nil {
		output = io.Discard
	}

	service, err := composeservice.NewComposeService(dockerCli, composeservice.WithStreams(output, output, nil))
	if err != nil {
		return err
	}

	config := &types.BuildConfig{
		Context:    contextDir,
		Dockerfile: build.Dockerfile,
		Args:       types.MappingWithEquals(args),
		Labels:     types.Labels{runLabel: run.ID()},
	}

	if cacheRef != "" {
		config.CacheFrom = types.StringList{"type=registry,ref=" + cacheRef}
		if push {
			config.CacheTo = types.StringList{"type=registry,ref=" + cacheRef + ",mode=max"}
		}
	}

	project := &types.Project{
		Name:       "shellhub-e2e-agent-" + run.ID(),
		WorkingDir: contextDir,
		Services:   types.Services{"agent": {Name: "agent", Image: image, Build: config}},
	}

	if err := service.Build(ctx, project, composeapi.BuildOptions{Services: []string{"agent"}, Progress: "plain", Out: output}); err != nil {
		return fmt.Errorf("build image: %w", err)
	}

	return nil
}
