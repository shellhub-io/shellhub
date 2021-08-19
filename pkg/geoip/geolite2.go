package geoip

import (
	"io"
	"net"

	"github.com/oschwald/geoip2-golang"
)

const dbFile = "/usr/share/GeoIP/GeoLite2-Country.mmdb"

// Check if geoLite2 implements Locator interface.
var _ Locator = (*geoLite2)(nil)

// Check if geoLite2 implements io.Closer interface.
var _ io.Closer = (*geoLite2)(nil)

// geoLite2 is a structure what stores a geoIp2Reader to a GeoIp2 database.
type geoLite2 struct {
	db *geoip2.Reader
}

// NewGeoLite2 opens a connection to GeoIp2 database and return a geoLite2 structure with the database connection.
func NewGeoLite2() (Locator, error) {
	db, err := geoip2.Open(dbFile)
	if err != nil {
		return nil, err
	}

	return &geoLite2{
		db: db,
	}, nil
}

// Close the connection with the GeoLite2 database, returning either error or nil.
func (g *geoLite2) Close() error {
	return g.db.Close()
}

// GetCountry gets an ip and return either an ISO 3166-1 code to a country or an empty string.
func (g *geoLite2) GetCountry(ip net.IP) (string, error) {
	record, err := g.db.Country(ip)
	if err != nil {
		return "", err
	}

	return record.Country.IsoCode, nil
}

// GetPosition gets an ip and return a Position structure with Longitude and Latitude with error nil or an empty Position structure with the error.
func (g *geoLite2) GetPosition(ip net.IP) (Position, error) {
	record, err := g.db.City(ip)
	if err != nil {
		return Position{}, err
	}

	return Position{
		Longitude: record.Location.Longitude,
		Latitude:  record.Location.Latitude,
	}, nil
}
