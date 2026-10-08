package environment

import (
	"net"
	"os"
	"path/filepath"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

const (
	geoIPDirEnv    = "SHELLHUB_TEST_GEOIP_DIR"
	geoIPMirrorEnv = "SHELLHUB_MAXMIND_MIRROR"

	unreachableGeoIPMirror = "http://geoip.invalid"
)

func (cfg Config) geoIPEnvs() (map[string]string, string, error) {
	dir, err := geoIPDir(cfg.Name)
	if err != nil {
		return nil, "", err
	}

	if err := writeGeoIPDatabases(dir, cfg.LocatedCountry); err != nil {
		return nil, dir, err
	}

	return map[string]string{
		geoIPDirEnv:    dir,
		geoIPMirrorEnv: unreachableGeoIPMirror,
	}, dir, nil
}

func geoIPDir(stack string) (string, error) {
	return filepath.Abs(filepath.Join(stackArtifactsDir, "geoip", stack))
}

func writeGeoIPDatabases(dir, country string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	located := mmdbtype.Map{"iso_code": mmdbtype.String(country)}

	databases := map[string]mmdbtype.Map{
		"GeoLite2-Country": {"country": located},
		"GeoLite2-City": {
			"country":  located,
			"location": mmdbtype.Map{"latitude": mmdbtype.Float64(0), "longitude": mmdbtype.Float64(0)},
		},
	}

	for kind, record := range databases {
		if err := writeGeoIPDatabase(filepath.Join(dir, kind+".mmdb"), kind, record); err != nil {
			return err
		}
	}

	return nil
}

func writeGeoIPDatabase(path, kind string, record mmdbtype.Map) error {
	tree, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType:            kind,
		IncludeReservedNetworks: true,
		DisableIPv4Aliasing:     true,
	})
	if err != nil {
		return err
	}

	_, everyAddress, err := net.ParseCIDR("::/0")
	if err != nil {
		return err
	}

	if err := tree.Insert(everyAddress, record); err != nil {
		return err
	}

	file, err := os.Create(path) //nolint:gosec // the path is the stack's own GeoIP directory, built from its name
	if err != nil {
		return err
	}

	if _, err := tree.WriteTo(file); err != nil {
		_ = file.Close()

		return err
	}

	return file.Close()
}
