package main

import (
	"errors"
	"fmt"
	"net/netip"
	"os"

	"github.com/oschwald/geoip2-golang/v2"
)

func check(path string) (string, error) {
	reader, err := geoip2.Open(path)
	if err != nil {
		return "", fmt.Errorf("not a GeoIP database: %w", err)
	}
	defer reader.Close() //nolint:errcheck // the database is only read

	record, err := reader.City(netip.MustParseAddr("8.8.8.8"))
	if err != nil {
		return "", err
	}

	if record.Location.Latitude == nil {
		return "", errors.New("the database places no address")
	}

	metadata := reader.Metadata()

	return metadata.DatabaseType + " built " + metadata.BuildTime().UTC().Format("2006-01-02"), nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: check DATABASE")
		os.Exit(2)
	}

	summary, err := check(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if _, err := fmt.Fprintln(os.Stdout, summary); err != nil {
		os.Exit(1)
	}
}
