package testrun_test

import (
	"bufio"
	"context"
	"crypto/rand"
	"maps"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/testimage"
	"github.com/shellhub-io/shellhub/pkg/testrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

const (
	helperEnv     = "SHELLHUB_TESTRUN_HELPER"
	helperProject = "SHELLHUB_TESTRUN_HELPER_PROJECT"
	helperImage   = "SHELLHUB_TESTRUN_HELPER_IMAGE"
	helperReady   = "testrun-helper-ready "
)

var wallClock = clock.New()

type runObjects struct {
	run       string
	image     string
	container string
	network   string
}

func dockerClient(t *testing.T) *client.Client {
	t.Helper()

	cli, err := testcontainers.NewDockerClientWithOpts(t.Context())
	require.NoError(t, err)

	t.Cleanup(func() { _ = cli.Close() })

	return cli.Client
}

func alpineImage() (string, error) {
	if image := envs.OS().Get(helperImage); image != "" {
		return image, nil
	}

	version, err := testimage.Version("../..", "ALPINE_VERSION")

	return testimage.DockerHub("alpine:" + version), err
}

func newID() string {
	return strings.ToLower(rand.Text()[:12])
}

func newRunImage(ctx context.Context, cli *client.Client, id string) (string, error) {
	image, err := alpineImage()
	if err != nil {
		return "", err
	}

	source, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image:  image,
		Config: &container.Config{Labels: map[string]string{testrun.RunLabel: id}},
	})
	if err != nil {
		return "", err
	}

	defer func() {
		_, _ = cli.ContainerRemove(context.WithoutCancel(ctx), source.ID, client.ContainerRemoveOptions{Force: true})
	}()

	ref := "shellhub-testrun/run-test:" + id
	if _, err := cli.ContainerCommit(ctx, source.ID, client.ContainerCommitOptions{
		Reference: ref,
		Changes:   []string{"LABEL " + testrun.RunLabel + "=" + id},
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

	network, err := cli.NetworkCreate(ctx, "shellhub-testrun-"+id, client.NetworkCreateOptions{
		Labels: map[string]string{testrun.RunLabel: id},
	})
	if err != nil {
		return runObjects{}, err
	}

	return runObjects{run: id, image: image, container: created.ID, network: network.ID}, nil
}

func newLeasedObjects(t *testing.T, leaseLabels map[string]string) runObjects {
	t.Helper()

	ctx := t.Context()
	cli := dockerClient(t)
	id := newID()

	labels := map[string]string{testrun.RunLabel: id, testrun.LeaseLabel: id}
	maps.Copy(labels, leaseLabels)

	objects, err := newRunObjects(ctx, cli, id, labels)
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = cli.ContainerRemove(ctx, objects.container, client.ContainerRemoveOptions{Force: true})
		_, _ = cli.NetworkRemove(ctx, objects.network, client.NetworkRemoveOptions{})
		_, _ = cli.ImageRemove(ctx, objects.image, client.ImageRemoveOptions{})
	})

	return objects
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

func liveRun(t *testing.T) *testrun.Run {
	t.Helper()

	run, err := testrun.Start(t.Context())
	require.NoError(t, err)

	t.Cleanup(func() { _ = run.Close(context.Background()) })

	return run
}

func sweepByStartingRun(t assert.TestingT, ctx context.Context) {
	run, err := testrun.Start(ctx)
	if assert.NoError(t, err) {
		assert.NoError(t, run.Close(ctx))
	}
}

func TestHelperProcess(t *testing.T) {
	mode := envs.OS().Get(helperEnv)
	if mode == "" {
		t.Skip("re-executed by the run tests")
	}

	ctx := context.Background()
	cli := dockerClient(t)

	var (
		objects runObjects
		run     *testrun.Run
	)

	switch mode {
	case "process", "stack":
		var err error

		if mode == "process" {
			run, err = testrun.Start(ctx)
		} else {
			run, err = testrun.StartStack(ctx, envs.OS().Get(helperProject), testrun.StackOptions{TTL: time.Hour})
		}

		require.NoError(t, err)

		objects, err = newRunObjects(ctx, cli, run.ID(), run.Labels())
		require.NoError(t, err)
	case "labelled":
		image, err := alpineImage()
		require.NoError(t, err)

		c, err := testcontainers.Run(ctx, image, testcontainers.WithCmd("sleep", "3600"), testrun.Labelled())
		require.NoError(t, err)

		inspected, err := c.Inspect(ctx)
		require.NoError(t, err)

		objects = runObjects{run: inspected.Config.Labels[testrun.RunLabel], image: "-", container: c.GetContainerID(), network: "-"}
	default:
		t.Fatalf("unknown helper mode %q", mode)
	}

	_, err := os.Stdout.WriteString(helperReady + objects.run + " " + objects.image + " " + objects.container + " " + objects.network + "\n")
	require.NoError(t, err)

	if mode == "stack" {
		return
	}

	time.Sleep(time.Hour)
	runtime.KeepAlive(run)
}

