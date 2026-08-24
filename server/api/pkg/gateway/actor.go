package gateway

// Actor is the authenticated identity a request carries: who is performing it. It names no
// namespace, role or admin flag.
//
// Which fields are set follows the credential the request authenticated with. A user token names
// the acting person, filling ID and, unless the account signed in through SAML, Username. On the
// admin console the ID arrives as X-Admin-ID. An API key and a device token name a namespace
// principal with no person behind it, so both leave ID and Username empty. APIKey is set only for a
// request that authenticated with an API key, and DeviceUID only for one that authenticated as a
// device.
type Actor struct {
	ID        string
	Username  string
	APIKey    string
	DeviceUID string
}

// IsZero reports whether the request carried no authenticated identity at all: ID, Username,
// APIKey and DeviceUID are all empty.
func (a Actor) IsZero() bool {
	return a.ID == "" && a.Username == "" && a.APIKey == "" && a.DeviceUID == ""
}
