// Package geoip helps in geolocation operations.
package geoip

import (
	"io"
	"net"
)

// nullGeoLite is a structure what stores a geoIp2Reader to a GeoIp2 database.
type nullGeoLite struct{}

// Check if geoLite2 implements Locator interface.
var _ Locator = (*nullGeoLite)(nil)

// Check if geoLite2 implements io.Closer interface.
var _ io.Closer = (*nullGeoLite)(nil)

// NewNullGeoLite returns a no-op [Locator] that resolves every IP to an empty
// result. It is the Community Edition fallback; the real MaxMind-backed locator
// lives in the cloud/enterprise build.
func NewNullGeoLite() Locator {
	return &nullGeoLite{}
}

// Close the connection with the GeoLite2 database, returning either error or nil.
func (g *nullGeoLite) Close() error {
	return nil
}

// GetCountry gets an ip and return either an ISO 3166-1 code to a country or an empty string.
func (g *nullGeoLite) GetCountry(_ net.IP) (string, error) {
	return "", nil
}

// GetPosition gets an ip and return a Position structure with Longitude and Latitude with error nil or an empty Position structure with the error.
func (g *nullGeoLite) GetPosition(_ net.IP) (Position, error) {
	return Position{}, nil
}
