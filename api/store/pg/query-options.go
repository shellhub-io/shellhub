package pg

import "github.com/shellhub-io/shellhub/api/store"

// TODO: maybe these methods can be deprecated with the new gorm

func (s *Store) Options() store.QueryOptions {
	return nil
}

func (s *Store) CountAcceptedDevices() store.NamespaceQueryOption {
	return nil
}

func (s *Store) EnrichMembersData() store.NamespaceQueryOption {
	return nil
}
