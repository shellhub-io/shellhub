package environment

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	runHelperEnv     = "SHELLHUB_E2E_RUN_HELPER"
	runHelperProject = "SHELLHUB_E2E_RUN_HELPER_PROJECT"
	runHelperReady   = "run-helper-ready "
)

type runObjects struct {
	run       string
	image     string
	container string
	network   string
}

func dockerClient(t *testing.T) *client.Client {
	t.Helper()

	require.NoError(t, pinDockerHost())

	cli, err := client.New(client.FromEnv)
	require.NoError(t, err)

	t.Cleanup(func() { _ = cli.Close() })

	return cli
}

func newRunImage(ctx context.Context, cli *client.Client, id string) (string, error) {
	image, err := leaseImage()
	if err != nil {
		return "", err
	}

	source, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image:  image,
		Config: &container.Config{Labels: map[string]string{runLabel: id}},
	})
	if err != nil {
		return "", err
	}

	defer func() {
		_, _ = cli.ContainerRemove(context.WithoutCancel(ctx), source.ID, client.ContainerRemoveOptions{Force: true})
	}()

	ref := "shellhub-e2e/run-test:" + id
	if _, err := cli.ContainerCommit(ctx, source.ID, client.ContainerCommitOptions{
		Reference: ref,
		Changes:   []string{"LABEL " + runLabel + "=" + id},
	}); err != nil {
		return "", err
	}

	return ref, nil
}

func newRunObjects(ctx context.Context, cli *client.Client, id string, labels map[string]string) (runObjects, error) {
	image, err := newRunImage(ctx, cli, id)
	if err != nil {
		return runObjects{}, err
	}

	created, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image:  image,
		Config: &container.Config{Labels: labels},
	})
	if err != nil {
		return runObjects{}, err
	}

	network, err := cli.NetworkCreate(ctx, "shellhub-e2e-run-test-"+id, client.NetworkCreateOptions{
		Labels: map[string]string{runLabel: id},
	})
	if err != nil {
		return runObjects{}, err
	}

	return runObjects{run: id, image: image, container: created.ID, network: network.ID}, nil
}

func imageExists(ctx context.Context, cli *client.Client, ref string) bool {
	_, err := cli.ImageInspect(ctx, ref)

	return err == nil
}

func containerExists(ctx context.Context, cli *client.Client, id string) bool {
	_, err := cli.ContainerInspect(ctx, id, client.ContainerInspectOptions{})

	return err == nil
}

func networkExists(ctx context.Context, cli *client.Client, id string) bool {
	_, err := cli.NetworkInspect(ctx, id, client.NetworkInspectOptions{})

	return err == nil
}

func assertObjectsExist(t assert.TestingT, ctx context.Context, cli *client.Client, objects runObjects) {
	assert.True(t, imageExists(ctx, cli, objects.image), "image %s", objects.image)
	assert.True(t, containerExists(ctx, cli, objects.container), "container %s", objects.container)
	assert.True(t, networkExists(ctx, cli, objects.network), "network %s", objects.network)
}

func assertObjectsGone(t assert.TestingT, ctx context.Context, cli *client.Client, objects runObjects) {
	assert.False(t, imageExists(ctx, cli, objects.image), "image %s", objects.image)
	assert.False(t, containerExists(ctx, cli, objects.container), "container %s", objects.container)
	assert.False(t, networkExists(ctx, cli, objects.network), "network %s", objects.network)
}

func liveRun(t *testing.T) *Run {
	t.Helper()

	run, err := StartRun(t.Context(), nil)
	require.NoError(t, err)

	t.Cleanup(func() { _ = run.Close(context.Background()) })

	return run
}

func sweepByStartingRun(t assert.TestingT, ctx context.Context) {
	run, err := StartRun(ctx, nil)
	if assert.NoError(t, err) {
		assert.NoError(t, run.Close(ctx))
	}
}

func TestRunHelperProcess(t *testing.T) {
	mode := os.Getenv(runHelperEnv)
	if mode == "" {
		t.Skip("re-executed by the run tests")
	}

	ctx := context.Background()

	var (
		run *Run
		err error
	)

	switch mode {
	case "process":
		run, err = StartRun(ctx, nil)
	case "stack":
		run, err = StartStackRun(ctx, os.Getenv(runHelperProject), nil)
	default:
		t.Fatalf("unknown helper mode %q", mode)
	}

	require.NoError(t, err)

	cli := dockerClient(t)

	objects, err := newRunObjects(ctx, cli, run.ID(), run.Labels())
	require.NoError(t, err)

	_, err = os.Stdout.WriteString(runHelperReady + run.ID() + " " + objects.image + " " + objects.container + " " + objects.network + "\n")
	require.NoError(t, err)

	if mode == "stack" {
		return
	}

	time.Sleep(time.Hour)
}

