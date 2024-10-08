package options

import (
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
)

// CountAcceptedDevices counts the devices with a status 'accepted'
// in the namespace.
func CountAcceptedDevices() store.NamespaceQueryOption {
	return mongo.CountAcceptedDevices()
}

// EnrichMembersData join the user's data into members array.
func EnrichMembersData() store.NamespaceQueryOption {
	return mongo.EnrichMembersData()
}
