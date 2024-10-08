package store

type StoreOptions interface {
	CountAcceptedDevices() NamespaceQueryOption
	EnrichMembersData() NamespaceQueryOption
}

// CountAcceptedDevices counts the devices with a status 'accepted'
// in the namespace.
func CountAcceptedDevices(s Store) NamespaceQueryOption {
	return s.Options().CountAcceptedDevices()
}

// EnrichMembersData join the user's data into members array.
func EnrichMembersData(s Store) NamespaceQueryOption {
	return s.Options().EnrichMembersData()
}
