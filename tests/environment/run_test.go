package environment

import (
	"context"
	"crypto/rand"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/shellhub-io/shellhub/pkg/testimage"
	"github.com/stretchr/testify/require"
)

func dockerClient(t *testing.T) *client.Client {
	t.Helper()

	require.NoError(t, pinDockerHost())

	cli, err := client.New(client.FromEnv)
	require.NoError(t, err)

	t.Cleanup(func() { _ = cli.Close() })

	return cli
}

func liveRun(t *testing.T) *Run {
	t.Helper()

	run, err := StartRun(t.Context(), nil)
	require.NoError(t, err)

	t.Cleanup(func() { _ = run.Close(context.Background()) })

	return run
}

func newRunID() string {
	return strings.ToLower(rand.Text()[:12])
}

func commitRunImage(ctx context.Context, cli *client.Client, ref, id string) (string, error) {
	root, err := testimage.Root()
	if err != nil {
		return "", err
	}

	alpine, err := testimage.Version(root, "ALPINE_VERSION")
	if err != nil {
		return "", err
	}

	source, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image:  testimage.DockerHub("alpine:" + alpine),
		Config: &container.Config{Labels: map[string]string{runLabel: id}},
	})
	if err != nil {
		return "", err
	}

	defer func() {
		_, _ = cli.ContainerRemove(context.WithoutCancel(ctx), source.ID, client.ContainerRemoveOptions{Force: true})
	}()

	if _, err := cli.ContainerCommit(ctx, source.ID, client.ContainerCommitOptions{
		Reference: ref,
		Changes:   []string{"LABEL " + runLabel + "=" + id},
	}); err != nil {
		return "", err
	}

	return ref, nil
}
