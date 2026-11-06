package geoip_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/geoip/geoiptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	zurich = geoiptest.Place{Country: "CH", Latitude: 47.37, Longitude: 8.54}
	berlin = geoiptest.Place{Country: "DE", Latitude: 52.52, Longitude: 13.40}
	lisbon = geoiptest.Place{Country: "PT", Latitude: 38.72, Longitude: -9.14}
)

func gzipped(t *testing.T, data []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write(data)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	return buf.Bytes()
}

func assertLocates(t *testing.T, locator geoip.Locator, at geoiptest.Place) {
	t.Helper()

	country, err := locator.GetCountry(net.ParseIP("203.0.113.7"))
	require.NoError(t, err)
	assert.Equal(t, at.Country, country)

	position, err := locator.GetPosition(net.ParseIP("2001:db8::1"))
	require.NoError(t, err)
	assert.InDelta(t, at.Latitude, position.Latitude, 1e-9)
	assert.InDelta(t, at.Longitude, position.Longitude, 1e-9)
}

func bakedAndDownloads(t *testing.T) (baked, downloads string) {
	t.Helper()

	baked = t.TempDir()
	geoiptest.WriteCityDatabase(t, filepath.Join(baked, "dbip-city-lite-2026-09.mmdb"), zurich)

	return baked, t.TempDir()
}

func november(day int) time.Time {
	return time.Date(2026, time.November, day, 3, 0, 0, 0, time.UTC)
}

func TestOpenDBIP(t *testing.T) {
	t.Run("reads the baked database when nothing was downloaded", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)

		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })

		assertLocates(t, locator, zurich)
	})

	t.Run("reads the newest baked database when the image carries more than one", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)
		geoiptest.WriteCityDatabase(t, filepath.Join(baked, "dbip-city-lite-2026-10.mmdb"), berlin)

		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })

		assertLocates(t, locator, berlin)
	})

	t.Run("reads the newest download over the baked database", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)
		geoiptest.WriteCityDatabase(t, filepath.Join(downloads, "dbip-city-lite-2026-10.mmdb"), berlin)
		geoiptest.WriteCityDatabase(t, filepath.Join(downloads, "dbip-city-lite-2026-11.mmdb"), lisbon)

		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })

		assertLocates(t, locator, lisbon)
	})

	t.Run("reads a baked database newer than every download", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)
		geoiptest.WriteCityDatabase(t, filepath.Join(downloads, "dbip-city-lite-2026-08.mmdb"), berlin)

		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })

		assertLocates(t, locator, zurich)
	})

	t.Run("sweeps a download an earlier process left half written", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)
		leftover := filepath.Join(downloads, ".download-123")
		require.NoError(t, os.WriteFile(leftover, []byte("partial"), 0o600))

		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })

		assert.NoFileExists(t, leftover)
	})

	t.Run("fails when there is no database at all", func(t *testing.T) {
		_, err := geoip.OpenDBIP(t.TempDir(), t.TempDir())

		require.ErrorIs(t, err, geoip.ErrNoDatabase)
	})
}

func dbipServer(t *testing.T, months map[string][]byte) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		body, ok := months[r.URL.Path]
		if !ok {
			http.NotFound(w, r)

			return
		}

		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	return server, &requests
}

