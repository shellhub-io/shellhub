// Package geoip helps in geolocation operations.
package geoip

import (
	"io"
	"net"
)

type nullGeoLite struct{}

var _ Locator = (*nullGeoLite)(nil)

var _ io.Closer = (*nullGeoLite)(nil)

// NewNullGeoLite returns a no-op [Locator] that resolves every IP to an empty country and the
// zero Position, never failing. The server falls back to it in every edition when no GeoIP
// database opens.
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
