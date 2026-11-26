package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/shellhub-io/shellhub/agent/pkg/agentd"
	"github.com/shellhub-io/shellhub/agent/pkg/connector"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/validator"
	log "github.com/sirupsen/logrus"
)

var _ connector.Connector = new(DockerConnector)

// DockerConnector is a struct that represents a connector that uses Docker as the container runtime.
type DockerConnector struct {
	mu              sync.Mutex
	server          string
	tenant          string
	provisioningKey string
	cli             dockerclient.APIClient
	privateKeys     string
	// Label is the label used to identify the containers managed by the ShellHub agent.
	Label   string
	cancels map[string]context.CancelFunc
}

// ConfigConnector provides the configuration for the agent connector service.
type ConfigConnector struct {
	// Set the ShellHub server address the agent will use to connect.
	// This is required.
	ServerAddress string `env:"SERVER_ADDRESS,required"`

	// Specify the path to store the devices/containers private keys.
	// If not provided, the agent will generate a new one.
	// This is required.
	PrivateKeys string `env:"PRIVATE_KEYS,required"`

	// Sets the account tenant id used during communication to associate the
	// devices to a specific tenant.
	TenantID string `env:"TENANT_ID" validate:"required_without=ProvisioningKey"`

	ProvisioningKey string `env:"PROVISIONING_KEY" validate:"required_without=TenantID"`

	// Determine the interval to send the keep alive message to the server. This
	// has a direct impact of the bandwidth used by the device when in idle
	// state. Default is 30 seconds.
	KeepAliveInterval int `env:"KEEPALIVE_INTERVAL,overwrite,default=30"`

	// Label is the label used to identify the containers managed by the ShellHub agent.
	Label string `env:"CONNECTOR_LABEL,default="`
}

func LoadConfigConnectorFromEnv(backend envs.Backend) (*ConfigConnector, map[string]any, error) {
	cfg, err := envs.ParseWithPrefix[ConfigConnector](backend, "SHELLHUB_")
	if err != nil {
		log.Fatal(err)
	}

	if fields, err := validator.New().StructWithFields(cfg); err != nil {
		log.WithFields(fields).Error("failed to validate the configuration loaded from envs")

		return nil, fields, err
	}

	return cfg, nil, nil
}

// NewDockerConnector creates a new [Connector] that uses Docker as the container runtime.
func NewDockerConnector(config *ConfigConnector) (connector.Connector, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerConnector{
		cli:             cli,
		server:          config.ServerAddress,
		tenant:          config.TenantID,
		provisioningKey: config.ProvisioningKey,
		privateKeys:     config.PrivateKeys,
		Label:           config.Label,
		cancels:         make(map[string]context.CancelFunc),
	}, nil
}

func (d *DockerConnector) events(ctx context.Context) (<-chan events.Message, <-chan error) {
	filters := filters.NewArgs()
	if d.Label != "" {
		filters.Add("label", d.Label)
	}

	return d.cli.Events(ctx, events.ListOptions{
		Filters: filters,
	})
}

// List returns the running containers that carry the connector's label, or every running
// container when no label is set. A container that cannot be inspected, usually because it was
// removed after the daemon listed it, is logged and left out. List returns the error the daemon
// gives when it cannot list containers, and ctx's error once ctx is cancelled, without logging
// the containers it did not reach.
func (d *DockerConnector) List(ctx context.Context) ([]connector.Container, error) {
	filters := filters.NewArgs()
	if d.Label != "" {
		filters.Add("label", d.Label)
	}

	containers, err := d.cli.ContainerList(ctx, container.ListOptions{
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}

	list := make([]connector.Container, 0, len(containers))
	for _, container := range containers {
		name, err := d.getContainerNameFromID(ctx, container.ID)
		if err != nil {
			if cancelled(ctx) {
				return nil, ctx.Err()
			}

			log.WithError(err).WithField("id", container.ID).
				Warn("Skipping a listed container the connector could not inspect")

			continue
		}

		list = append(list, connector.Container{ID: container.ID, Name: name})
	}

	return list, nil
}

// Start starts the agent for the container with the given ID.
func (d *DockerConnector) Start(ctx context.Context, id string, name string) {
	id = id[:12]

	d.mu.Lock()
	ctx, d.cancels[id] = context.WithCancel(ctx)
	d.mu.Unlock()

	privateKey := containerKeyPath(d.privateKeys, id)
	go initContainerAgent(ctx, d.cli, connector.Container{
		ID:              id,
		Name:            name,
		ServerAddress:   d.server,
		Tenant:          d.tenant,
		ProvisioningKey: d.provisioningKey,
		PrivateKey:      privateKey,
	})
}

// Stop stops the agent for the container with the given ID.
func (d *DockerConnector) Stop(_ context.Context, id string) {
	id = id[:12]

	d.mu.Lock()
	defer d.mu.Unlock()

	cancel, ok := d.cancels[id]
	if ok {
		cancel()
		delete(d.cancels, id)
	}
}

func containerKeyPath(dir, id string) string {
	key := filepath.Join(dir, id+".key")
	if mightExist(key) {
		return key
	}

	legacy := filepath.Join(filepath.Dir(dir), id+".key")
	if mightExist(legacy) {
		log.WithFields(log.Fields{"id": id, "path": legacy, "directory": dir}).
			Warn("Using a container key from the parent of SHELLHUB_PRIVATE_KEYS, where earlier versions stored it")

		return legacy
	}

	return key
}

func mightExist(path string) bool {
	_, err := os.Stat(path)

	return !errors.Is(err, fs.ErrNotExist)
}

func cancelled(ctx context.Context) bool {
	return ctx.Err() != nil
}

func (d *DockerConnector) getContainerNameFromID(ctx context.Context, id string) (string, error) {
	container, err := d.cli.ContainerInspect(ctx, id)
	if err != nil {
		return "", err
	}

	name := container.Name[1:]

	name = strings.ReplaceAll(name, ".", "_")
	if len(name) > 64 {
		name = name[:64]
	}

	return name, nil
}

// Listen starts an agent for each container [DockerConnector.List] returns, then follows the
// daemon's events, starting an agent when a container starts and stopping it when the container
// dies. A started container that cannot be inspected is logged and skipped. Listen returns nil
// once ctx is cancelled, even when the daemon reports the cancellation as an error. Otherwise it
// returns the error from [DockerConnector.List] or the error the events stream reports.
func (d *DockerConnector) Listen(ctx context.Context) error {
	containers, err := d.List(ctx)
	if cancelled(ctx) {
		return nil
	}

	if err != nil {
		return err
	}

	for _, container := range containers {
		d.Start(ctx, container.ID, container.Name)
	}

	events, errs := d.events(ctx)
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errs:
			if cancelled(ctx) {
				return nil
			}

			return err
		case container := <-events:
			switch container.Action {
			case "start":
				name, err := d.getContainerNameFromID(ctx, container.Actor.ID)
				if cancelled(ctx) {
					return nil
				}

				if err != nil {
					log.WithError(err).WithField("id", container.Actor.ID).
						Warn("Skipping a started container the connector could not inspect")

					continue
				}

				d.Start(ctx, container.Actor.ID, name)
			case "die":
				d.Stop(ctx, container.Actor.ID)
			default:
			}
		}
	}
}

