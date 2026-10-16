package environment

import (
	"context"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/moby/moby/client"
	"github.com/shellhub-io/shellhub/pkg/testimage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gitFixture(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	commitFiles(t, dir, files)

	return dir
}

func commitFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()

	for name, content := range files {
		path := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}

	runGit(t, dir, "add", "-A")
	runGit(t, dir, "-c", "user.name=e2e", "-c", "user.email=e2e@shellhub.io", "commit", "-q", "--allow-empty", "-m", "fixture")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", dir}, args...)...) //nolint:gosec // runs git on the test's own fixture repository
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func TestPullOrBuild(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	pushing := liveRun(t)

	registry, err := StartRegistry(ctx, pushing)
	require.NoError(t, err)

	t.Cleanup(func() { _ = registry.Close(context.Background()) })

	newRemote := func(t *testing.T) string {
		t.Helper()

		remote := registry.Address + "/shellhub-e2e-cache:gateway-" + newRunID()
		t.Cleanup(func() { _, _ = cli.ImageRemove(context.Background(), remote, client.ImageRemoveOptions{}) })

		return remote
	}

	images := func(run *Run, remote string) []stackImage {
		return []stackImage{
			{service: "gateway", local: "shellhub-e2e/cache-test-gateway:" + run.ID(), remote: remote},
			{service: "ui", local: "shellhub-e2e/cache-test-ui:" + run.ID()},
		}
	}

	build := func(run *Run, built *[]string) func(context.Context, []string) error {
		return func(ctx context.Context, services []string) error {
			*built = append(*built, services...)

			for _, image := range images(run, "") {
				if slices.Contains(services, image.service) {
					if _, err := commitRunImage(ctx, cli, image.local, run.ID()); err != nil {
						return err
					}
				}
			}

			return nil
		}
	}

	pullable := func(t *testing.T, remote string) bool {
		t.Helper()

		pulled, err := cli.ImagePull(ctx, remote, client.ImagePullOptions{RegistryAuth: anonymousRegistryAuth})
		if err == nil {
			err = pulled.Wait(ctx)
		}

		if err != nil {
			return false
		}

		_, _ = cli.ImageRemove(ctx, remote, client.ImageRemoveOptions{})

		return true
	}

	t.Run("a miss builds and pushes nothing unless enabled", func(t *testing.T) {
		remote := newRemote(t)

		var built []string

		require.NoError(t, imageCache{}.pullOrBuild(ctx, pushing, images(pushing, remote), build(pushing, &built)))

		assert.Equal(t, []string{"gateway", "ui"}, built)
		assert.False(t, pullable(t, remote), "pushed %s with push disabled", remote)
	})

	t.Run("a miss builds and pushes what it can reuse", func(t *testing.T) {
		remote := newRemote(t)

		var built []string

		require.NoError(t, imageCache{push: true}.pullOrBuild(ctx, pushing, images(pushing, remote), build(pushing, &built)))

		assert.Equal(t, []string{"gateway", "ui"}, built)
		assert.True(t, pullable(t, remote), "did not push %s", remote)
	})

	t.Run("a hit takes the run's tag and label and skips the build", func(t *testing.T) {
		remote := newRemote(t)

		var seeded []string

		require.NoError(t, imageCache{push: true}.pullOrBuild(ctx, pushing, images(pushing, remote), build(pushing, &seeded)))

		reusing := liveRun(t)

		var built []string

		require.NoError(t, imageCache{}.pullOrBuild(ctx, reusing, images(reusing, remote), build(reusing, &built)))

		assert.Equal(t, []string{"ui"}, built)

		inspected, err := cli.ImageInspect(ctx, images(reusing, remote)[0].local)
		require.NoError(t, err)
		assert.Equal(t, reusing.ID(), inspected.Config.Labels[runLabel])
	})
}

