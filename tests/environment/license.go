package environment

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const licenseFileEnv = "SHELLHUB_LICENSE_FILE"

var errNoLicense = errors.New("enterprise and cloud accept devices and open the admin panel only under a license")

func licenseEnvs(overridePath string) (map[string]string, error) {
	values, err := shellOrOverride(overridePath, licenseFileEnv)
	if err != nil {
		return nil, err
	}

	path := values[licenseFileEnv]
	if path == "" {
		return nil, fmt.Errorf("%w: set %s in the shell or %s", errNoLicense, licenseFileEnv, overridePath)
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(overridePath), path)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving %s: %w", licenseFileEnv, err)
	}

	if _, err := os.Stat(abs); err != nil {
		return nil, fmt.Errorf("%s: %w", licenseFileEnv, err)
	}

	return map[string]string{licenseFileEnv: abs}, nil
}
