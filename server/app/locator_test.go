package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/geoip/geoiptest"
	geoipmocks "github.com/shellhub-io/shellhub/pkg/geoip/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var bakedPlace = geoiptest.Place{Country: "CH", Latitude: 47.37, Longitude: 8.54}

func locatorTestConfig(t *testing.T, database string) (locatorConfig, *atomic.Int32) {
	t.Helper()

	baked := t.TempDir()
	geoiptest.WriteCityDatabase(t, filepath.Join(baked, "dbip-city-lite-2026-09.mmdb"), bakedPlace)

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	return locatorConfig{
		database:    database,
		downloadDir: t.TempDir(),
		bakedDir:    baked,
		update:      true,
		updateURL:   server.URL,
		clock:       clock.NewFixed(time.Date(2026, time.November, 5, 0, 0, 0, 0, time.UTC)),
	}, &requests
}

func maxmindFactory(locator geoip.Locator, err error) func(context.Context) (geoip.Locator, error) {
	return func(context.Context) (geoip.Locator, error) { return locator, err }
}

func assertLocatesBaked(t *testing.T, locator geoip.Locator) {
	t.Helper()

	country, err := locator.GetCountry(net.ParseIP("203.0.113.7"))
	require.NoError(t, err)
	assert.Equal(t, bakedPlace.Country, country)
}

func TestNewLocator(t *testing.T) {
	public := net.ParseIP("203.0.113.7")

	t.Run("uses DB-IP by default, reporting it", func(t *testing.T) {
		cfg, _ := locatorTestConfig(t, "dbip")
		cfg.factory = maxmindFactory(geoipmocks.NewMockLocator(t), nil)

		chosen, err := newLocator(context.Background(), cfg)
		require.NoError(t, err)

		assertLocatesBaked(t, chosen.locator)
		assert.Equal(t, geoipSourceDBIP, chosen.source)
	})

	t.Run("uses MaxMind when chosen and it loads, reporting it and refreshing nothing", func(t *testing.T) {
		cfg, _ := locatorTestConfig(t, "maxmind")
		maxmind := geoipmocks.NewMockLocator(t)
		maxmind.On("GetCountry", public).Return("US", nil).Once()
		cfg.factory = maxmindFactory(maxmind, nil)

		chosen, err := newLocator(context.Background(), cfg)
		require.NoError(t, err)

		country, err := chosen.locator.GetCountry(public)
		require.NoError(t, err)
		assert.Equal(t, "US", country)
		assert.Equal(t, geoipSourceMaxMind, chosen.source)
		assert.Nil(t, chosen.keepFresh)
	})

	t.Run("falls back to DB-IP when MaxMind is chosen and fails", func(t *testing.T) {
		cfg, _ := locatorTestConfig(t, "maxmind")
		cfg.factory = maxmindFactory(nil, errors.New("maxmind unreachable"))

		chosen, err := newLocator(context.Background(), cfg)
		require.NoError(t, err)

		assertLocatesBaked(t, chosen.locator)
		assert.Equal(t, geoipSourceDBIP, chosen.source)
	})

	t.Run("falls back to DB-IP when MaxMind is chosen and not configured", func(t *testing.T) {
		cfg, _ := locatorTestConfig(t, "maxmind")
		cfg.factory = maxmindFactory(nil, nil)

		chosen, err := newLocator(context.Background(), cfg)
		require.NoError(t, err)

		assertLocatesBaked(t, chosen.locator)
		assert.Equal(t, geoipSourceDBIP, chosen.source)
	})

	t.Run("falls back to DB-IP when MaxMind is chosen in an edition without it", func(t *testing.T) {
		cfg, _ := locatorTestConfig(t, "maxmind")

		chosen, err := newLocator(context.Background(), cfg)
		require.NoError(t, err)

		assertLocatesBaked(t, chosen.locator)
		assert.Equal(t, geoipSourceDBIP, chosen.source)
	})

	t.Run("refuses a database it does not know", func(t *testing.T) {
		cfg, _ := locatorTestConfig(t, "ip2location")

		_, err := newLocator(context.Background(), cfg)

		require.ErrorIs(t, err, errUnknownGeoIPDatabase)
	})

	t.Run("keeps DB-IP fresh from the update URL when the refresh is on", func(t *testing.T) {
		cfg, requests := locatorTestConfig(t, "dbip")

		chosen, err := newLocator(context.Background(), cfg)
		require.NoError(t, err)
		require.NotNil(t, chosen.keepFresh)

		ctx, cancel := context.WithCancel(context.Background())
		stopped := make(chan struct{})
		go func() {
			chosen.keepFresh(ctx)
			close(stopped)
		}()

		assert.EventuallyWithT(t, func(tt *assert.CollectT) {
			assert.Positive(tt, requests.Load())
		}, 5*time.Second, 10*time.Millisecond)
		cancel()
		<-stopped
	})

	t.Run("asks DB-IP for nothing when the refresh is off", func(t *testing.T) {
		cfg, requests := locatorTestConfig(t, "dbip")
		cfg.update = false

		chosen, err := newLocator(context.Background(), cfg)
		require.NoError(t, err)

		assert.NotNil(t, chosen.locator)
		assert.Nil(t, chosen.keepFresh)
		assert.Zero(t, requests.Load())
	})

	t.Run("places no device and reports no source when no database opens", func(t *testing.T) {
		cfg, _ := locatorTestConfig(t, "dbip")
		cfg.bakedDir = t.TempDir()

		chosen, err := newLocator(context.Background(), cfg)
		require.NoError(t, err)

		position, err := chosen.locator.GetPosition(public)
		require.NoError(t, err)
		assert.Equal(t, geoip.Position{}, position)
		assert.Empty(t, chosen.source)
		assert.Nil(t, chosen.keepFresh)
	})

	t.Run("places private addresses in cities with fake, over DB-IP", func(t *testing.T) {
		cfg, _ := locatorTestConfig(t, "fake")

		chosen, err := newLocator(context.Background(), cfg)
		require.NoError(t, err)

		position, err := chosen.locator.GetPosition(net.ParseIP("172.18.0.4"))
		require.NoError(t, err)
		assert.NotEqual(t, geoip.Position{}, position)
		assertLocatesBaked(t, chosen.locator)
		assert.Equal(t, geoipSourceDBIP, chosen.source)
	})
}