func TestImageKey(t *testing.T) {
	repoFiles := map[string]string{
		"versions.env":  "GOLANG_VERSION=1\n",
		".dockerignore": ".git\n",

		"docker-compose.test.build.yml": "services: {}\n",
		"go.mod":                        "module github.com/shellhub-io/shellhub\n",
		"go.sum":                        "",
		"install.sh":                    "#!/bin/sh\n",
		"pkg/a.go":                      "package pkg\n",
		"openapi/spec.yaml":             "openapi: 3.1.0\n",
		"server/Dockerfile":             "FROM scratch\n",
		"server/main.go":                "package main\n",
		"gateway/Dockerfile":            "FROM scratch\n",
		"gateway/main.go":               "package main\n",
		"agent/main.go":                 "package main\n",
		"ui/index.html":                 "<html></html>\n",
	}

	cloudFiles := map[string]string{"go.mod": "module github.com/shellhub-io/cloud\n"}

	type key struct {
		service string
		edition Edition
		ldflags string
	}

	keyOf := func(t *testing.T, repo, cloud string, k key) string {
		t.Helper()

		got, err := imageKey(t.Context(), repo, cloud, k.service, k.edition, k.ldflags)
		require.NoError(t, err)

		return got
	}

	enterprise := key{service: "server", edition: EditionEnterprise, ldflags: "-X key=a"}

	cases := []struct {
		name    string
		key     key
		repo    map[string]string
		cloud   map[string]string
		changed bool
	}{
		{name: "a file under a server path", key: key{service: "server", edition: EditionCommunity}, repo: map[string]string{"server/main.go": "package main // changed\n"}, changed: true},
		{name: "a file under a shared path", key: key{service: "gateway", edition: EditionCommunity}, repo: map[string]string{"pkg/b.go": "package pkg\n"}, changed: true},
		{name: "the root module", key: key{service: "gateway", edition: EditionCommunity}, repo: map[string]string{"go.sum": "changed\n"}, changed: true},
		{name: "the Dockerfile", key: key{service: "gateway", edition: EditionCommunity}, repo: map[string]string{"gateway/Dockerfile": "FROM alpine\n"}, changed: true},
		{name: "versions.env", key: key{service: "server", edition: EditionCommunity}, repo: map[string]string{"versions.env": "GOLANG_VERSION=2\n"}, changed: true},
		{name: "the build's target and arguments", key: key{service: "gateway", edition: EditionCommunity}, repo: map[string]string{"docker-compose.test.build.yml": "services: {gateway: {}}\n"}, changed: true},
		{name: "the build context's ignore file", key: key{service: "server", edition: EditionCommunity}, repo: map[string]string{".dockerignore": ".git\n**/node_modules\n"}, changed: true},
		{name: "the install script", key: key{service: "server", edition: EditionCommunity}, repo: map[string]string{"install.sh": "#!/bin/sh -e\n"}, changed: true},
		{name: "the cloud source for the enterprise server", key: enterprise, cloud: map[string]string{"cmd/server/main.go": "package main\n"}, changed: true},
		{name: "a file outside the gateway's paths", key: key{service: "gateway", edition: EditionCommunity}, repo: map[string]string{"server/main.go": "package main // changed\n"}},
		{name: "a file outside every path", key: key{service: "server", edition: EditionCommunity}, repo: map[string]string{"agent/main.go": "package main // changed\n", "ui/index.html": "<html>changed</html>\n"}},
		{name: "the cloud source for the community server", key: key{service: "server", edition: EditionCommunity}, cloud: map[string]string{"cmd/server/main.go": "package main\n"}},
		{name: "the cloud source for the gateway", key: key{service: "gateway", edition: EditionEnterprise}, cloud: map[string]string{"cmd/server/main.go": "package main\n"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := gitFixture(t, repoFiles)
			cloud := gitFixture(t, cloudFiles)

			before := keyOf(t, repo, cloud, tc.key)

			if tc.repo != nil {
				commitFiles(t, repo, tc.repo)
			}

			if tc.cloud != nil {
				commitFiles(t, cloud, tc.cloud)
			}

			after := keyOf(t, repo, cloud, tc.key)

			if tc.changed {
				assert.NotEqual(t, before, after)
			} else {
				assert.Equal(t, before, after)
			}
		})
	}

	t.Run("the build arguments", func(t *testing.T) {
		repo := gitFixture(t, repoFiles)
		cloud := gitFixture(t, cloudFiles)

		community := keyOf(t, repo, cloud, key{service: "server", edition: EditionCommunity})
		base := keyOf(t, repo, cloud, enterprise)

		assert.NotEqual(t, community, base, "the edition")
		assert.NotEqual(t, base, keyOf(t, repo, cloud, key{service: "server", edition: EditionEnterprise, ldflags: "-X key=b"}), "the issuer key")
		assert.Equal(t, base, keyOf(t, repo, cloud, key{service: "server", edition: EditionCloud, ldflags: "-X key=a"}), "the cloud edition runs the enterprise server")
		assert.NotEqual(t, community, keyOf(t, repo, cloud, key{service: "gateway", edition: EditionCommunity}), "the image")
	})

	t.Run("fails for an image it cannot reuse", func(t *testing.T) {
		repo := gitFixture(t, repoFiles)

		_, err := imageKey(t.Context(), repo, "", "ui", EditionCommunity, "")
		require.Error(t, err)
	})
}

