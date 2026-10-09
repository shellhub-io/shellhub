package environment

import (
	"os"
	"path/filepath"
)

func (cfg Config) licensingEnvs() (map[string]string, []string, error) {
	issuer := cfg.Run.LicenseIssuer()

	if cfg.Unlicensed && cfg.License != nil {
		return nil, nil, errLicensedAndUnlicensed
	}

	var ldflags map[string]string
	if issuer != nil {
		flags, err := issuer.ldflags()
		if err != nil {
			return nil, nil, err
		}

		ldflags = map[string]string{goLDFlagsEnv: flags}
	}

	switch {
	case cfg.Unlicensed:
		envs, err := mergeEnvs(nil, ldflags, map[string]string{licenseFileEnv: ""})

		return envs, nil, err
	case issuer == nil:
		return nil, nil, errRunIssuesNoLicense
	}

	license := FullLicense()
	if cfg.License != nil {
		license = *cfg.License
	}

	signed, err := issuer.Sign(license)
	if err != nil {
		return nil, nil, err
	}

	path, err := licenseFilePath(cfg.Name)
	if err != nil {
		return nil, nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, nil, err
	}

	if err := os.WriteFile(path, signed, 0o644); err != nil { //nolint:gosec // the server container reads the file as whatever user it runs as, and a test license is no secret
		return nil, []string{path}, err
	}

	envs, err := mergeEnvs(nil, ldflags, map[string]string{licenseFileEnv: path})

	return envs, []string{path}, err
}

func licenseFilePath(stack string) (string, error) {
	return filepath.Abs(filepath.Join(stackArtifactsDir, "licenses", stack+".dat"))
}
