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
	dir := t.TempDir()
	require.NoError(t, writeGeoIPDatabases(dir, "BR"))

	for _, kind := range []string{"GeoLite2-Country", "GeoLite2-City"} {
		t.Run(kind, func(t *testing.T) {
			db, err := maxminddb.Open(filepath.Join(dir, kind+".mmdb"))
			require.NoError(t, err)

			t.Cleanup(func() { assert.NoError(t, db.Close()) })

			assert.Equal(t, kind, db.Metadata.DatabaseType)

			for _, address := range []string{"127.0.0.1", "10.0.2.100", "172.18.0.1", "8.8.8.8", "::1", "2001:db8::1"} {
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
}
