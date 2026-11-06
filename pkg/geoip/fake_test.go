package geoip_test

import (
	"net"
	"path/filepath"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/geoip/geoiptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func realLocator(t *testing.T, at geoiptest.Place) geoip.Locator {
	t.Helper()

	baked := t.TempDir()
	geoiptest.WriteCityDatabase(t, filepath.Join(baked, "dbip-city-lite-2026-09.mmdb"), at)

	locator, err := geoip.OpenDBIP(t.TempDir(), baked)
	require.NoError(t, err)
	t.Cleanup(func() { _ = locator.Close() })

	return locator
}

func TestFake(t *testing.T) {
	private := []string{"172.18.0.4", "172.18.0.9", "192.168.1.20", "10.0.3.7", "127.0.0.1", "fd00::5", "::1"}

	t.Run("places a private address in a city", func(t *testing.T) {
		fake := geoip.NewFake(realLocator(t, zurich))

		for _, ip := range private {
			position, err := fake.GetPosition(net.ParseIP(ip))
			require.NoError(t, err)
			assert.NotEqual(t, geoip.Position{}, position, ip)
			assert.NotEqual(t, geoip.Position{Latitude: zurich.Latitude, Longitude: zurich.Longitude}, position, ip)

			country, err := fake.GetCountry(net.ParseIP(ip))
			require.NoError(t, err)
			assert.Len(t, country, 2, ip)
		}
	})

	t.Run("places an address in the same city every time", func(t *testing.T) {
		fake := geoip.NewFake(realLocator(t, zurich))

		first, err := fake.GetPosition(net.ParseIP("172.18.0.4"))
		require.NoError(t, err)
		again, err := fake.GetPosition(net.ParseIP("172.18.0.4"))
		require.NoError(t, err)

		assert.Equal(t, first, again)
	})

	t.Run("spreads private addresses over several cities", func(t *testing.T) {
		fake := geoip.NewFake(realLocator(t, zurich))

		positions := map[geoip.Position]bool{}
		for _, ip := range private {
			position, err := fake.GetPosition(net.ParseIP(ip))
			require.NoError(t, err)
			positions[position] = true
		}

		assert.Greater(t, len(positions), 2)
	})

	t.Run("looks a public address up in the real database", func(t *testing.T) {
		fake := geoip.NewFake(realLocator(t, zurich))

		assertLocates(t, fake, zurich)
	})
}
