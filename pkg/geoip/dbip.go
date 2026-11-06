package geoip

import (
	"compress/gzip"
	"context"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang/v2"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/errors"
)

// ErrLayer names the geoip layer in the errors it returns.
const ErrLayer = "geoip"

const (
	// ErrCodeNoDatabase is the code of ErrNoDatabase.
	ErrCodeNoDatabase = iota + 1
	// ErrCodeDownload is the code of ErrDownload.
	ErrCodeDownload
	// ErrCodeInvalidAddress is the code of ErrInvalidAddress.
	ErrCodeInvalidAddress
)

var (
	// ErrNoDatabase is returned by [OpenDBIP] when no database opens. It is joined with the error
	// opening the last one tried, or with [os.ErrNotExist] when there was none to try.
	ErrNoDatabase = errors.New("no GeoIP database to open", ErrLayer, ErrCodeNoDatabase)
	// ErrDownload is returned by [DBIP.Refresh] when the month's database cannot be fetched, written
	// or opened. It is joined with the cause.
	ErrDownload = errors.New("cannot download the GeoIP database", ErrLayer, ErrCodeDownload)
	// ErrInvalidAddress is returned by a lookup given an IP that is neither IPv4 nor IPv6.
	ErrInvalidAddress = errors.New("invalid IP address", ErrLayer, ErrCodeInvalidAddress)
)

// DBIPDownloadURL is where DB-IP publishes its free City Lite databases, one per month.
const DBIPDownloadURL = "https://download.db-ip.com/free"

const (
	maxDatabaseSize = 1 << 30

	partialPrefix = ".download-"
	dbipPrefix    = "dbip-city-lite-"
	dbipExtension = ".mmdb"
	monthLayout   = "2006-01"
)

func dbipName(month string) string {
	return dbipPrefix + month + dbipExtension
}

func dbipMonth(path string) string {
	name := filepath.Base(path)
	if !strings.HasPrefix(name, dbipPrefix) || !strings.HasSuffix(name, dbipExtension) {
		return ""
	}

	return strings.TrimSuffix(strings.TrimPrefix(name, dbipPrefix), dbipExtension)
}

// DBIP locates addresses with a DB-IP City Lite database, which answers both the country and the
// position. It reads the newest database it finds, downloaded or baked into the image, and
// [DBIP.Refresh] swaps in a newer month without a restart. It is safe for concurrent use.
type DBIP struct {
	dir string

	mu     sync.RWMutex
	reader *geoip2.Reader
	month  string
}

var (
	_ Locator   = (*DBIP)(nil)
	_ io.Closer = (*DBIP)(nil)
)

func databases(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && dbipMonth(entry.Name()) != "" {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}

	slices.Sort(paths)

	return paths
}

func (d *DBIP) downloads() []string {
	return databases(d.dir)
}

func sweepPartials(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), partialPrefix) {
			_ = os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
}

// OpenDBIP opens the newest database by month among those downloaded into dir and those baked into
// the image in baked, trying an older one when the newest does not open. Databases are named
// dbip-city-lite-YYYY-MM.mmdb, which is how the month is known. It first deletes any download an
// earlier process left half written in dir. It returns ErrNoDatabase when nothing opens.
func OpenDBIP(dir, baked string) (*DBIP, error) {
	d := &DBIP{dir: dir}

	sweepPartials(dir)

	candidates := append(d.downloads(), databases(baked)...)
	slices.SortStableFunc(candidates, func(a, b string) int {
		return strings.Compare(dbipMonth(b), dbipMonth(a))
	})

	last := os.ErrNotExist
	for _, path := range candidates {
		reader, err := geoip2.Open(path)
		if err != nil {
			last = err

			continue
		}

		d.reader = reader
		d.month = dbipMonth(path)

		return d, nil
	}

	return nil, errors.Wrap(ErrNoDatabase, last)
}

func toAddr(ip net.IP) (netip.Addr, error) {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return netip.Addr{}, ErrInvalidAddress
	}

	return addr.Unmap(), nil
}

