package pg

import "github.com/shellhub-io/shellhub/api/store"

// TODO: maybe these methods can be deprecated with bun

func (pg *pg) Options() store.QueryOptions {
	return nil
}

func (pg *pg) CountAcceptedDevices() store.NamespaceQueryOption {
	return nil
}

func (pg *pg) EnrichMembersData() store.NamespaceQueryOption {
	return nil
}