func startHelper(t *testing.T, mode string, env ...string) (*exec.Cmd, runObjects) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestRunHelperProcess$", "-test.timeout=10m") //nolint:gosec // re-executes this test binary
	cmd.Env = append(append(os.Environ(), runHelperEnv+"="+mode), env...)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())

	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line, ok := strings.CutPrefix(scanner.Text(), runHelperReady)
		if !ok {
			continue
		}

		fields := strings.Fields(line)
		require.Len(t, fields, 4)

		go func() {
			for scanner.Scan() {
			}
		}()

		return cmd, runObjects{run: fields[0], image: fields[1], container: fields[2], network: fields[3]}
	}

	require.FailNow(t, "the helper exited before it was ready", "%v", scanner.Err())

	return nil, runObjects{}
}

func TestStartRunSweepsKilledRun(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	helper, objects := startHelper(t, "process")
	assertObjectsExist(t, ctx, cli, objects)

	require.NoError(t, helper.Process.Signal(syscall.SIGKILL))
	_ = helper.Wait()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		sweepByStartingRun(tt, ctx)
		assertObjectsGone(tt, ctx, cli, objects)
	}, 30*time.Second, time.Second)
}

func TestStartRunLeavesLiveRunAlone(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	live := liveRun(t)

	objects, err := newRunObjects(ctx, cli, live.ID(), live.Labels())
	require.NoError(t, err)

	sweepByStartingRun(t, ctx)

	assertObjectsExist(t, ctx, cli, objects)
}

func TestStackRunLivesWhileItsContainersExist(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	project := "shellhub-e2e-run-test-" + newRunID()

	helper, objects := startHelper(t, "stack", runHelperProject+"="+project)
	require.NoError(t, helper.Wait())

	t.Cleanup(func() {
		_, _ = cli.ContainerRemove(context.Background(), objects.container, client.ContainerRemoveOptions{Force: true})
	})

	_, err := cli.ContainerStart(ctx, objects.container, client.ContainerStartOptions{})
	require.NoError(t, err)

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		inspected, err := cli.ContainerInspect(ctx, objects.container, client.ContainerInspectOptions{})
		if assert.NoError(tt, err) {
			assert.Equal(tt, container.StateExited, inspected.Container.State.Status)
		}
	}, 30*time.Second, time.Second, "the stack's container never stopped")

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		leases, err := cli.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: labelEquals(leaseLabel, objects.run)})
		if assert.NoError(tt, err) && assert.Len(tt, leases.Items, 1) {
			assert.Equal(tt, objects.container, leases.Items[0].ID)
		}
	}, 30*time.Second, time.Second, "the helper's own lease outlived it")

	sweepByStartingRun(t, ctx)
	assertObjectsExist(t, ctx, cli, objects)

	_, err = cli.ContainerRemove(ctx, objects.container, client.ContainerRemoveOptions{Force: true})
	require.NoError(t, err)

	sweepByStartingRun(t, ctx)
	assertObjectsGone(t, ctx, cli, objects)
}

func TestSweepSkipsImageInUse(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	helper, objects := startHelper(t, "process")
	live := liveRun(t)

	user, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image:  objects.image,
		Config: &container.Config{Labels: live.Labels()},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = cli.ContainerRemove(context.Background(), user.ID, client.ContainerRemoveOptions{Force: true})
		_, _ = cli.ImageRemove(context.Background(), objects.image, client.ImageRemoveOptions{})
	})

	require.NoError(t, helper.Process.Signal(syscall.SIGKILL))
	_ = helper.Wait()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		sweepByStartingRun(tt, ctx)
		assert.False(tt, containerExists(ctx, cli, objects.container), "container %s", objects.container)
	}, 30*time.Second, time.Second)

	assert.True(t, imageExists(ctx, cli, objects.image), "image %s", objects.image)
}

func TestCloseUntagsOnlyItsOwnImages(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	closing := liveRun(t)
	other := liveRun(t)

	own, err := newRunImage(ctx, cli, closing.ID())
	require.NoError(t, err)

	others, err := newRunImage(ctx, cli, other.ID())
	require.NoError(t, err)

	require.NoError(t, closing.Close(ctx))

	assert.False(t, imageExists(ctx, cli, own), "image %s", own)
	assert.True(t, imageExists(ctx, cli, others), "image %s", others)
}
