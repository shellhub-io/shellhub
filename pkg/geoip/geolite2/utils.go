package geolite2

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/mholt/archiver/v4"
	"golang.org/x/sync/errgroup"
)

// ensureDatabasePath ensures that [dbPath] exists and creates if not.
func ensureDatabasePath() error {
	if err := os.MkdirAll(dbPath, 0o755); err != nil {
		return errors.New("failed to create dbPath: " + err.Error())
	}

	return nil
}

// fetchDBs concurrently downloads the GeoIP database files from the provided URLs and extracts
// them to [dbPath]. It will halt and return an error if any download or extraction fails.
func fetchDBs(ctx context.Context, urls []string) error {
	g, ctx := errgroup.WithContext(ctx)

	for _, url := range urls {
		g.Go(fetchDB(ctx, url))
	}

	return g.Wait()
}

// fetchDB downalods the Geolite database file from the provided URL and extracts them to [dbPath].
func fetchDB(ctx context.Context, url string) func() error {
	return func() error {
		r, err := resty.New().R().Get(url)
		if err != nil {
			return err
		}

		if r.StatusCode() != http.StatusOK {
			return errors.New("cannot download geolite db: status " + r.Status())
		}

		format := archiver.CompressedArchive{Compression: archiver.Gz{}, Archival: archiver.Tar{}}
		if err := format.Extract(ctx, bytes.NewReader(r.Body()), nil, saveDB()); err != nil { //nolint:revive
			return err
		}

		return nil
	}
}

// saveDB saves extracted GeoLite2 database files to [dbPath].
// Only files with the expected [dbExtension] will be saved.
func saveDB() archiver.FileHandler {
	return func(_ context.Context, f archiver.File) error {
		if !strings.HasSuffix(f.Name(), dbExtension) {
			return nil
		}

		srcFile, err := f.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destPath := filepath.Join(dbPath, f.Name())

		outFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, srcFile)

		return err
	}
}