func (d *DockerConnector) CheckUpdate() (*semver.Version, error) {
	api, err := client.NewClient(d.server)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"version": connector.ConnectorVersion,
		}).Error("Failed to create HTTP client to check agent version")

		return nil, err
	}

	info, err := api.GetInfo(connector.ConnectorVersion)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"version": connector.ConnectorVersion,
		}).Error("Failed to get info from ShellHub's server")

		return nil, err
	}

	return semver.NewVersion(info.Version)
}

func initContainerAgent(ctx context.Context, cli dockerclient.APIClient, container connector.Container) {
	cfg := &agentd.Config{
		ServerAddress:     container.ServerAddress,
		TenantID:          container.Tenant,
		ProvisioningKey:   container.ProvisioningKey,
		PrivateKey:        container.PrivateKey,
		PreferredIdentity: container.ID,
		PreferredHostname: container.Name,
		KeepAliveInterval: 30,
		TransportVersion:  agentd.TransportV2,
		Version:           connector.ConnectorVersion,
		Platform:          "connector",
	}

	log.WithFields(log.Fields{
		"id":             container.ID,
		"identity":       cfg.PreferredIdentity,
		"hostname":       cfg.PreferredHostname,
		"tenant_id":      cfg.TenantID,
		"server_address": cfg.ServerAddress,
		"timestamp":      time.Now(), //nolint:forbidigo // a log line's timestamp is wall-clock time
		"version":        cfg.Version,
	}).Info("Connector container started")

	mode, err := agentd.NewConnectorMode(cli, container.ID)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"id":             container.ID,
			"identity":       cfg.PreferredIdentity,
			"hostname":       cfg.PreferredHostname,
			"tenant_id":      cfg.TenantID,
			"server_address": cfg.ServerAddress,
			"timestamp":      time.Now(), //nolint:forbidigo // a log line's timestamp is wall-clock time
			"version":        cfg.Version,
		}).Error("Failed to create connector mode")

		return
	}

	ag, err := agentd.NewAgentWithConfig(cfg, mode)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"id":             container.ID,
			"identity":       cfg.PreferredIdentity,
			"hostname":       cfg.PreferredHostname,
			"tenant_id":      cfg.TenantID,
			"server_address": cfg.ServerAddress,
			"version":        cfg.Version,
		}).Error("Failed to create agent")

		return
	}

	if err := ag.Initialize(ctx); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"id":             container.ID,
			"identity":       cfg.PreferredIdentity,
			"hostname":       cfg.PreferredHostname,
			"tenant_id":      cfg.TenantID,
			"server_address": cfg.ServerAddress,
			"version":        cfg.Version,
		}).Error("Failed to initialize agent")

		return
	}

	log.WithFields(log.Fields{
		"id":             container.ID,
		"identity":       cfg.PreferredIdentity,
		"hostname":       cfg.PreferredHostname,
		"tenant_id":      cfg.TenantID,
		"server_address": cfg.ServerAddress,
		"timestamp":      time.Now(), //nolint:forbidigo // a log line's timestamp is wall-clock time
		"version":        cfg.Version,
	}).Info("Listening for connections")

	if err := ag.Listen(ctx); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"id":             container.ID,
			"identity":       cfg.PreferredIdentity,
			"hostname":       cfg.PreferredHostname,
			"tenant_id":      cfg.TenantID,
			"server_address": cfg.ServerAddress,
			"timestamp":      time.Now(), //nolint:forbidigo // a log line's timestamp is wall-clock time
			"version":        cfg.Version,
		}).Error("Failed to listen for connections")

		return
	}

	log.WithFields(log.Fields{
		"id":             container.ID,
		"identity":       cfg.PreferredIdentity,
		"hostname":       cfg.PreferredHostname,
		"tenant_id":      cfg.TenantID,
		"server_address": cfg.ServerAddress,
		"version":        cfg.Version,
	}).Info("Connector container done")
}
