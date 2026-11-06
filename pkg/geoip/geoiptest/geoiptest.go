// Package geoiptest writes small GeoIP databases for tests, so no test needs a real one.
package geoiptest

import (
	"bytes"
	"net"
	"os"
	"testing"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/stretchr/testify/require"
)

// Place is where a test database puts every address: Country, an ISO 3166-1 code such as "CH",
// and a position in degrees.
type Place struct {
	Country   string
	Latitude  float64
	Longitude float64
}

// CityDatabase returns a City database, in the DB-IP City Lite layout, that places every IPv4 and
// IPv6 address at at. It fails tb when the database cannot be built.
func CityDatabase(tb testing.TB, at Place) []byte {
	tb.Helper()

	tree, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType:            "DBIP-City-Lite",
		IncludeReservedNetworks: true,
	})
	require.NoError(tb, err)

	record := mmdbtype.Map{
		"country": mmdbtype.Map{"iso_code": mmdbtype.String(at.Country)},
		"location": mmdbtype.Map{
			"latitude":  mmdbtype.Float64(at.Latitude),
			"longitude": mmdbtype.Float64(at.Longitude),
		},
	}
	for _, cidr := range []string{"0.0.0.0/0", "::/0"} {
		_, network, err := net.ParseCIDR(cidr)
		require.NoError(tb, err)
		require.NoError(tb, tree.Insert(network, record))
	}

	var buf bytes.Buffer
	_, err = tree.WriteTo(&buf)
	require.NoError(tb, err)

	return buf.Bytes()
}

// WriteCityDatabase writes [CityDatabase] for at to path. It fails tb when the file cannot be
// written.
func WriteCityDatabase(tb testing.TB, path string, at Place) {
	tb.Helper()

	require.NoError(tb, os.WriteFile(path, CityDatabase(tb, at), 0o600))
}
