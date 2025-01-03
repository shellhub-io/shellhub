package connector

import (
	"context"
)

// ConnectorVersion stores the version of the ShellHub Instane that is running the connector.
// It is used in the ShellHub Agents initialized by the connector when a container is started.
var ConnectorVersion string

// Container is a struct that represents a container that will be managed by the connector.
type Container struct {
	Cancel        context.CancelFunc
	ID            string
	Name          string
	ServerAddress string
	Tenant        string
	PrivateKey    string
}

// Connector is an interface that defines the methods that a connector must implement.
type Connector interface {
	// List lists all containers running on the host.
	List(ctx context.Context) ([]Container, error)
	// Start starts the agent for the container with the given ID.
	Start(ctx context.Context, id string, name string)
	// Stop stops the agent for the container with the given ID.
	Stop(ctx context.Context, id string)
	// Listen listens for events and starts or stops the agent for the container that was created or removed.
	Listen(ctx context.Context) error
}
