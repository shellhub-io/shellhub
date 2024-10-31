package geolite2

import (
	"context"
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
)

// GeoliteFetcher defines a function type for asynchronously fetching and downloading GeoIP databases.
type GeoliteFetcher func(context.Context) error

// FetchFromLicenseKey initializes a GeoipFetcher that downloads the GeoLite2 database files
// from MaxMind, using the provided license key to authenticate.
func FetchFromLicenseKey(licenseKey string) GeoliteFetcher {
	return func(ctx context.Context) error {
		urls := []string{}
		for _, id := range []string{dbCountryID, dbCityID} {
			_, err := os.Stat(filepath.Join(dbPath, id+dbExtension))
			switch {
			case errors.Is(err, fs.ErrNotExist):
				query := url.Values{}
				query.Add("suffix", "tar.gz")
				query.Add("license_key", licenseKey)
				query.Add("editon_id", id)

				urls = append(urls, "https://download.maxmind.com/app/geoip_download?"+query.Encode())
			default:
				return err
			}
		}

		if len(urls) > 0 {
			if err := fetchDBs(ctx, urls); err != nil {
				return err
			}
		}

		return nil
	}
}

func FetchFromMirror(mirror string) GeoliteFetcher {
	return func(ctx context.Context) error {
		urls := []string{}
		for _, id := range []string{dbCountryID, dbCityID} {
			_, err := os.Stat(filepath.Join(dbPath, id+dbExtension))
			switch {
			case errors.Is(err, fs.ErrNotExist):
				u, err := url.Parse(mirror)
				if err != nil {
					return err
				}

				u.Path = "/" + id + ".tar.gz"

				urls = append(urls, u.String())
			default:
				return err
			}
		}

		if len(urls) > 0 {
			if err := fetchDBs(ctx, urls); err != nil {
				return err
			}
		}

		return nil
	}
}
