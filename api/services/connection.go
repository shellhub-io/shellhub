package services

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/egress"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

type ConnectionService interface {
	// CreateConnection creates a personal connection owned by the caller.
	CreateConnection(ctx context.Context, req *requests.ConnectionCreate) (*models.Connection, error)
	// ListConnections lists the connections owned by the caller.
	ListConnections(ctx context.Context, req *requests.ConnectionList) ([]models.Connection, int, error)
	// UpdateConnection updates a connection owned by the caller.
	UpdateConnection(ctx context.Context, req *requests.ConnectionUpdate) (*models.Connection, error)
	// GetConnection returns a single connection owned by the caller.
	GetConnection(ctx context.Context, req *requests.ConnectionGet) (*models.Connection, error)
	// DeleteConnection deletes a connection owned by the caller.
	DeleteConnection(ctx context.Context, req *requests.ConnectionDelete) error
	// ConnectionStatus reports whether the connection's target is reachable. For
	// external connections it probes the host:port over TCP; for device connections
	// it reflects the device's connection state.
	ConnectionStatus(ctx context.Context, req *requests.ConnectionGet) (bool, error)
	// ProbeReachability reports whether an arbitrary host:port is reachable over
	// TCP. Used before saving an external connection to surface NAT/firewall issues.
	ProbeReachability(ctx context.Context, req *requests.ConnectionProbe) (bool, error)
}

// ErrEgressBlocked means the target isn't a permitted connection endpoint: the
// SSRF guardian rejected it (loopback, link-local/metadata, reserved, or a
// private address that isn't allowlisted). Distinct from a host that is simply
// unreachable. It aliases egress.ErrBlocked so route checks (errors.Is) keep working.
var ErrEgressBlocked = egress.ErrBlocked

// resolveOwnedConnection fetches a connection owned by the caller. A connection
// owned by someone else (or absent) resolves to ErrConnectionNotFound, so we
// never leak the existence of another user's connection.
func (s *service) resolveOwnedConnection(ctx context.Context, tenantID, userID, id string) (*models.Connection, error) {
	// id maps to a uuid-typed column; a non-UUID can never match, so short-circuit
	// to not-found instead of letting the cast raise a SQL error (500).
	if _, err := uuid.Parse(id); err != nil {
		return nil, NewErrConnectionNotFound(id, err)
	}

	connection, err := s.store.ConnectionResolve(
		ctx,
		store.ConnectionIDResolver,
		id,
		s.store.Options().InNamespace(tenantID),
		s.store.Options().OwnedBy(userID),
	)
	if err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return nil, NewErrConnectionNotFound(id, err)
		}

		return nil, err
	}

	return connection, nil
}

// validateDeviceTarget ensures a device-kind connection points at a device that
// exists in the caller's namespace. Without this, a connection could reference a
// device in another namespace and leak its status through ConnectionStatus.
func (s *service) validateDeviceTarget(ctx context.Context, tenantID, kind, deviceUID string) error {
	if models.ConnectionKind(kind) != models.ConnectionKindDevice {
		return nil
	}

	_, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, deviceUID, s.store.Options().InNamespace(tenantID))
	if err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrDeviceNotFound(models.UID(deviceUID), err)
		}

		return err
	}

	return nil
}

func (s *service) CreateConnection(ctx context.Context, req *requests.ConnectionCreate) (*models.Connection, error) {
	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	if err := s.validateDeviceTarget(ctx, req.TenantID, req.Kind, req.DeviceUID); err != nil {
		return nil, err
	}

	connection := &models.Connection{
		ID:             uuid.Generate(),
		TenantID:       req.TenantID,
		OwnerID:        req.UserID,
		Label:          req.Label,
		Kind:           models.ConnectionKind(req.Kind),
		Username:       req.Username,
		AuthMethod:     req.AuthMethod,
		KeyFingerprint: req.KeyFingerprint,
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
	connection, err := s.resolveOwnedConnection(ctx, req.TenantID, req.UserID, req.ID)
	if err != nil {
		return nil, err
	}

	// The target kind is fixed at creation; an external host and a device are
	// distinct target shapes, so changing it on update is rejected.
	if models.ConnectionKind(req.Kind) != connection.Kind {
		return nil, NewErrConnectionKindImmutable()
	}

	if err := s.validateDeviceTarget(ctx, req.TenantID, req.Kind, req.DeviceUID); err != nil {
		return nil, err
	}

	connection.Label = req.Label
	connection.Kind = models.ConnectionKind(req.Kind)
	connection.Username = req.Username
	connection.AuthMethod = req.AuthMethod
	connection.KeyFingerprint = req.KeyFingerprint
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
		s.store.Options().OwnedBy(req.UserID),
		s.store.Options().Sort(&req.Sorter),
		s.store.Options().Paginate(&req.Paginator),
	)
}

func (s *service) GetConnection(ctx context.Context, req *requests.ConnectionGet) (*models.Connection, error) {
	return s.resolveOwnedConnection(ctx, req.TenantID, req.UserID, req.ID)
}

func (s *service) DeleteConnection(ctx context.Context, req *requests.ConnectionDelete) error {
	connection, err := s.resolveOwnedConnection(ctx, req.TenantID, req.UserID, req.ID)
	if err != nil {
		return err
	}

	return s.store.ConnectionDelete(ctx, connection)
}

func (s *service) ConnectionStatus(ctx context.Context, req *requests.ConnectionGet) (bool, error) {
	connection, err := s.resolveOwnedConnection(ctx, req.TenantID, req.UserID, req.ID)
	if err != nil {
		return false, err
	}

	// Device connections reflect the agent's connection state. A resolve failure
	// just means we can't confirm it's up, so report offline.
	if connection.Kind == models.ConnectionKindDevice {
		device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, connection.DeviceUID, s.store.Options().InNamespace(connection.TenantID))
		if err == nil {
			return device.DisconnectedAt == nil, nil
		}

		return false, nil
	}

	// External connections have no agent, so probe the endpoint over TCP.
	return egress.Reachable(ctx, connection.Host, connection.Port), nil
}

func (s *service) ProbeReachability(ctx context.Context, req *requests.ConnectionProbe) (bool, error) {
	reachable, blocked := egress.Probe(ctx, req.Host, req.Port)
	if blocked {
		return false, ErrEgressBlocked
	}

	return reachable, nil
}