func startHelper(t *testing.T, mode, dir string, env ...string) (*exec.Cmd, runObjects) {
	t.Helper()

	image, err := alpineImage()
	require.NoError(t, err)

	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestHelperProcess$", "-test.timeout=10m") //nolint:gosec // re-executes this test binary
	cmd.Env = append(append(os.Environ(), helperEnv+"="+mode, helperImage+"="+image, "TESTCONTAINERS_RYUK_DISABLED=true"), env...)
	cmd.Dir = dir
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
		line, ok := strings.CutPrefix(scanner.Text(), helperReady)
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

func kill(t *testing.T, helper *exec.Cmd) {
	t.Helper()

	require.NoError(t, helper.Process.Signal(syscall.SIGKILL))
	_ = helper.Wait()
}

func TestStartSweepsKilledRun(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	helper, objects := startHelper(t, "process", "")
	assertObjectsExist(t, ctx, cli, objects)

	kill(t, helper)

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		sweepByStartingRun(tt, ctx)
		assertObjectsGone(tt, ctx, cli, objects)
	}, 30*time.Second, time.Second)
}

func TestStartLeavesLiveRunAlone(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	live := liveRun(t)

	objects, err := newRunObjects(ctx, cli, live.ID(), live.Labels())
	require.NoError(t, err)

	sweepByStartingRun(t, ctx)

	assertObjectsExist(t, ctx, cli, objects)
}

func TestStartRunsWithoutVersionsEnvAboveTheWorkingDirectory(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	_, objects := startHelper(t, "process", t.TempDir())

	leases, err := cli.ContainerList(ctx, client.ContainerListOptions{
		Filters: make(client.Filters).Add("label", testrun.LeaseLabel+"="+objects.run),
	})
	require.NoError(t, err)
	require.Len(t, leases.Items, 1)
	assert.Equal(t, container.StateRunning, leases.Items[0].State)
}

func TestStackRunLivesWhileItsContainersExist(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	project := "shellhub-testrun-" + newID()

	helper, objects := startHelper(t, "stack", "", helperProject+"="+project)
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
		leases, err := cli.ContainerList(ctx, client.ContainerListOptions{
			All:     true,
			Filters: make(client.Filters).Add("label", testrun.LeaseLabel+"="+objects.run),
		})
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

func TestSweepRemovesCreatedLeaseOlderThanTheGracePeriod(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	defaultGrace := newLeasedObjects(t, map[string]string{testrun.KindLabel: testrun.KindProcess})
	shortenedGrace := newLeasedObjects(t, map[string]string{testrun.KindLabel: testrun.KindProcess})

	t.Cleanup(testrun.SetProcessGraceOf(shortenedGrace.run, time.Second))

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		sweepByStartingRun(tt, ctx)
		assertObjectsGone(tt, ctx, cli, shortenedGrace)
	}, 30*time.Second, time.Second)

	assertObjectsExist(t, ctx, cli, defaultGrace)
}

func TestSweepKeepsCreatedLeaseWithinTheGracePeriod(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	objects := newLeasedObjects(t, map[string]string{testrun.KindLabel: testrun.KindProcess})

	sweepByStartingRun(t, ctx)

	assertObjectsExist(t, ctx, cli, objects)
}

func TestSweepRemovesExpiredStackRun(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	objects := newLeasedObjects(t, map[string]string{
		testrun.KindLabel:    testrun.KindStack,
		testrun.ExpiresLabel: wallClock.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
	})

	sweepByStartingRun(t, ctx)

	assertObjectsGone(t, ctx, cli, objects)
}

func TestSweepKeepsUnexpiredStackRun(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	objects := newLeasedObjects(t, map[string]string{
		testrun.KindLabel:    testrun.KindStack,
		testrun.ExpiresLabel: wallClock.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	})

	sweepByStartingRun(t, ctx)

	assertObjectsExist(t, ctx, cli, objects)
}

