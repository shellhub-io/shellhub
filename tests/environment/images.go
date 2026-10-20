package environment

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config"
	composeapi "github.com/docker/compose/v5/pkg/api"
	composeservice "github.com/docker/compose/v5/pkg/compose"
	"github.com/moby/moby/client"
	"github.com/shellhub-io/shellhub/pkg/testimage"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// BuildStackImages builds, at once, the images every stack of run starts from: the community
// server, gateway and UI, and the enterprise server that the enterprise and cloud stacks share. An
// Up after it builds nothing. Without the cloud source in cloudDir (a go.mod) it builds the
// community images only, and an enterprise or cloud Up still fails for the missing source; an
// empty cloudDir is the one Up defaults to. It reuses and publishes the server and gateway images
// as [BuildEditionImages] does. It returns an error when run is nil, and otherwise the first error
// among: checking cloudDir, resolving it to an absolute path, encoding the run issuer's key into
// the enterprise build, reading the edition's env files, resolving the Docker host, setting up the
// Docker CLI, creating the compose service, loading the compose project, computing an image's
// key, the build itself, and a push; a failed build writes its output to
// .stack/build-<edition>.log and names the file in the error, or joins the failure to write that
// file to the build's error. The first failure, or ctx ending, cancels the other build. After an
// error it marks no edition built, so every Up builds as if BuildStackImages had not run.
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

	for _, covered := range builds {
		markImagesBuilt(run, covered...)
	}

	return nil
}

// BuildEditionImages builds the server, gateway and UI images a stack of edition in run starts
// from, so an Up of that edition after it builds nothing. An empty cloudDir is the one Up defaults
// to. With SHELLHUB_E2E_IMAGE_REPOSITORY set, it pulls the server and gateway images from that one
// repository instead, tagged <image>-<key>, the key covering the build's arguments and the sources
// each copies as committed at HEAD, so uncommitted changes reuse the image of their commit, and
// builds the ones it does not find there; with SHELLHUB_E2E_IMAGE_PUSH=true as well, it pushes each
// image it built there under its tag. With the repository set, the builds also read their layer
// cache from it, tagged cache-<image> and cache-server-<edition> for the server, and with the push
// enabled write it back there, so a build that misses its image still skips the layers its change
// leaves alone; a layer cache it cannot read is a slower build, not an error. It returns an error
// when run is nil, and otherwise the first error among: checking and resolving cloudDir, encoding
// the run issuer's key into the build, reading the edition's env files, resolving the Docker host,
// setting up the Docker CLI, creating the compose service, loading the compose project, computing
// an image's key, the build, which writes its output to .stack/build-<edition>.log as
// [BuildStackImages] does, and a push. ctx ending stops the pull, build or push in progress: a pull
// it stops counts as a miss, and the build or push then returns the error. After an error it marks
// nothing built, so Up builds as if BuildEditionImages had not run.
func BuildEditionImages(ctx context.Context, run *Run, edition Edition, cloudDir string) error {
	if run == nil {
		return errors.New("building the stack images needs a run to own them")
	}

	if cloudDir == "" {
		cloudDir = defaultCloudDir
	}

	if err := buildEditionImages(ctx, run, edition, cloudDir, []string{"server", "gateway", "ui"}); err != nil {
		return fmt.Errorf("building the %s images: %w", edition, err)
	}

	markImagesBuilt(run, edition)

	return nil
}

func markImagesBuilt(run *Run, editions ...Edition) {
	stackImages.Lock()
	defer stackImages.Unlock()

	if stackImages.built == nil {
		stackImages.built = make(map[imageBuild]bool)
	}

	for _, e := range editions {
		stackImages.built[imageBuild{run: run.ID(), edition: e}] = true
	}
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
		ConfigPaths:       append(build.composeFiles, buildComposeFiles()...),
		ProjectOptionsFns: []cli.ProjectOptionsFn{cli.WithEnv(env)},
	})
	if err != nil {
		return err
	}

	cache := imageCacheFromEnv()

	images := make([]stackImage, 0, len(services))
	for _, name := range services {
		image := stackImage{service: name, local: project.Services[name].Image}

		if _, reusable := reusableImageSources[name]; reusable && cache.repository != "" {
			if image.remote, err = cache.reference(ctx, name, edition, cloudDir, build.envs[goLDFlagsEnv]); err != nil {
				return err
			}
		}

		images = append(images, image)
	}

	return cache.pullOrBuild(ctx, run, images, func(ctx context.Context, services []string) error {
		if err := service.Build(ctx, project, composeapi.BuildOptions{Services: services}); err != nil {
			return keepBuildLog(err, edition, output.Bytes())
		}

		return nil
	})
}

const (
	imageRepositoryEnv = "SHELLHUB_E2E_IMAGE_REPOSITORY"
	imagePushEnv       = "SHELLHUB_E2E_IMAGE_PUSH"
)

type imageCache struct {
	repository string
	push       bool
}

func imageCacheFromEnv() imageCache {
	repository := os.Getenv(imageRepositoryEnv)

	return imageCache{repository: repository, push: repository != "" && os.Getenv(imagePushEnv) == "true"}
}

func (c imageCache) reference(ctx context.Context, service string, edition Edition, cloudDir, ldflags string) (string, error) {
	root, err := testimage.Root()
	if err != nil {
		return "", err
	}

	key, err := imageKey(ctx, root, cloudDir, service, edition, ldflags)
	if err != nil {
		return "", fmt.Errorf("computing the %s image's key: %w", service, err)
	}

	return c.repository + ":" + service + "-" + key, nil
}

type stackImage struct {
	service string
	local   string
	remote  string
}

