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

// NewNullGeoLite opens a connection to GeoIp2 database and return a geoLite2 structure with the database connection.
//
// The connection uses the local database or try to download it from MaxMind's server (the download required `MAXMIND_LICENSE`).
func NewNullGeoLite() Locator {
	return &nullGeoLite{}
}

// Close the connection with the GeoLite2 database, returning either error or nil.
func (g *nullGeoLite) Close() error {
	return nil
}

// GetCountry gets an ip and return either an ISO 3166-1 code to a country or an empty string.
func (g *nullGeoLite) GetCountry(ip net.IP) (string, error) {
	return "", nil
}

// GetPosition gets an ip and return a Position structure with Longitude and Latitude with error nil or an empty Position structure with the error.
func (g *nullGeoLite) GetPosition(ip net.IP) (Position, error) {
	return Position{}, nil
}
