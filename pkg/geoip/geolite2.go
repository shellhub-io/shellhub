// Package geoip helps in geolocation operations.
package geoip

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/mholt/archiver/v3"

	"github.com/oschwald/geoip2-golang"
)

// dbPath is the default path for Database.
var dbPath = "/usr/share/GeoIP/"

const (
	// city is used to access DB's connection to GeoLite2-City.
	city = iota
	// country is used to access DB's connection to GeoLite2-Country.
	country
)

// geoLite2Info contains data about which geoLite2's databases are used.
var geoLite2Info = []map[string]string{
	{"type": "City", "file": "GeoLite2-City.mmdb"},
	{"type": "Country", "file": "GeoLite2-Country.mmdb"},
}

// Check if geoLite2 implements Locator interface.
var _ Locator = (*geoLite2)(nil)

// Check if geoLite2 implements io.Closer interface.
var _ io.Closer = (*geoLite2)(nil)

// geoLite2 is a structure what stores a geoIp2Reader to a GeoIp2 database.
type geoLite2 struct {
	db []*geoip2.Reader
}

// downloadGeoLite2Db downloads the GeoLite2 databases and extract the files into the dbPath.
func downloadGeoLite2Db(maxmindDBLicense, maxmindDBType string) error {
	// Download the GeoLite2Db .tar.gz file with the database inside it.
	r, err := http.Get(fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-%s&license_key=%s&suffix=tar.gz", maxmindDBType, maxmindDBLicense))
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
		if ok, _ := regexp.MatchString("GeoLite2-([a-zA-z]+)\\.mmdb", i.Name()); ok {
			// Move from temporary directory to geoip.geoLite2DbName to geoip.dbPath.
			err := os.Rename(p, dbPath+i.Name())
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

// NewGeoLite2 opens connections to GeoIp2 databases and return a geoLite2 structure with the databases connections.
//
// The connection uses the local database or try to download it from MaxMind's server (to download, it is required `MAXMIND_LICENSE` set).
func NewGeoLite2() (Locator, error) {
	for _, info := range geoLite2Info {
		if _, err := os.Stat(dbPath + info["file"]); os.IsNotExist(err) {
			if license, ok := os.LookupEnv("MAXMIND_LICENSE"); ok {
				err := downloadGeoLite2Db(license, info["type"])
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New("geoip feature is enable, but MAXMIND_LICENSE is not set")
			}
		}
	}

	geolite2Db := new(geoLite2)
	for _, info := range geoLite2Info {
		db, err := geoip2.Open(dbPath + info["file"])
		if err != nil {
			return nil, err
		}

		geolite2Db.db = append(geolite2Db.db, db)
	}

	return geolite2Db, nil
}

// Close the connection with the GeoLite2 database, returning either error or nil.
func (g *geoLite2) Close() error {
	for i := range geoLite2Info {
		err := g.db[i].Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetCountry gets an ip and return either an ISO 3166-1 code to a country or an empty string.
func (g *geoLite2) GetCountry(ip net.IP) (string, error) {
	record, err := g.db[country].Country(ip)
	if err != nil {
		return "", err
	}

	return record.Country.IsoCode, nil
}

// GetPosition gets an ip and return a Position structure with Longitude and Latitude with error nil or an empty Position structure with the error.
func (g *geoLite2) GetPosition(ip net.IP) (Position, error) {
	record, err := g.db[city].City(ip)
	if err != nil {
		return Position{}, err
	}

	return Position{
		Longitude: record.Location.Longitude,
		Latitude:  record.Location.Latitude,
	}, nil
}
