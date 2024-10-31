package geoip

import "net"

type Locator interface {
	// GetCountry retrieves the ISO country code for a given IP address.
	// Returns an error if the IP lookup fails.
	GetCountry(ip net.IP) (string, error)

	// GetPosition retrieves the geographical [Position] for a given IP
	// address. Returns an error if the IP lookup fails.
	GetPosition(ip net.IP) (Position, error)
}
