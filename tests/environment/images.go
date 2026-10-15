package environment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/cli"
	composeapi "github.com/docker/compose/v5/pkg/api"
	composeservice "github.com/docker/compose/v5/pkg/compose"
	"golang.org/x/sync/errgroup"
)

// BuildStackImages builds, at once, the images every stack of run starts from: the community
// server, gateway and UI, and the enterprise server that the enterprise and cloud stacks share. An
// Up after it builds nothing. Without the cloud source in cloudDir (a go.mod) it builds the
// community images only, and an enterprise or cloud Up still fails for the missing source; an
// empty cloudDir is the one Up defaults to. It returns an error when run is nil, and otherwise the
// first error among: checking cloudDir, resolving it to an absolute path, encoding the run
// issuer's key into the enterprise build, reading the edition's env files, resolving the Docker
// host, setting up the Docker CLI, creating the compose service, loading the compose project, and
// the build itself; a failed build writes its output to .stack/build-<edition>.log and names the
// file in the error, or joins the failure to write that file to the build's error. The first
// failure, or ctx ending, cancels the other build. After an error it marks no edition built, so
// every Up builds as if BuildStackImages had not run.
func BuildStackImages(ctx context.Context, run *Run, cloudDir string) error {
	if run == nil {
		return errors.New("building the stack images needs a run to own them")
	}

	if cloudDir == "" {
		cloudDir = defaultCloudDir
	}

	builds := map[Edition][]Edition{EditionCommunity: {EditionCommunity}}

	if err := statCloudSource(cloudDir); err == nil {
		builds[EditionEnterprise] = []Edition{EditionEnterprise, EditionCloud}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	group, ctx := errgroup.WithContext(ctx)

	for edition := range builds {
		group.Go(func() error {
			services := []string{"server", "gateway", "ui"}
			if edition != EditionCommunity {
				services = []string{"server"}
			}

			if err := buildEditionImages(ctx, run, edition, cloudDir, services); err != nil {
				return fmt.Errorf("building the %s images: %w", edition, err)
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	stackImages.Lock()
	defer stackImages.Unlock()

	if stackImages.built == nil {
		stackImages.built = make(map[imageBuild]bool)
	}

	for _, covered := range builds {
		for _, e := range covered {
			stackImages.built[imageBuild{run: run.ID(), edition: e}] = true
		}
	}

	return nil
}

func buildEditionImages(ctx context.Context, run *Run, edition Edition, cloudDir string, services []string) error {
	build, err := newStackBuild(edition, cloudDir, run)
	if err != nil {
		return err
	}

	envs, err := mergeEnvs(build.envFiles, build.envs, run.composeEnvs())
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

	defer dockerCli.Client().Close() //nolint:errcheck // the images are built or the build failed; closing the client changes neither

	var output bytes.Buffer

	service, err := composeservice.NewComposeService(dockerCli, composeservice.WithStreams(&output, &output, nil))
	if err != nil {
		return err
	}

	env := make([]string, 0, len(envs))
	for k, v := range envs {
		env = append(env, k+"="+v)
	}

	project, err := service.LoadProject(ctx, composeapi.ProjectLoadOptions{
		ProjectName:       "shellhub-e2e-build-" + run.ID() + "-" + string(edition),
		ConfigPaths:       append(build.composeFiles, buildComposeFile),
		ProjectOptionsFns: []cli.ProjectOptionsFn{cli.WithEnv(env)},
	})
	if err != nil {
		return err
	}

	if err := service.Build(ctx, project, composeapi.BuildOptions{Services: services}); err != nil {
		return keepBuildLog(err, edition, output.Bytes())
	}

	return nil
}

func keepBuildLog(err error, edition Edition, output []byte) error {
	path := stackLogPath("build-" + string(edition))

	if mkdirErr := os.MkdirAll(filepath.Dir(path), 0o750); mkdirErr != nil {
		return errors.Join(err, mkdirErr)
	}

	if writeErr := os.WriteFile(path, output, 0o600); writeErr != nil {
		return errors.Join(err, writeErr)
	}

	return fmt.Errorf("%w; the build output is in %s", err, path)
}
