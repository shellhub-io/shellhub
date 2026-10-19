// Package testrun owns the Docker objects a test process creates on a shared daemon. Every object
// carries the label of the run that created it, every run holds a lease while it is alive, and
// every run that starts sweeps the objects of the runs whose leases are dead, whichever suite
// started them.
package testrun

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/testimage"
	log "github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
)

const (
	// RunLabel names the run an object belongs to.
	RunLabel = "io.shellhub.e2e.run"
	// LeaseLabel marks the containers that keep a run alive.
	LeaseLabel = "io.shellhub.e2e.lease"
	// KindLabel says how a lease container keeps its run alive: KindProcess or KindStack.
	KindLabel = "io.shellhub.e2e.kind"
	// ExpiresLabel holds, in RFC 3339, the time a stack run's leases stop keeping it alive.
	ExpiresLabel = "io.shellhub.e2e.expires"
	// CheckoutLabel holds the absolute path of the repository checkout that started a stack run.
	CheckoutLabel = "io.shellhub.e2e.checkout"

	// KindProcess is the kind of a lease a process holds open for as long as it lives.
	KindProcess = "process"
	// KindStack is the kind of a lease a kept compose stack's containers carry.
	KindStack = "stack"

	composeProjectLabel = "com.docker.compose.project"
	runIDSize           = 12
	legacyStackTTL      = 24 * time.Hour
	processGrace        = time.Minute
)