func TestSweepRemovesContainerOfKilledLabelledProcess(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	helper, objects := startHelper(t, "labelled", "")
	require.NotEmpty(t, objects.run, "the container carries no run label")

	t.Cleanup(func() {
		_, _ = cli.ContainerRemove(context.Background(), objects.container, client.ContainerRemoveOptions{Force: true})
	})

	kill(t, helper)

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		sweepByStartingRun(tt, ctx)
		assert.False(tt, containerExists(ctx, cli, objects.container), "container %s", objects.container)
	}, 30*time.Second, time.Second)
}

func TestSweepSkipsImageInUse(t *testing.T) {
	ctx := t.Context()
	cli := dockerClient(t)

	helper, objects := startHelper(t, "process", "")
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

	kill(t, helper)

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

func TestStartStackRefusesNonPositiveTTL(t *testing.T) {
	_, err := testrun.StartStack(t.Context(), "shellhub-testrun-ttl", testrun.StackOptions{})
	require.ErrorContains(t, err, "TTL")
}

func TestStackRunLabelsCarryItsExpiryAndCheckout(t *testing.T) {
	before := wallClock.Now().Add(2 * time.Hour).Truncate(time.Second)

	run, err := testrun.StartStack(t.Context(), "shellhub-testrun-"+newID(), testrun.StackOptions{TTL: 2 * time.Hour, Checkout: "/src/shellhub"})
	require.NoError(t, err)

	t.Cleanup(func() { _ = run.Close(context.Background()) })

	labels := run.Labels()
	assert.Equal(t, testrun.KindStack, labels[testrun.KindLabel])
	assert.Equal(t, run.ID(), labels[testrun.LeaseLabel])
	assert.Equal(t, "/src/shellhub", labels[testrun.CheckoutLabel])

	expires, err := time.Parse(time.RFC3339, labels[testrun.ExpiresLabel])
	require.NoError(t, err)
	assert.WithinRange(t, expires, before.Add(-time.Second), wallClock.Now().Add(2*time.Hour))
}

func TestLeaseAlive(t *testing.T) {
	now := wallClock.Now()
	ago := func(d time.Duration) int64 { return now.Add(-d).Unix() }

	for _, test := range []struct {
		name  string
		lease container.Summary
		want  bool
	}{
		{
			name:  "a running process lease",
			lease: container.Summary{State: container.StateRunning, Created: ago(time.Hour), Labels: map[string]string{testrun.KindLabel: testrun.KindProcess}},
			want:  true,
		},
		{
			name:  "a process lease created within the grace period and never started",
			lease: container.Summary{State: container.StateCreated, Created: ago(time.Second), Labels: map[string]string{testrun.KindLabel: testrun.KindProcess}},
			want:  true,
		},
		{
			name:  "a process lease created past the grace period and never started",
			lease: container.Summary{State: container.StateCreated, Created: ago(61 * time.Second), Labels: map[string]string{testrun.KindLabel: testrun.KindProcess}},
		},
		{
			name:  "an exited stack lease before its expiry",
			lease: container.Summary{State: container.StateExited, Created: ago(48 * time.Hour), Labels: map[string]string{testrun.KindLabel: testrun.KindStack, testrun.ExpiresLabel: now.Add(time.Minute).Format(time.RFC3339)}},
			want:  true,
		},
		{
			name:  "a running stack lease past its expiry",
			lease: container.Summary{State: container.StateRunning, Created: ago(time.Minute), Labels: map[string]string{testrun.KindLabel: testrun.KindStack, testrun.ExpiresLabel: now.Add(-time.Minute).Format(time.RFC3339)}},
		},
		{
			name:  "a stack lease with an unreadable expiry",
			lease: container.Summary{State: container.StateRunning, Created: ago(time.Minute), Labels: map[string]string{testrun.KindLabel: testrun.KindStack, testrun.ExpiresLabel: "tomorrow"}},
		},
		{
			name:  "a legacy compose lease younger than a day",
			lease: container.Summary{State: container.StateRunning, Created: ago(time.Hour), Labels: map[string]string{"com.docker.compose.project": "shellhub-e2e-x"}},
			want:  true,
		},
		{
			name:  "a legacy compose lease older than a day",
			lease: container.Summary{State: container.StateRunning, Created: ago(25 * time.Hour), Labels: map[string]string{"com.docker.compose.project": "shellhub-e2e-x"}},
		},
		{
			name:  "a legacy process lease that is running",
			lease: container.Summary{State: container.StateRunning, Created: ago(48 * time.Hour)},
			want:  true,
		},
		{
			name:  "a legacy process lease that never started",
			lease: container.Summary{State: container.StateCreated, Created: ago(48 * time.Hour)},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, testrun.LeaseAlive(test.lease, now))
		})
	}
}
