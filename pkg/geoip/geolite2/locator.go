package geolite2

import (
	"context"
	"io"
	"net"
	"path/filepath"

	"github.com/oschwald/geoip2-golang"
	"github.com/shellhub-io/shellhub/pkg/geoip"
)

const (
	dbPath      = "/usr/share/GeoIP/" // Directory where the GeoIP database files are stored.
	dbCountryID = "GeoLite2-Country"  // GeoLite2 country database filename without extension.
	dbCityID    = "GeoLite2-City"     // GeoLite2 city database filename without extension.
	dbExtension = ".mmdb"             // Database file extension.
)

type geoLite2 struct {
	countryDB *geoip2.Reader
	cityDB    *geoip2.Reader
}

// Check if geoLite2 implements the geoip.Locator and io.Closer interfaces.
var (
	_ geoip.Locator = (*geoLite2)(nil)
	_ io.Closer     = (*geoLite2)(nil)
)

// NewLocator initializes a new geoip.Locator by setting up access to the GeoIP databases.
// If the databases do not exist locally, they will be downloaded using the provided fetcher method.
func NewLocator(ctx context.Context, fetcher GeoliteFetcher) (geoip.Locator, error) {
	if err := ensureDatabasePath(); err != nil {
		return nil, err
	}

	if err := fetcher(ctx); err != nil {
		return nil, err
	}

	countryDB, err := geoip2.Open(filepath.Join(dbPath, dbCountryID+dbExtension))
	if err != nil {
		return nil, err
	}

	cityDB, err := geoip2.Open(filepath.Join(dbPath, dbCityID+dbExtension))
	if err != nil {
		countryDB.Close()

		return nil, err
	}

	return &geoLite2{countryDB: countryDB, cityDB: cityDB}, nil
}

func (g *geoLite2) GetCountry(ip net.IP) (string, error) {
	record, err := g.countryDB.Country(ip)
	if err != nil {
		return "", err
	}

	return record.Country.IsoCode, nil
}

func (g *geoLite2) GetPosition(ip net.IP) (geoip.Position, error) {
	record, err := g.cityDB.City(ip)
	if err != nil {
		return geoip.Position{}, err
	}

	pos := geoip.Position{
		Longitude: record.Location.Longitude,
		Latitude:  record.Location.Latitude,
	}

	return pos, nil
}

func (g *geoLite2) Close() error {
	if err := g.countryDB.Close(); err != nil {
		return err
	}

	if err := g.cityDB.Close(); err != nil { //nolint:revive
		return err
	}

	return nil
}
