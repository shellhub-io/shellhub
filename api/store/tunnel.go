package store

import "context"

type TunnelStore interface {
	// TunnelUpdateDeviceUID changes all tunnels from oldUID to newUID within the specified tenantID.
	TunnelUpdateDeviceUID(ctx context.Context, tenantID, oldUID, newUID string) error
}
