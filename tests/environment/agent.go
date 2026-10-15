package environment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/docker/cli/cli/config"
	controlapi "github.com/moby/buildkit/api/services/control"
	buildkit "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/progress/progressui"
	buildapi "github.com/moby/moby/api/types/build"
	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/client"
	"github.com/shellhub-io/shellhub/pkg/testimage"
	tc "github.com/testcontainers/testcontainers-go"
)

const (
	defaultAgentRepository = "shellhub-e2e/agent"
	defaultAgentDockerfile = "agent/Dockerfile.test"

	buildKitTraceID = "moby.buildkit.trace"

	// AgentUsername is the user the agent built by [BuildAgentImage] accepts logins for.
	AgentUsername = "root"
	// AgentPassword is the password of [AgentUsername] in the agent [BuildAgentImage] builds.
	AgentPassword = "password"
)

// AgentBuild describes the agent image [BuildAgentImage] builds. Context is the build context
// and is required. Repository defaults to shellhub-e2e/agent and Dockerfile, relative to
// Context, to agent/Dockerfile.test. Version, when set, is the version the agent reports instead
// of latest. Output receives BuildKit's plain progress; nil discards it.
type AgentBuild struct {
	Repository string
	Context    string
	Dockerfile string
	Version    string
	Output     io.Writer
}

// BuildAgentImage builds the agent the e2e tests start with BuildKit, through the run's Docker
// client, accepting [AgentUsername] and [AgentPassword], labelled with run and tagged with the
// run's ID in build's repository, followed by -<version> when build sets a Version, so a versioned
// agent never replaces the default one. Base images are pulled with the registry credentials of
// the Docker CLI config, as docker build pulls them. It returns the image reference,
// repository:tag, or an error when the working directory cannot be resolved, no versions.env is
// found at or above it, the file cannot be read or has a line that is not KEY=VALUE, the build
// context or its .dockerignore cannot be read, the daemon refuses the build, the build's output
// cannot be read or decoded, or a build step fails, carrying the daemon's message. When the build
// fails and the session could not connect to the daemon, the error carries that failure too. A
// write to Output that fails is dropped and does not fail the build. Cancelling ctx aborts the
// build, whether the daemon is starting it or streaming it, and returns an error.
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

	request := tc.ContainerRequest{FromDockerfile: tc.FromDockerfile{Context: build.Context, Dockerfile: build.Dockerfile}}

	buildContext, err := request.GetContext()
	if err != nil {
		return "", fmt.Errorf("archive build context: %w", err)
	}

	if archive, ok := buildContext.(io.Closer); ok {
		defer archive.Close() //nolint:errcheck // the daemon has read the context by the time the build ends
	}

	image := build.Repository + ":" + tag

	if err := buildInSession(ctx, run.client, buildContext, client.ImageBuildOptions{
		Version:    buildapi.BuilderBuildKit,
		Dockerfile: build.Dockerfile,
		Tags:       []string{image},
		BuildArgs:  buildArgs,
		Labels:     map[string]string{runLabel: run.ID()},
	}, build.Output); err != nil {
		return "", err
	}

	return image, nil
}

func buildInSession(ctx context.Context, cli *client.Client, buildContext io.Reader, options client.ImageBuildOptions, output io.Writer) error {
	buildSession, err := session.NewSession(ctx, "")
	if err != nil {
		return fmt.Errorf("new build session: %w", err)
	}

	buildSession.Allow(authprovider.NewDockerAuthProvider(authprovider.DockerAuthProviderConfig{
		AuthConfigProvider: authprovider.LoadAuthConfig(config.LoadDefaultConfigFile(io.Discard)),
	}))

	sessionEnded := make(chan error, 1)

	go func() {
		sessionEnded <- buildSession.Run(ctx, func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
			return cli.DialHijack(ctx, "/session", proto, meta)
		})
	}()

	options.SessionID = buildSession.ID()
	buildErr := streamBuild(ctx, cli, buildContext, options, output)

	_ = buildSession.Close()

	if sessionErr := <-sessionEnded; buildErr != nil && sessionErr != nil {
		return errors.Join(buildErr, fmt.Errorf("build session: %w", sessionErr))
	}

	return buildErr
}

func streamBuild(ctx context.Context, cli *client.Client, buildContext io.Reader, options client.ImageBuildOptions, output io.Writer) error {
	result, err := cli.ImageBuild(ctx, buildContext, options)
	if err != nil {
		return fmt.Errorf("build image: %w", err)
	}

	defer result.Body.Close() //nolint:errcheck // the stream is read to its end or the build has failed

	return displayBuild(ctx, result.Body, output)
}

func displayBuild(ctx context.Context, stream io.Reader, output io.Writer) error {
	if output == nil {
		output = io.Discard
	}

	display, err := progressui.NewDisplay(output, progressui.PlainMode)
	if err != nil {
		return fmt.Errorf("build progress display: %w", err)
	}

	statuses := make(chan *buildkit.SolveStatus)
	displayed := make(chan struct{})

	go func() {
		defer close(displayed)

		_, _ = display.UpdateFrom(context.WithoutCancel(ctx), statuses)
	}()

	err = decodeBuildStream(stream, statuses)
	close(statuses)
	<-displayed

	return err
}

func decodeBuildStream(stream io.Reader, statuses chan<- *buildkit.SolveStatus) error {
	decoder := json.NewDecoder(stream)

	for {
		var message jsonstream.Message
		if err := decoder.Decode(&message); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("read build output: %w", err)
		}

		if message.Error != nil {
			return fmt.Errorf("build image: %s", message.Error.Message)
		}

		if message.ID != buildKitTraceID || message.Aux == nil {
			continue
		}

		var trace []byte
		if err := json.Unmarshal(*message.Aux, &trace); err != nil {
			return fmt.Errorf("decode build trace: %w", err)
		}

		var status controlapi.StatusResponse
		if err := status.UnmarshalVT(trace); err != nil {
			return fmt.Errorf("decode build status: %w", err)
		}

		statuses <- buildkit.NewSolveStatus(&status)
	}
}