var (
	wallClock     clock.Clock = clock.New()
	processGraces             = map[string]time.Duration{}

	validStackProject = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,99}$`)
)

// Run owns the objects one test process or one kept stack creates on a shared daemon: the images
// it builds and the containers and networks it starts, each labelled with the run's ID.
//
// A run is alive while any of its leases is. A process lease is a container whose stdin the
// process holds open; it is alive while it runs and for a grace period after its creation, which
// covers the moment between creating and starting it. A stack lease is one of a kept stack's own
// containers; it is alive while it exists and its expiry is in the future. A lease with no kind
// counts as a stack lease, alive for 24 hours after its creation, when its container belongs to a
// compose project, and as a process lease otherwise.
type Run struct {
	id       string
	stack    bool
	expires  time.Time
	checkout string
	client   *client.Client
	lease    *client.HijackedResponse
}

// StackOptions sets how long a kept stack lives and which checkout it belongs to. A later run
// sweeps the stack once it expires.
type StackOptions struct {
	// TTL is how long after the run starts its stack stays alive.
	TTL time.Duration
	// Checkout is the absolute path of the repository checkout that started the stack.
	Checkout string
}

// Start starts a run for the calling process. Its lease is a container whose stdin the process
// holds open, so the daemon releases the lease however the process ends. It sweeps the dead runs
// on the daemon once the lease exists, logging what it cannot remove. It returns an error when the
// package was built with -trimpath, versions.env cannot be read, has a line that is not KEY=VALUE
// or does not pin ALPINE_VERSION, the daemon cannot be reached, or the lease cannot be created. The lease outlives ctx;
// [Run.Close] releases it.
func Start(ctx context.Context) (*Run, error) {
	return start(ctx, &Run{id: strings.ToLower(rand.Text()[:runIDSize])})
}

// StartStack starts the run of a kept compose stack, identified by its project name. The stack's
// containers carry the lease, so the run stays alive after the calling process exits, for as long
// as any of them exists and opts.TTL has not passed. The calling process holds a process lease of
// its own until it exits or calls [Run.Close], covering the objects it creates before the first
// stack container. It returns an error when project is not a valid image tag and project name,
// when opts.TTL is not positive, and in the cases [Start] does.
func StartStack(ctx context.Context, project string, opts StackOptions) (*Run, error) {
	if !validStackProject.MatchString(project) {
		return nil, fmt.Errorf("invalid stack project %q: must match %s", project, validStackProject.String())
	}

	if opts.TTL <= 0 {
		return nil, fmt.Errorf("invalid stack TTL %s: must be positive", opts.TTL)
	}

	return start(ctx, &Run{
		id:       project,
		stack:    true,
		expires:  wallClock.Now().Add(opts.TTL).UTC().Truncate(time.Second),
		checkout: opts.Checkout,
	})
}

func start(ctx context.Context, r *Run) (*Run, error) {
	image, err := leaseImage()
	if err != nil {
		return nil, err
	}

	cli, err := testcontainers.NewDockerClientWithOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("new docker client: %w", err)
	}

	r.client = cli.Client

	if err := r.holdLease(ctx, image); err != nil {
		_ = cli.Close()

		return nil, err
	}

	r.sweep(ctx)

	return r, nil
}

var labelledRun = sync.OnceValues(func() (*Run, error) {
	return Start(context.Background())
})

// Labelled joins the container a request starts to a run, so the next run on the daemon sweeps
// it when the process dies without removing it. It labels the container only, not an image the
// request builds. On the first request it customizes, it starts a run for the calling process,
// once per process, which the process never closes. That start ignores the request's context,
// which testcontainers does not pass to the option. The request fails with the error [Start]
// returns when that run cannot start, and so does every later request in the process.
func Labelled() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		run, err := labelledRun()
		if err != nil {
			return fmt.Errorf("joining a test run: %w", err)
		}

		if req.Labels == nil {
			req.Labels = make(map[string]string)
		}

		maps.Copy(req.Labels, run.Labels())

		return nil
	}
}

// ID returns the run's ID, valid as an image tag and as a compose project name.
func (r *Run) ID() string { return r.id }

// Labels returns the labels every container the run starts must carry. A stack run's include the
// lease, its kind and its expiry, so its containers keep the run alive until it expires, and its
// checkout when [StackOptions.Checkout] was set.
func (r *Run) Labels() map[string]string {
	labels := map[string]string{RunLabel: r.id}
	if r.stack {
		labels[LeaseLabel] = r.id
		labels[KindLabel] = KindStack
		labels[ExpiresLabel] = r.expires.Format(time.RFC3339)

		if r.checkout != "" {
			labels[CheckoutLabel] = r.checkout
		}
	}

	return labels
}

// Close removes every object labelled with the run, then releases the lease the process holds and
// closes the run's Docker client. It is best-effort: it returns every listing or removal that
// failed, which the next sweep retries once the run is dead, joined with any error closing the
// client. ctx bounds the removals; the lease is released whatever ctx's state.
func (r *Run) Close(ctx context.Context) error {
	err := r.remove(ctx, r.id)

	if r.lease != nil {
		r.lease.Close()
		r.lease = nil
	}

	return errors.Join(err, r.client.Close())
}

func leaseImage() (string, error) {
	_, source, _, ok := runtime.Caller(0)
	if !ok || !filepath.IsAbs(source) {
		return "", errors.New("testrun was built without its source path (-trimpath?), so it cannot find versions.env")
	}

	alpine, err := testimage.Version(filepath.Join(filepath.Dir(source), "..", ".."), "ALPINE_VERSION")
	if err != nil {
		return "", err
	}

	return "alpine:" + alpine, nil
}

func (r *Run) holdLease(ctx context.Context, image string) error {
	if _, err := r.client.ImageInspect(ctx, image); err != nil {
		pull, err := r.client.ImagePull(ctx, image, client.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("pulling %s: %w", image, err)
		}

		if err := pull.Wait(ctx); err != nil {
			return fmt.Errorf("pulling %s: %w", image, err)
		}
	}

	created, err := r.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image: image,
		Config: &container.Config{
			Cmd:         []string{"cat"},
			AttachStdin: true,
			OpenStdin:   true,
			StdinOnce:   true,
			Labels:      map[string]string{LeaseLabel: r.id, RunLabel: r.id, KindLabel: KindProcess},
		},
		HostConfig: &container.HostConfig{
			AutoRemove:  true,
			NetworkMode: "none",
		},
	})
	if err != nil {
		return fmt.Errorf("creating the lease container: %w", err)
	}

	attached, err := r.client.ContainerAttach(context.WithoutCancel(ctx), created.ID, client.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
	})
	if err != nil {
		_, _ = r.client.ContainerRemove(context.WithoutCancel(ctx), created.ID, client.ContainerRemoveOptions{Force: true})

		return fmt.Errorf("attaching to the lease container: %w", err)
	}

	r.lease = &attached.HijackedResponse

	if _, err := r.client.ContainerStart(ctx, created.ID, client.ContainerStartOptions{}); err != nil {
		r.lease.Close()
		_, _ = r.client.ContainerRemove(context.WithoutCancel(ctx), created.ID, client.ContainerRemoveOptions{Force: true})

		return fmt.Errorf("starting the lease container: %w", err)
	}

	return nil
}

func (r *Run) sweep(ctx context.Context) {
	ids, err := r.labelledRuns(ctx)
	if err != nil {
		log.WithError(err).Warn("test run sweep: could not list the runs on the daemon")

		return
	}

	for id := range ids {
		if id == r.id {
			continue
		}

		alive, err := r.alive(ctx, id)
		if err != nil {
			log.WithError(err).WithField("run", id).Warn("test run sweep: could not read the run's leases")

			continue
		}

		if alive {
			continue
		}

		if err := r.remove(ctx, id); err != nil {
			log.WithError(err).WithField("run", id).Warn("test run sweep: left objects of a dead run behind")
		}
	}
}

func runLabelled() client.Filters {
	return make(client.Filters).Add("label", RunLabel)
}

func labelEquals(label, value string) client.Filters {
	return make(client.Filters).Add("label", label+"="+value)
}

func (r *Run) labelledRuns(ctx context.Context) (map[string]struct{}, error) {
	ids := make(map[string]struct{})
	collect := func(labels map[string]string) {
		if id := labels[RunLabel]; id != "" {
			ids[id] = struct{}{}
		}
	}

	containers, err := r.client.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: runLabelled()})
	if err != nil {
		return nil, err
	}

	for _, c := range containers.Items {
		collect(c.Labels)
	}

	images, err := r.client.ImageList(ctx, client.ImageListOptions{Filters: runLabelled()})
	if err != nil {
		return nil, err
	}

	for _, i := range images.Items {
		collect(i.Labels)
	}

	networks, err := r.client.NetworkList(ctx, client.NetworkListOptions{Filters: runLabelled()})
	if err != nil {
		return nil, err
	}

	for _, n := range networks.Items {
		collect(n.Labels)
	}

	return ids, nil
}

func (r *Run) alive(ctx context.Context, id string) (bool, error) {
	leases, err := r.client.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: labelEquals(LeaseLabel, id)})
	if err != nil {
		return false, err
	}

	now := wallClock.Now()
	for _, lease := range leases.Items {
		if leaseAlive(lease, now) {
			return true, nil
		}
	}

	return false, nil
}

func leaseAlive(lease container.Summary, now time.Time) bool {
	created := time.Unix(lease.Created, 0)

	kind := lease.Labels[KindLabel]

	switch {
	case kind == KindStack:
		expires, err := time.Parse(time.RFC3339, lease.Labels[ExpiresLabel])

		return err == nil && now.Before(expires)
	case kind == "" && lease.Labels[composeProjectLabel] != "":
		return now.Before(created.Add(legacyStackTTL))
	default:
		return lease.State == container.StateRunning || now.Before(created.Add(processGraceOf(lease.Labels[RunLabel])))
	}
}

func processGraceOf(run string) time.Duration {
	if grace, ok := processGraces[run]; ok {
		return grace
	}

	return processGrace
}

func (r *Run) remove(ctx context.Context, id string) error {
	var errs []error

	containers, err := r.client.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: labelEquals(RunLabel, id)})
	errs = append(errs, err)

	for _, c := range containers.Items {
		_, err := r.client.ContainerRemove(ctx, c.ID, client.ContainerRemoveOptions{Force: true, RemoveVolumes: true})
		errs = append(errs, err)
	}

	networks, err := r.client.NetworkList(ctx, client.NetworkListOptions{Filters: labelEquals(RunLabel, id)})
	errs = append(errs, err)

	for _, n := range networks.Items {
		_, err := r.client.NetworkRemove(ctx, n.ID, client.NetworkRemoveOptions{})
		errs = append(errs, err)
	}

	images, err := r.client.ImageList(ctx, client.ImageListOptions{Filters: labelEquals(RunLabel, id)})
	errs = append(errs, err)

	for _, i := range images.Items {
		refs := i.RepoTags
		if len(refs) == 0 {
			refs = []string{i.ID}
		}

		for _, ref := range refs {
			_, err := r.client.ImageRemove(ctx, ref, client.ImageRemoveOptions{})
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
