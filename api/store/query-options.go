package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type QueryOption func(ctx context.Context) error

type NamespaceQueryOption func(ctx context.Context, ns *models.Namespace) error

type QueryOptions interface {
	// CountAcceptedDevices counts the devices with a status 'accepted'
	// in the namespace.
	CountAcceptedDevices() NamespaceQueryOption

	// EnrichMembersData join the user's data into members array.
	EnrichMembersData() NamespaceQueryOption

	Paginate(offset, limit int) QueryOption
	Order(column, direction string) QueryOption
	WithMember(userID string) QueryOption
}
