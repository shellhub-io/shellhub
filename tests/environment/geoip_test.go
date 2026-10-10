package environment

import (
	"net/netip"
	"path/filepath"
	"testing"

	"github.com/oschwald/maxminddb-golang/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteGeoIPDatabases(t *testing.T) {
	florianopolis := Location{Country: "BR", Latitude: -27.5954, Longitude: -48.548}

	dir := t.TempDir()
	require.NoError(t, writeGeoIPDatabases(dir, florianopolis))

	addresses := []string{"127.0.0.1", "10.0.2.100", "172.18.0.1", "8.8.8.8", "::1", "2001:db8::1"}

	for _, kind := range []string{"GeoLite2-Country", "GeoLite2-City"} {
		t.Run(kind, func(t *testing.T) {
			db := openGeoIPDatabase(t, filepath.Join(dir, kind+".mmdb"))

			assert.Equal(t, kind, db.Metadata.DatabaseType)

			for _, address := range addresses {
				var record struct {
					Country struct {
						ISOCode string `maxminddb:"iso_code"`
					} `maxminddb:"country"`
				}

				require.NoError(t, db.Lookup(netip.MustParseAddr(address)).Decode(&record))
				assert.Equal(t, "BR", record.Country.ISOCode, address)
			}
		})
	}

	t.Run("the city database places every address at the location", func(t *testing.T) {
		type position struct {
			Latitude  float64 `maxminddb:"latitude"`
			Longitude float64 `maxminddb:"longitude"`
		}

		db := openGeoIPDatabase(t, filepath.Join(dir, "GeoLite2-City.mmdb"))

		for _, address := range addresses {
			var record struct {
				Location position `maxminddb:"location"`
			}

			require.NoError(t, db.Lookup(netip.MustParseAddr(address)).Decode(&record))
			assert.Equal(t, position{Latitude: florianopolis.Latitude, Longitude: florianopolis.Longitude}, record.Location, address)
		}
	})
}

func openGeoIPDatabase(t *testing.T, path string) *maxminddb.Reader {
	t.Helper()

	db, err := maxminddb.Open(path)
	require.NoError(t, err)

	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	return db
}
