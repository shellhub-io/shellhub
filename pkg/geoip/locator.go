package geoip

import "net"

// Locator resolves an IP address to where it is. Implementations answer from a local database, so
// a lookup is offline and the answer is a guess with no confidence attached.
type Locator interface {
	// GetCountry retrieves the ISO country code for a given IP address.
	// Returns an error if the IP lookup fails.
	GetCountry(ip net.IP) (string, error)

	// GetPosition retrieves the geographical [Position] for a given IP
	// address. Returns an error if the IP lookup fails.
	GetPosition(ip net.IP) (Position, error)
}
