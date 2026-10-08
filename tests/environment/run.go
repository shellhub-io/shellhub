package environment

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	log "github.com/sirupsen/logrus"
)

const (
	runLabel   = "io.shellhub.e2e.run"
	leaseLabel = "io.shellhub.e2e.lease"
	leaseImage = "alpine:3.24.2"

	composeProjectLabel = "com.docker.compose.project"
	composeServiceLabel = "com.docker.compose.service"
	runIDSize           = 12
)

var validStackProject = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,99}$`)

// Run owns the Docker objects one e2e run creates on a shared daemon: the images it builds and
// the containers and networks it starts, each labelled with the run's ID. A run is alive
// while a container carrying its lease label exists, in any state; every run sweeps the objects
// of dead runs when it starts.
type Run struct {
	id     string
	stack  bool
	client *client.Client
	lease  *client.HijackedResponse
	issuer *LicenseIssuer
}

// StartRun starts a run for the calling process. Its lease is a container whose stdin the
// process holds open, so the daemon releases the lease however the process ends. It sweeps the
// dead runs on the daemon once the lease exists, logging what it cannot remove. It returns an
// error when the daemon cannot be reached or the lease cannot be created. The lease outlives
// ctx; [Run.Close] releases it.
//
// A non-nil issuer makes the run compile issuer's public key into its enterprise server in place
// of the production license key, and start every enterprise stack under a license issuer signs
// rather than the one SHELLHUB_LICENSE_FILE names. A stack then picks its license with
// [Config.License] and runs without one under [Config.Unlicensed].
func StartRun(ctx context.Context, issuer *LicenseIssuer) (*Run, error) {
	return startRun(ctx, newRunID(), false, issuer)
}

// StartStackRun starts the run of a kept compose stack, identified by its project name. The
// stack's containers carry the lease, so the run stays alive after the calling process exits
// for as long as any of them exists. The calling process holds a lease of its own until it
// exits or calls [Run.Close], covering the objects it creates before the first stack
// container. It returns an error when project is not a valid image tag and project name, and
// in the cases [StartRun] does. issuer works as it does for [StartRun].
func StartStackRun(ctx context.Context, project string, issuer *LicenseIssuer) (*Run, error) {
	if !validStackProject.MatchString(project) {
		return nil, fmt.Errorf("invalid stack project %q: must match %s", project, validStackProject.String())
	}

	return startRun(ctx, project, true, issuer)
}

func startRun(ctx context.Context, id string, stack bool, issuer *LicenseIssuer) (*Run, error) {
	if err := pinDockerHost(); err != nil {
		return nil, err
	}

	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("new docker client: %w", err)
	}

	r := &Run{id: id, stack: stack, client: cli, issuer: issuer}

	if err := r.holdLease(ctx); err != nil {
		_ = cli.Close()

		return nil, err
	}

	r.sweep(ctx)

	return r, nil
}

func newRunID() string {
	return strings.ToLower(rand.Text()[:runIDSize])
}

// ID returns the run's ID, valid as an image tag and as a compose project name.
func (r *Run) ID() string { return r.id }

// LicenseIssuer returns the issuer [StartRun] or [StartStackRun] gave the run, or nil when the
// run's stacks run under the license SHELLHUB_LICENSE_FILE names.
func (r *Run) LicenseIssuer() *LicenseIssuer { return r.issuer }

// Labels returns the labels every container the run starts must carry. A stack run's include
// the lease, so its containers keep the run alive.
func (r *Run) Labels() map[string]string {
	labels := map[string]string{runLabel: r.id}
	if r.stack {
		labels[leaseLabel] = r.id
	}

	return labels
}

// Close removes every object labelled with the run, then releases the lease the process holds
// and closes the run's Docker client. It is best-effort: it returns every listing or removal
// that failed, which the next sweep retries once the run is dead, joined with any error closing
// the client.
// ctx bounds the removals; the lease is released whatever ctx's state.
func (r *Run) Close(ctx context.Context) error {
	err := r.remove(ctx, r.id)

	if r.lease != nil {
		r.lease.Close()
		r.lease = nil
	}

	return errors.Join(err, r.client.Close())
}

func (r *Run) composeEnvs() map[string]string {
	return map[string]string{
		"SHELLHUB_E2E_RUN":   r.id,
		"SHELLHUB_E2E_LEASE": r.Labels()[leaseLabel],
	}
}

func (r *Run) holdLease(ctx context.Context) error {
	if _, err := r.client.ImageInspect(ctx, leaseImage); err != nil {
		pull, err := r.client.ImagePull(ctx, leaseImage, client.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("pulling %s: %w", leaseImage, err)
		}

		if err := pull.Wait(ctx); err != nil {
			return fmt.Errorf("pulling %s: %w", leaseImage, err)
		}
	}

	created, err := r.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Image: leaseImage,
		Config: &container.Config{
			Cmd:         []string{"cat"},
			AttachStdin: true,
			OpenStdin:   true,
			StdinOnce:   true,
			Labels:      map[string]string{leaseLabel: r.id, runLabel: r.id},
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
		log.WithError(err).Warn("e2e sweep: could not list the runs on the daemon")

		return
	}

	for id := range ids {
		if id == r.id {
			continue
		}

		alive, err := r.alive(ctx, id)
		if err != nil {
			log.WithError(err).WithField("run", id).Warn("e2e sweep: could not read the run's lease")

			continue
		}

		if alive {
			continue
		}

		if err := r.remove(ctx, id); err != nil {
			log.WithError(err).WithField("run", id).Warn("e2e sweep: left objects of a dead run behind")
		}
	}
}

func runLabelled() client.Filters {
	return make(client.Filters).Add("label", runLabel)
}

func labelEquals(label, value string) client.Filters {
	return make(client.Filters).Add("label", label+"="+value)
}

func (r *Run) labelledRuns(ctx context.Context) (map[string]struct{}, error) {
	ids := make(map[string]struct{})
	collect := func(labels map[string]string) {
		if id := labels[runLabel]; id != "" {
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
	leases, err := r.client.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: labelEquals(leaseLabel, id)})
	if err != nil {
		return false, err
	}

	return len(leases.Items) > 0, nil
}

func (r *Run) remove(ctx context.Context, id string) error {
	var errs []error

	containers, err := r.client.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: labelEquals(runLabel, id)})
	errs = append(errs, err)

	for _, c := range containers.Items {
		_, err := r.client.ContainerRemove(ctx, c.ID, client.ContainerRemoveOptions{Force: true, RemoveVolumes: true})
		errs = append(errs, err)
	}

	networks, err := r.client.NetworkList(ctx, client.NetworkListOptions{Filters: labelEquals(runLabel, id)})
	errs = append(errs, err)

	for _, n := range networks.Items {
		_, err := r.client.NetworkRemove(ctx, n.ID, client.NetworkRemoveOptions{})
		errs = append(errs, err)
	}

	images, err := r.client.ImageList(ctx, client.ImageListOptions{Filters: labelEquals(runLabel, id)})
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

func (r *Run) requireLabels(ctx context.Context, project string) error {
	containers, err := r.client.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: labelEquals(composeProjectLabel, project)})
	if err != nil {
		return err
	}

	for _, c := range containers.Items {
		for label, value := range r.Labels() {
			if c.Labels[label] != value {
				return fmt.Errorf("service %s carries no %s=%s label", c.Labels[composeServiceLabel], label, value)
			}
		}
	}

	return nil
}

func (r *Run) createNetwork(ctx context.Context, name string) error {
	_, err := r.client.NetworkCreate(ctx, name, client.NetworkCreateOptions{Labels: map[string]string{runLabel: r.id}})

	return err
}

func (r *Run) removeNetwork(ctx context.Context, name string) error {
	_, err := r.client.NetworkRemove(ctx, name, client.NetworkRemoveOptions{})

	return err
}
