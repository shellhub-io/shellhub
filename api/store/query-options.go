package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type NamespaceQueryOption func(ctx context.Context, ns *models.Namespace) error

type QueryOptions interface {
	// CountAcceptedDevices counts the devices with a status 'accepted'
	// in the namespace.
	CountAcceptedDevices() NamespaceQueryOption

	// EnrichMembersData join the user's data into members array.
	EnrichMembersData() NamespaceQueryOption
}
