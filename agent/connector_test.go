package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	dockerclient "github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

const (
	vanishedContainerID = "0000000000000000vanished"
	liveContainerID     = "1111111111111111live"
	laterContainerID    = "2222222222222222later"
	hangingContainerID  = "3333333333333333hanging"
)

var (
	errNoSuchContainer = errors.New("no such container")
	errEventsStream    = errors.New("events stream closed by the daemon")
)

type fakeDockerDaemon struct {
	dockerclient.APIClient
	running   []container.Summary
	names     map[string]string
	events    chan events.Message
	errs      chan error
	inspected chan string
}

func newFakeDockerDaemon(running ...string) *fakeDockerDaemon {
	summaries := make([]container.Summary, len(running))
	for i, id := range running {
		summaries[i] = container.Summary{ID: id}
	}

	return &fakeDockerDaemon{
		running: summaries,
		names: map[string]string{
			liveContainerID:  "live",
			laterContainerID: "later",
		},
		events:    make(chan events.Message, 8),
		errs:      make(chan error, 2),
		inspected: make(chan string, 16),
	}
}

func (f *fakeDockerDaemon) ContainerList(context.Context, container.ListOptions) ([]container.Summary, error) {
	return f.running, nil
}

func (f *fakeDockerDaemon) Events(ctx context.Context, _ events.ListOptions) (<-chan events.Message, <-chan error) {
	go func() {
		<-ctx.Done()
		select {
		case f.errs <- ctx.Err():
		default:
		}
	}()

	return f.events, f.errs
}

func (f *fakeDockerDaemon) ContainerInspect(ctx context.Context, id string) (container.InspectResponse, error) {
	if err := ctx.Err(); err != nil {
		return container.InspectResponse{}, err
	}

	select {
	case f.inspected <- id:
	case <-ctx.Done():
		return container.InspectResponse{}, ctx.Err()
	}

	if id == hangingContainerID {
		<-ctx.Done()
		f.errs <- ctx.Err()

		return container.InspectResponse{}, ctx.Err()
	}

	name, ok := f.names[id]
	if !ok {
		return container.InspectResponse{}, errNoSuchContainer
	}

	return container.InspectResponse{ContainerJSONBase: &container.ContainerJSONBase{Name: "/" + name}}, nil
}

func (f *fakeDockerDaemon) start(id string) {
	f.events <- events.Message{Action: "start", Actor: events.Actor{ID: id}}
}

func listen(t *testing.T, daemon *fakeDockerDaemon) (context.CancelFunc, <-chan error) {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	return cancel, listenUntil(t, ctx, daemon)
}

func listenUntil(t *testing.T, ctx context.Context, daemon *fakeDockerDaemon) <-chan error {
	t.Helper()

	connector := &DockerConnector{
		cli:         daemon,
		privateKeys: t.TempDir(),
		cancels:     make(map[string]context.CancelFunc),
	}

	done := make(chan error, 1)
	go func() { done <- connector.Listen(ctx) }()

	return done
}

func requireListenReturned(t *testing.T, done <-chan error) error {
	t.Helper()

	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("Listen kept running")

		return nil
	}
}

func requireInspected(t *testing.T, daemon *fakeDockerDaemon, done <-chan error, want string) {
	t.Helper()

	for {
		select {
		case id := <-daemon.inspected:
			if id == want {
				return
			}
		case err := <-done:
			t.Fatalf("Listen returned %v before inspecting %s", err, want)
		case <-time.After(5 * time.Second):
			t.Fatalf("Listen never inspected %s", want)
		}
	}
}

func TestDockerConnectorListen(t *testing.T) {
	t.Run("a start event for a container that vanished leaves Listen running", func(t *testing.T) {
		daemon := newFakeDockerDaemon()
		cancel, done := listen(t, daemon)

		daemon.start(vanishedContainerID)
		daemon.start(liveContainerID)
		requireInspected(t, daemon, done, liveContainerID)

		cancel()
		require.NoError(t, requireListenReturned(t, done))
	})

	t.Run("a listed container that vanishes before its inspect leaves Listen running", func(t *testing.T) {
		daemon := newFakeDockerDaemon(vanishedContainerID, liveContainerID)
		cancel, done := listen(t, daemon)

		requireInspected(t, daemon, done, liveContainerID)
		daemon.start(laterContainerID)
		requireInspected(t, daemon, done, laterContainerID)

		cancel()
		require.NoError(t, requireListenReturned(t, done))
	})

	t.Run("an events stream error ends Listen with that error", func(t *testing.T) {
		daemon := newFakeDockerDaemon()
		_, done := listen(t, daemon)

		daemon.errs <- errEventsStream

		require.ErrorIs(t, requireListenReturned(t, done), errEventsStream)
	})

	t.Run("a cancel while Listen inspects a container ends Listen with nil", func(t *testing.T) {
		daemon := newFakeDockerDaemon()
		cancel, done := listen(t, daemon)

		daemon.start(hangingContainerID)
		requireInspected(t, daemon, done, hangingContainerID)

		cancel()
		require.NoError(t, requireListenReturned(t, done))
	})

	t.Run("a context cancelled before the startup list ends Listen with nil", func(t *testing.T) {
		daemon := newFakeDockerDaemon(liveContainerID)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		require.NoError(t, requireListenReturned(t, listenUntil(t, ctx, daemon)))
	})
}