func (c imageCache) pullOrBuild(ctx context.Context, run *Run, images []stackImage, build func(context.Context, []string) error) error {
	var (
		services []string
		built    []stackImage
	)

	for _, image := range images {
		if image.remote != "" {
			err := run.adoptImage(ctx, image)
			if err == nil {
				log.WithFields(log.Fields{"image": image.local, "from": image.remote}).Info("reused the image another run built")

				continue
			}

			log.WithError(err).WithField("image", image.remote).Info("found no image to reuse, building it")
		}

		services = append(services, image.service)
		built = append(built, image)
	}

	if len(services) > 0 {
		if err := build(ctx, services); err != nil {
			return err
		}
	}

	if !c.push {
		return nil
	}

	for _, image := range built {
		if image.remote == "" {
			continue
		}

		if err := run.publishImage(ctx, image); err != nil {
			return err
		}
	}

	return nil
}

func (r *Run) adoptImage(ctx context.Context, image stackImage) error {
	auth, err := registryAuth(image.remote)
	if err != nil {
		return err
	}

	pulled, err := r.client.ImagePull(ctx, image.remote, client.ImagePullOptions{RegistryAuth: auth})
	if err != nil {
		return err
	}

	if err := pulled.Wait(ctx); err != nil {
		return err
	}

	return r.relabel(ctx, image.remote, image.local, r.ID())
}

func (r *Run) publishImage(ctx context.Context, image stackImage) error {
	if err := r.relabel(ctx, image.local, image.remote, ""); err != nil {
		if _, inspectErr := r.client.ImageInspect(ctx, image.remote); inspectErr != nil {
			return err
		}

		log.WithError(err).WithField("image", image.remote).Info("another run labeled the image first, pushing it")
	}

	auth, err := registryAuth(image.remote)
	if err != nil {
		return err
	}

	pushed, err := r.client.ImagePush(ctx, image.remote, client.ImagePushOptions{RegistryAuth: auth})
	if err != nil {
		return fmt.Errorf("push %s: %w", image.remote, err)
	}

	if err := pushed.Wait(ctx); err != nil {
		return fmt.Errorf("push %s: %w", image.remote, err)
	}

	return nil
}

func (r *Run) relabel(ctx context.Context, source, target, run string) error {
	created, err := r.client.ContainerCreate(ctx, client.ContainerCreateOptions{Image: source})
	if err != nil {
		return fmt.Errorf("relabel %s as %s: %w", source, target, err)
	}

	defer r.client.ContainerRemove(context.WithoutCancel(ctx), created.ID, client.ContainerRemoveOptions{Force: true, RemoveVolumes: true}) //nolint:errcheck // the container never ran; one left behind holds only the image's layers and empty volumes

	if _, err := r.client.ContainerCommit(ctx, created.ID, client.ContainerCommitOptions{
		Reference: target,
		Changes:   []string{"LABEL " + runLabel + "=" + run},
	}); err != nil {
		return fmt.Errorf("relabel %s as %s: %w", source, target, err)
	}

	return nil
}

func registryAuth(image string) (string, error) {
	return command.RetrieveAuthTokenFromImage(config.LoadDefaultConfigFile(io.Discard), image)
}

var reusableImageSources = map[string][]string{
	"server": {
		"server/Dockerfile", "docker-compose.test.build.yml", "versions.env", ".dockerignore",
		"go.mod", "go.sum", "pkg", "openapi", "server", "install.sh",
	},
	"gateway": {
		"gateway/Dockerfile", "docker-compose.test.build.yml", "versions.env", ".dockerignore",
		"go.mod", "go.sum", "pkg", "gateway",
	},
	"agent": {
		defaultAgentDockerfile, "versions.env", ".dockerignore",
		"go.mod", "go.sum", "pkg", "agent", "tests/cmd/signal-reporter",
	},
}

func sourceTrees(ctx context.Context, repoRoot, service string) (*strings.Builder, error) {
	paths, ok := reusableImageSources[service]
	if !ok {
		return nil, fmt.Errorf("the %s image is built on every run, never reused", service)
	}

	revisions := make([]string, len(paths))
	for i, path := range paths {
		revisions[i] = "HEAD:" + path
	}

	trees, err := gitRevParse(ctx, repoRoot, revisions...)
	if err != nil {
		return nil, err
	}

	var inputs strings.Builder

	for i, path := range paths {
		inputs.WriteString(path + " " + trees[i] + "\n")
	}

	return &inputs, nil
}

func imageKey(ctx context.Context, repoRoot, cloudDir, service string, edition Edition, ldflags string) (string, error) {
	inputs, err := sourceTrees(ctx, repoRoot, service)
	if err != nil {
		return "", err
	}

	if service == "server" {
		inputs.WriteString("edition " + edition.Build() + "\nldflags " + ldflags + "\n")

		if edition != EditionCommunity {
			cloudTree, err := gitRevParse(ctx, cloudDir, "HEAD^{tree}")
			if err != nil {
				return "", err
			}

			inputs.WriteString("cloud " + cloudTree[0] + "\n")
		}
	}

	sum := sha256.Sum256([]byte(inputs.String()))

	return hex.EncodeToString(sum[:]), nil
}

func gitRevParse(ctx context.Context, dir string, revisions ...string) ([]string, error) {
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir, "rev-parse"}, revisions...)...) //nolint:gosec // runs git; the revisions name HEAD's trees
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git rev-parse in %s: %w: %s", dir, err, bytes.TrimSpace(stderr.Bytes()))
	}

	objects := strings.Fields(string(out))
	if len(objects) != len(revisions) {
		return nil, fmt.Errorf("git rev-parse in %s resolved %d of %d revisions", dir, len(objects), len(revisions))
	}

	return objects, nil
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