func (d *DBIP) city(ip net.IP) (*geoip2.City, error) {
	addr, err := toAddr(ip)
	if err != nil {
		return nil, err
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.reader.City(addr)
}

// GetCountry returns the ISO 3166-1 code of ip's country, empty when the database has none. It
// returns ErrInvalidAddress for an ip that is neither IPv4 nor IPv6, and the database reader's
// error when the lookup fails, as it does after Close.
func (d *DBIP) GetCountry(ip net.IP) (string, error) {
	record, err := d.city(ip)
	if err != nil {
		return "", err
	}

	return record.Country.ISOCode, nil
}

// GetPosition returns ip's position, the zero Position when the database has none. It returns
// ErrInvalidAddress for an ip that is neither IPv4 nor IPv6, and the database reader's error when
// the lookup fails, as it does after Close.
func (d *DBIP) GetPosition(ip net.IP) (Position, error) {
	record, err := d.city(ip)
	if err != nil {
		return Position{}, err
	}

	position := Position{}
	if record.Location.Latitude != nil {
		position.Latitude = *record.Location.Latitude
	}

	if record.Location.Longitude != nil {
		position.Longitude = *record.Location.Longitude
	}

	return position, nil
}

// Close closes the database in use and returns the reader's error, if any. Lookups after Close
// fail, and a Refresh after it swaps in a database nothing closes.
func (d *DBIP) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.reader.Close()
}

// Refresh fetches now's month from baseURL, published there as dbip-city-lite-YYYY-MM.mmdb.gz,
// when the database in use is older. It writes it into the directory, swaps it in and deletes the
// downloads it replaced, never the baked database. It requests nothing when the database in use is
// already that month's, and returns nil, keeping its database, while the month is not published.
// It returns ErrDownload, keeping its database and leaving no partial file behind, when the fetch,
// the write or the opening fails. It fetches under ctx.
func (d *DBIP) Refresh(ctx context.Context, baseURL string, now time.Time) error {
	month := now.UTC().Format(monthLayout)

	d.mu.RLock()
	current := d.month
	d.mu.RUnlock()

	if current >= month {
		return nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/"+dbipName(month)+".gz", nil)
	if err != nil {
		return errors.Wrap(ErrDownload, err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return errors.Wrap(ErrDownload, err)
	}
	defer response.Body.Close() //nolint:errcheck // a close error changes nothing Refresh returns

	if response.StatusCode == http.StatusNotFound {
		return nil
	}

	if response.StatusCode != http.StatusOK {
		return errors.Wrap(ErrDownload, errors.New(response.Status, ErrLayer, ErrCodeDownload))
	}

	path := filepath.Join(d.dir, dbipName(month))
	reader, err := d.write(response.Body, path)
	if err != nil {
		return errors.Wrap(ErrDownload, err)
	}

	d.mu.Lock()
	previous := d.reader
	d.reader = reader
	d.month = month
	d.mu.Unlock()

	_ = previous.Close()

	for _, download := range d.downloads() {
		if download != path {
			_ = os.Remove(download)
		}
	}

	return nil
}

// KeepFresh calls Refresh with c's time at once and then every every until ctx ends, giving each
// attempt at most timeout. It hands each failure, a stalled download included, to failed and goes
// on with the database it has. It blocks until ctx ends.
func (d *DBIP) KeepFresh(ctx context.Context, baseURL string, c clock.Clock, every, timeout time.Duration, failed func(error)) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		attempt, cancel := context.WithTimeout(ctx, timeout)
		err := d.Refresh(attempt, baseURL, c.Now())
		cancel()

		if err != nil && ctx.Err() == nil {
			failed(err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (d *DBIP) write(body io.Reader, path string) (*geoip2.Reader, error) {
	if err := os.MkdirAll(d.dir, 0o750); err != nil {
		return nil, err
	}

	unzipped, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}

	partial, err := os.CreateTemp(d.dir, partialPrefix+"*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(partial.Name()) //nolint:errcheck // the file was renamed into place, or is a leftover to drop

	if _, err := io.Copy(partial, io.LimitReader(unzipped, maxDatabaseSize)); err != nil {
		_ = partial.Close()

		return nil, err
	}

	if err := partial.Close(); err != nil {
		return nil, err
	}

	reader, err := geoip2.Open(partial.Name())
	if err != nil {
		return nil, err
	}

	if err := os.Rename(partial.Name(), path); err != nil {
		_ = reader.Close()

		return nil, err
	}

	return reader, nil
}