func TestDBIPRefresh(t *testing.T) {
	t.Run("swaps in the month's database and deletes the previous download", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)
		previous := filepath.Join(downloads, "dbip-city-lite-2026-10.mmdb")
		geoiptest.WriteCityDatabase(t, previous, berlin)
		server, _ := dbipServer(t, map[string][]byte{
			"/dbip-city-lite-2026-11.mmdb.gz": gzipped(t, geoiptest.CityDatabase(t, lisbon)),
		})
		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })

		require.NoError(t, locator.Refresh(context.Background(), server.URL, november(5)))

		assertLocates(t, locator, lisbon)
		assert.NoFileExists(t, previous)
		assert.FileExists(t, filepath.Join(downloads, "dbip-city-lite-2026-11.mmdb"))
		assert.FileExists(t, filepath.Join(baked, "dbip-city-lite-2026-09.mmdb"))
	})

	t.Run("asks nothing once it holds the month's database", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)
		geoiptest.WriteCityDatabase(t, filepath.Join(downloads, "dbip-city-lite-2026-11.mmdb"), lisbon)
		server, requests := dbipServer(t, nil)
		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })

		require.NoError(t, locator.Refresh(context.Background(), server.URL, november(20)))

		assert.Zero(t, requests.Load())
	})

	t.Run("keeps its database while the month's is not published", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)
		server, _ := dbipServer(t, nil)
		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })

		require.NoError(t, locator.Refresh(context.Background(), server.URL, november(1)))

		assertLocates(t, locator, zurich)
	})

	t.Run("keeps its database when the download fails", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)
		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })

		err = locator.Refresh(context.Background(), server.URL, november(5))

		require.ErrorIs(t, err, geoip.ErrDownload)
		assertLocates(t, locator, zurich)
		entries, err := os.ReadDir(downloads)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("keeps its database when the download is not a database", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)
		server, _ := dbipServer(t, map[string][]byte{
			"/dbip-city-lite-2026-11.mmdb.gz": gzipped(t, []byte("not a database")),
		})
		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })

		err = locator.Refresh(context.Background(), server.URL, november(5))

		require.ErrorIs(t, err, geoip.ErrDownload)
		assertLocates(t, locator, zurich)
		entries, err := os.ReadDir(downloads)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
}

func TestDBIPKeepFresh(t *testing.T) {
	t.Run("refreshes as soon as it starts and stops with its context", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)
		server, requests := dbipServer(t, map[string][]byte{
			"/dbip-city-lite-2026-11.mmdb.gz": gzipped(t, geoiptest.CityDatabase(t, lisbon)),
		})
		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })
		ctx, cancel := context.WithCancel(context.Background())
		stopped := make(chan struct{})

		go func() {
			locator.KeepFresh(ctx, server.URL, clock.NewFixed(november(5)), time.Hour, time.Minute, func(error) {})
			close(stopped)
		}()

		assert.EventuallyWithT(t, func(tt *assert.CollectT) {
			position, err := locator.GetPosition(net.ParseIP("203.0.113.7"))
			require.NoError(tt, err)
			assert.InDelta(tt, lisbon.Latitude, position.Latitude, 1e-9)
		}, 5*time.Second, 10*time.Millisecond)
		assert.Equal(t, int32(1), requests.Load())

		cancel()
		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			t.Fatal("KeepFresh did not stop with its context")
		}
	})

	t.Run("reports a failed refresh and keeps going", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)
		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		failures := make(chan error, 1)

		go locator.KeepFresh(ctx, server.URL, clock.NewFixed(november(5)), time.Hour, time.Minute, func(err error) {
			select {
			case failures <- err:
			default:
			}
		})

		select {
		case err := <-failures:
			require.ErrorIs(t, err, geoip.ErrDownload)
		case <-time.After(5 * time.Second):
			t.Fatal("KeepFresh reported no failure")
		}
		assertLocates(t, locator, zurich)
	})

	t.Run("gives up on a download that stalls and keeps going", func(t *testing.T) {
		baked, downloads := bakedAndDownloads(t)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			<-r.Context().Done()
		}))
		t.Cleanup(server.Close)
		locator, err := geoip.OpenDBIP(downloads, baked)
		require.NoError(t, err)
		t.Cleanup(func() { _ = locator.Close() })
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		failures := make(chan error, 1)

		go locator.KeepFresh(ctx, server.URL, clock.NewFixed(november(5)), time.Hour, 50*time.Millisecond, func(err error) {
			select {
			case failures <- err:
			default:
			}
		})

		select {
		case err := <-failures:
			require.ErrorIs(t, err, geoip.ErrDownload)
		case <-time.After(5 * time.Second):
			t.Fatal("a stalled download was never given up")
		}
		assertLocates(t, locator, zurich)
	})
}
