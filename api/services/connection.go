package services

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

type ConnectionService interface {
	// CreateConnection creates a direct connection in the given namespace.
	CreateConnection(ctx context.Context, req *requests.ConnectionCreate) (*models.Connection, error)
	// ListConnections lists the connections in the given namespace.
	ListConnections(ctx context.Context, req *requests.ConnectionList) ([]models.Connection, int, error)
	// UpdateConnection updates a connection scoped to the namespace.
	UpdateConnection(ctx context.Context, req *requests.ConnectionUpdate) (*models.Connection, error)
	// GetConnection returns a single connection scoped to the namespace.
	GetConnection(ctx context.Context, req *requests.ConnectionGet) (*models.Connection, error)
	// DeleteConnection deletes a connection scoped to the namespace.
	DeleteConnection(ctx context.Context, req *requests.ConnectionDelete) error
	// ConnectionStatus reports whether the connection's target is reachable. For
	// direct connections it probes the host:port over TCP; for device connections
	// it reflects the device's connection state.
	ConnectionStatus(ctx context.Context, req *requests.ConnectionGet) (bool, error)
	// ProbeReachability reports whether an arbitrary host:port is reachable over
	// TCP. Used before saving a direct connection to surface NAT/firewall issues.
	ProbeReachability(ctx context.Context, req *requests.ConnectionProbe) (bool, error)
}

// probeTCP reports whether host:port accepts a TCP connection within a short
// timeout. A failure means unreachable, not an error.
func probeTCP(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), 4*time.Second)
	if err != nil {
		return false
	}
	conn.Close() //nolint:errcheck

	return true
}

func (s *service) CreateConnection(ctx context.Context, req *requests.ConnectionCreate) (*models.Connection, error) {
	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	connection := &models.Connection{
		ID:       uuid.Generate(),
		TenantID: req.TenantID,
		Label:    req.Label,
		Username: req.Username,
		Kind:     models.ConnectionKind(req.Kind),
	}
	applyTarget(connection, req.Kind, req.Host, req.Port, req.DeviceUID)

	if _, err := s.store.ConnectionCreate(ctx, connection); err != nil {
		return nil, err
	}

	return connection, nil
}

// applyTarget sets the target fields on the connection based on its kind,
// clearing the ones that don't apply so rows stay clean.
func applyTarget(c *models.Connection, kind, host string, port int, deviceUID string) {
	switch models.ConnectionKind(kind) {
	case models.ConnectionKindDevice:
		c.DeviceUID = deviceUID
		c.Host = ""
		c.Port = 0
	default:
		c.Host = host
		c.Port = port
		c.DeviceUID = ""
	}
}

func (s *service) UpdateConnection(ctx context.Context, req *requests.ConnectionUpdate) (*models.Connection, error) {
	connection, err := s.store.ConnectionResolve(ctx, store.ConnectionIDResolver, req.ID, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return nil, NewErrConnectionNotFound(req.ID, err)
		}

		return nil, err
	}

	connection.Label = req.Label
	connection.Username = req.Username
	connection.Kind = models.ConnectionKind(req.Kind)
	applyTarget(connection, req.Kind, req.Host, req.Port, req.DeviceUID)

	if err := s.store.ConnectionUpdate(ctx, connection); err != nil {
		return nil, err
	}

	return connection, nil
}

func (s *service) ListConnections(ctx context.Context, req *requests.ConnectionList) ([]models.Connection, int, error) {
	if req.Sorter.By == "" {
		req.Sorter.By = "created_at"
	}

	req.Sorter.Tiebreak = "id"

	return s.store.ConnectionList(
		ctx,
		s.store.Options().InNamespace(req.TenantID),
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	)
}

func (s *service) GetConnection(ctx context.Context, req *requests.ConnectionGet) (*models.Connection, error) {
	connection, err := s.store.ConnectionResolve(ctx, store.ConnectionIDResolver, req.ID, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return nil, NewErrConnectionNotFound(req.ID, err)
		}

		return nil, err
	}

	return connection, nil
}

func (s *service) DeleteConnection(ctx context.Context, req *requests.ConnectionDelete) error {
	connection, err := s.store.ConnectionResolve(ctx, store.ConnectionIDResolver, req.ID, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrConnectionNotFound(req.ID, err)
		}

		return err
	}

	return s.store.ConnectionDelete(ctx, connection)
}

func (s *service) ConnectionStatus(ctx context.Context, req *requests.ConnectionGet) (bool, error) {
	connection, err := s.store.ConnectionResolve(ctx, store.ConnectionIDResolver, req.ID, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return false, NewErrConnectionNotFound(req.ID, err)
		}

		return false, err
	}

	// Device connections reflect the agent's connection state. A resolve failure
	// just means we can't confirm it's up, so report offline.
	if connection.Kind == models.ConnectionKindDevice {
		if device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, connection.DeviceUID); err == nil {
			return device.DisconnectedAt == nil, nil
		}

		return false, nil
	}

	// Direct connections have no agent, so probe the endpoint over TCP.
	return probeTCP(connection.Host, connection.Port), nil
}

func (s *service) ProbeReachability(_ context.Context, req *requests.ConnectionProbe) (bool, error) {
	return probeTCP(req.Host, req.Port), nil
}