func TestReusableImageSourcesCoverTheirDockerfiles(t *testing.T) {
	root, err := testimage.Root()
	require.NoError(t, err)

	dockerfiles := map[string]string{
		"server":  "server/Dockerfile",
		"gateway": "gateway/Dockerfile",
		"agent":   defaultAgentDockerfile,
	}

	require.ElementsMatch(t, slices.Collect(maps.Keys(reusableImageSources)), slices.Collect(maps.Keys(dockerfiles)))

	for service, dockerfile := range dockerfiles {
		t.Run(service, func(t *testing.T) {
			paths := reusableImageSources[service]
			assert.Contains(t, paths, dockerfile)

			content, err := os.ReadFile(filepath.Join(root, dockerfile)) //nolint:gosec // the path is one of the fixed Dockerfiles listed above
			require.NoError(t, err)

			for _, source := range copiedSources(string(content)) {
				covered := slices.ContainsFunc(paths, func(path string) bool {
					return source == path || strings.HasPrefix(source, path+"/")
				})
				assert.True(t, covered, "%s copies %s, which no path in the %s image's key covers", dockerfile, source, service)
			}
		})
	}
}

func copiedSources(dockerfile string) []string {
	var sources []string

	for instruction := range strings.Lines(strings.ReplaceAll(dockerfile, "\\\n", " ")) {
		fields := strings.Fields(instruction)
		if len(fields) == 0 {
			continue
		}

		switch strings.ToUpper(fields[0]) {
		case "COPY", "ADD":
			sources = append(sources, contextSources(fields[1:])...)
		case "RUN":
			sources = append(sources, boundSources(fields[1:])...)
		}
	}

	return sources
}

func contextSources(args []string) []string {
	if len(args) < 2 || slices.ContainsFunc(args, func(arg string) bool { return strings.HasPrefix(arg, "--from=") }) {
		return nil
	}

	var sources []string

	for _, arg := range args[:len(args)-1] {
		if !strings.HasPrefix(arg, "--") && !strings.Contains(arg, "://") {
			sources = append(sources, contextPath(arg))
		}
	}

	return sources
}

func boundSources(args []string) []string {
	var sources []string

	for _, arg := range args {
		options, ok := strings.CutPrefix(arg, "--mount=")
		if !ok {
			continue
		}

		mount := map[string]string{"source": "."}
		for option := range strings.SplitSeq(options, ",") {
			key, value, _ := strings.Cut(option, "=")
			mount[key] = value
		}

		if mount["type"] == "bind" && mount["from"] == "" {
			sources = append(sources, contextPath(mount["source"]))
		}
	}

	return sources
}

func contextPath(path string) string {
	return strings.TrimSuffix(strings.TrimPrefix(path, "./"), "/")
}

func TestCopiedSources(t *testing.T) {
	dockerfile := strings.Join([]string{
		"FROM golang AS builder",
		"COPY ./go.mod ./go.sum ./",
		"COPY --chown=1000:1000 ./pkg /src/pkg",
		"COPY --from=builder /server /server",
		"ADD ./install.sh /install.sh",
		"ADD https://example.test/tool.tar.gz /tool.tar.gz",
		"COPY ./server \\",
		"     ./openapi/ /src/",
		"RUN --mount=type=bind,source=./gateway,target=/gateway \\",
		"    --mount=type=bind,from=cloud-src,target=/cloud \\",
		"    --mount=type=cache,id=go,target=/go \\",
		"    go build",
		"RUN --mount=type=bind,target=/context true",
		"",
	}, "\n")

	assert.Equal(t,
		[]string{"go.mod", "go.sum", "pkg", "install.sh", "server", "openapi", "gateway", "."},
		copiedSources(dockerfile),
	)
}
