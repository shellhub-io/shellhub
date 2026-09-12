package gateway

// Actor is the authenticated identity a request carries: who is performing it, established before
// any namespace is considered. An actor is not yet a member of anything — resolving it within a
// namespace scope is what produces the acting member.
//
// Which fields are set follows the credential the request authenticated with. A user token names
// the acting person, filling ID and Username; an API key and a device token name a namespace
// principal with no person behind it, so both leave ID and Username empty.
type Actor struct {
	// ID is the acting user's ID, empty when the credential names no person.
	ID string

	// Username is the acting user's username. It is the only identifier an admin-console request
	// carries, because that surface deliberately strips the user's ID.
	Username string

	// APIKey is the key the request authenticated with, empty otherwise.
	APIKey string

	// DeviceUID is the device the request authenticated as, empty otherwise.
	DeviceUID string
}

// IsZero reports whether the request carried no authenticated identity at all. A route that
// requires an actor refuses such a request; an anonymous route is the one place it is expected.
func (a Actor) IsZero() bool {
	return a.ID == "" && a.Username == "" && a.APIKey == "" && a.DeviceUID == ""
}
