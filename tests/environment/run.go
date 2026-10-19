package environment

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/shellhub-io/shellhub/pkg/testrun"
)

const (
	runLabel = testrun.RunLabel

	composeProjectLabel = "com.docker.compose.project"
	composeServiceLabel = "com.docker.compose.service"
)

// Run is the [testrun.Run] of an e2e run, with the license issuer its enterprise stacks run under
// and the Docker client its stacks and images use. The images pulled from or pushed to an image
// repository by [BuildEditionImages] also keep their repository tag, which carries no run, so no
// run removes it: the editions of a stack share one daemon and pull or push the same tag at once.
type Run struct {
	run    *testrun.Run
	client *client.Client
	issuer *LicenseIssuer

	logsMu     sync.Mutex
	logsOpened map[string]bool
}

func (r *Run) keepProjectLogs(ctx context.Context, project string, path string) (int, error) {
	containers, err := r.client.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: labelEquals(composeProjectLabel, project)})
	if err != nil {
		return 0, err
	}

	if len(containers.Items) == 0 {
		return 0, nil
	}

	file, err := r.openStackLog(path)
	if err != nil {
		return 0, err
	}

	defer func() { _ = file.Close() }()

	var (
		kept int
		errs []error
	)

	for _, c := range containers.Items {
		service := c.Labels[composeServiceLabel]

		if err := r.writeContainerLogs(ctx, file, service, c.ID); err != nil {
			errs = append(errs, fmt.Errorf("logs for %s: %w", service, err))

			continue
		}

		kept++
	}

	return kept, errors.Join(errs...)
}

func (r *Run) writeContainerLogs(ctx context.Context, w io.Writer, service, id string) error {
	if _, err := fmt.Fprintf(w, "===== %s =====\n", service); err != nil {
		return err
	}

	logs, err := r.client.ContainerLogs(ctx, id, client.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return err
	}

	defer func() { _ = logs.Close() }()

	_, err = stdcopy.StdCopy(w, w, logs)

	return err
}

func (r *Run) openStackLog(path string) (*os.File, error) {
	r.logsMu.Lock()
	defer r.logsMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, err
	}

	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if r.logsOpened[path] {
		flags = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}

	file, err := os.OpenFile(path, flags, 0o600) //nolint:gosec // the path is a test's own name under the tests directory
	if err != nil {
		return nil, err
	}

	if r.logsOpened == nil {
		r.logsOpened = make(map[string]bool)
	}

	r.logsOpened[path] = true

	return file, nil
}

// StartRun starts a run for the calling process with [testrun.Start]. It returns an error when the
// Docker host cannot be resolved, the Docker client cannot be built, and in the cases
// [testrun.Start] does.
//
// A non-nil issuer makes the run compile issuer's public key into its enterprise server in place
// of the production license key, and start every enterprise stack under a license issuer signs. A
// stack then picks its license with [Config.License] and runs without one under
// [Config.Unlicensed]. A run with a nil issuer brings up an enterprise or cloud stack only under
// [Config.Unlicensed].
func StartRun(ctx context.Context, issuer *LicenseIssuer) (*Run, error) {
	return startRun(issuer, func() (*testrun.Run, error) { return testrun.Start(ctx) })
}

// StartStackRun starts the run of a kept compose stack, identified by its project name, with
// [testrun.StartStack] and opts. It returns an error in the cases [StartRun] and
// [testrun.StartStack] do. issuer works as it does for [StartRun].
func StartStackRun(ctx context.Context, project string, opts testrun.StackOptions, issuer *LicenseIssuer) (*Run, error) {
	return startRun(issuer, func() (*testrun.Run, error) { return testrun.StartStack(ctx, project, opts) })
}

func startRun(issuer *LicenseIssuer, start func() (*testrun.Run, error)) (*Run, error) {
	if err := pinDockerHost(); err != nil {
		return nil, err
	}

	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("new docker client: %w", err)
	}

	run, err := start()
	if err != nil {
		_ = cli.Close()

		return nil, err
	}

	return &Run{run: run, client: cli, issuer: issuer}, nil
}

// ID returns the run's ID, valid as an image tag and as a compose project name.
func (r *Run) ID() string { return r.run.ID() }

// LicenseIssuer returns the issuer [StartRun] or [StartStackRun] gave the run, or nil when the run
// issues no licenses.
func (r *Run) LicenseIssuer() *LicenseIssuer { return r.issuer }

// Labels returns the labels every container the run starts must carry: the run's ID, and on a stack
// run the lease, kind and expiry that keep it alive, and its checkout when one was set.
func (r *Run) Labels() map[string]string { return r.run.Labels() }

// Close removes every object labelled with the run and releases its lease, as [testrun.Run.Close]
// does, then closes the run's Docker client. It returns those errors joined.
func (r *Run) Close(ctx context.Context) error {
	return errors.Join(r.run.Close(ctx), r.client.Close())
}

func (r *Run) composeEnvs() map[string]string {
	labels := r.Labels()

	return map[string]string{
		"SHELLHUB_E2E_RUN":      labels[testrun.RunLabel],
		"SHELLHUB_E2E_LEASE":    labels[testrun.LeaseLabel],
		"SHELLHUB_E2E_KIND":     labels[testrun.KindLabel],
		"SHELLHUB_E2E_EXPIRES":  labels[testrun.ExpiresLabel],
		"SHELLHUB_E2E_CHECKOUT": labels[testrun.CheckoutLabel],
	}
}

func (r *Run) unhealthyServices(ctx context.Context, project string) error {
	containers, err := r.client.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: labelEquals(composeProjectLabel, project)})
	if err != nil {
		return err
	}

	var errs []error

	for _, c := range containers.Items {
		inspected, err := r.client.ContainerInspect(ctx, c.ID, client.ContainerInspectOptions{})
		if err != nil {
			errs = append(errs, err)

			continue
		}

		service := c.Labels[composeServiceLabel]
		state := inspected.Container.State

		switch {
		case state.Status == container.StateExited:
			errs = append(errs, fmt.Errorf("service %s exited with code %d", service, state.ExitCode))
		case state.Status != container.StateRunning:
			errs = append(errs, fmt.Errorf("service %s is %s", service, state.Status))
		case state.Health != nil && state.Health.Status != container.Healthy:
			var checks []string
			for _, check := range state.Health.Log {
				checks = append(checks, fmt.Sprintf("exit %d: %s", check.ExitCode, strings.TrimSpace(check.Output)))
			}

			errs = append(errs, fmt.Errorf("service %s is %s; its last health checks: %s", service, state.Health.Status, strings.Join(checks, " | ")))
		}
	}

	return errors.Join(errs...)
}

func labelEquals(label, value string) client.Filters {
	return make(client.Filters).Add("label", label+"="+value)
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
	_, err := r.client.NetworkCreate(ctx, name, client.NetworkCreateOptions{Labels: map[string]string{runLabel: r.ID()}})

	return err
}

func (r *Run) removeNetwork(ctx context.Context, name string) error {
	_, err := r.client.NetworkRemove(ctx, name, client.NetworkRemoveOptions{})

	return err
}
