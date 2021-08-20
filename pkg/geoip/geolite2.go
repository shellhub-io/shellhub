// Package geoip helps in geolocation operations.
package geoip

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v3"

	"github.com/oschwald/geoip2-golang"
)

// dbPath is the default path for Database.
var dbPath = "/usr/share/GeoIP/"

// geoLite2DbName is the default database name of GeoLite2 when extracted.
var geoLite2DbName = "GeoLite2-City.mmdb"

// Check if geoLite2 implements Locator interface.
var _ Locator = (*geoLite2)(nil)

// Check if geoLite2 implements io.Closer interface.
var _ io.Closer = (*geoLite2)(nil)

// geoLite2 is a structure what stores a geoIp2Reader to a GeoIp2 database.
type geoLite2 struct {
	db *geoip2.Reader
}

// downloadGeoLite2Db downloads the GeoLite2 database and extract the files into the dbPath.
func downloadGeoLite2Db(maxmindDBLicense string) error {
	// Download the GeoLite2Db .tar.gz file with the database inside it.
	r, err := http.Get(fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=%s&suffix=tar.gz", maxmindDBLicense))
	if err != nil {
		return err
	}

	// Create a temporary directory to untar downloaded .tar.gz with database.
	tempDir, err := ioutil.TempDir("", "geoip")
	// Delete temporary directory.
	defer func(tempDir string) {
		err := os.RemoveAll(tempDir)
		if err != nil {
			return
		}
	}(tempDir)
	if err != nil {
		return err
	}

	// Create a temporary file to store downloaded .tar.gz with database.
	tempFile, err := ioutil.TempFile("", "geoip*.tar.gz")
	// Delete temporary file.
	defer func(tempFile *os.File) {
		err := os.Remove(tempFile.Name())
		if err != nil {
			return
		}
	}(tempFile)
	if err != nil {
		return err
	}

	// Copy bytes from downloaded file to temporary file.
	_, err = io.Copy(tempFile, r.Body)
	if err != nil {
		return err
	}

	// Untar the downloaded file to the temporary directory.
	err = archiver.Unarchive(tempFile.Name(), tempDir)
	if err != nil {
		return err
	}

	// Create the path to move decompressed database file.
	err = os.MkdirAll(dbPath, 0o755)
	if err != nil {
		return err
	}
	// Find geoip.geoLite2DbName inside the tempDir.
	err = filepath.Walk(tempDir, func(p string, i fs.FileInfo, err error) error {
		if i.Name() == geoLite2DbName {
			// Move from temporary directory to geoip.geoLite2DbName to geoip.dbPath.
			err := os.Rename(p, dbPath+geoLite2DbName)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// NewGeoLite2 opens a connection to GeoIp2 database and return a geoLite2 structure with the database connection.
//
// The connection uses the local database or try to download it from MaxMind's server (the download required `MAXMIND_LICENSE`).
func NewGeoLite2() (Locator, error) {
	if _, err := os.Stat(dbPath + geoLite2DbName); os.IsNotExist(err) {
		if os.Getenv("MAXMIND_LICENSE") != "" {
			err := downloadGeoLite2Db(os.Getenv("MAXMIND_LICENSE"))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	db, err := geoip2.Open(dbPath + geoLite2DbName)
	if err != nil {
		return nil, err
	}

	return &geoLite2{
		db: db,
	}, nil
}

// Close the connection with the GeoLite2 database, returning either error or nil.
func (g *geoLite2) Close() error {
	return g.db.Close()
}

// GetCountry gets an ip and return either an ISO 3166-1 code to a country or an empty string.
func (g *geoLite2) GetCountry(ip net.IP) (string, error) {
	record, err := g.db.Country(ip)
	if err != nil {
		return "", err
	}

	return record.Country.IsoCode, nil
}

// GetPosition gets an ip and return a Position structure with Longitude and Latitude with error nil or an empty Position structure with the error.
func (g *geoLite2) GetPosition(ip net.IP) (Position, error) {
	record, err := g.db.City(ip)
	if err != nil {
		return Position{}, err
	}

	return Position{
		Longitude: record.Location.Longitude,
		Latitude:  record.Location.Latitude,
	}, nil
}
