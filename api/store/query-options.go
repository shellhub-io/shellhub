package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type NamespaceQueryOption func(ctx context.Context, ns *models.Namespace) error

type QueryOption func(ctx context.Context) error

type QueryOptions interface {
	// InNamespace matches a document that belongs to the provided namespace
	InNamespace(tenantID string) QueryOption

	// WithDeviceStatus matches a device with the provided status
	WithDeviceStatus(models.DeviceStatus) QueryOption

	// CountAcceptedDevices counts the devices with a status 'accepted'
	// in the namespace.
	CountAcceptedDevices() NamespaceQueryOption

	// EnrichMembersData join the user's data into members array.
	EnrichMembersData() NamespaceQueryOption
}
