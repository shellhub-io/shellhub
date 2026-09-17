package environment

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Edition selects which ShellHub compose overlays and environment a stack runs.
type Edition string

const (
	// EditionCommunity is the open-source, self-hosted edition.
	EditionCommunity Edition = "community"
	// EditionEnterprise adds the enterprise overlay (redis, minio, enterprise server image).
	EditionEnterprise Edition = "enterprise"
	// EditionCloud adds the cloud overlay on top of enterprise (billing, email, cloud endpoints).
	EditionCloud Edition = "cloud"
)

// ParseEdition converts a string to an [Edition], returning an error for unknown values.
func ParseEdition(s string) (Edition, error) {
	switch Edition(s) {
	case EditionCommunity:
		return EditionCommunity, nil
	case EditionEnterprise:
		return EditionEnterprise, nil
	case EditionCloud:
		return EditionCloud, nil
	default:
		return "", fmt.Errorf("unknown edition %q: must be community, enterprise, or cloud", s)
	}
}

// Build returns the build-time edition flag: community builds as "community", enterprise and
// cloud both build as "enterprise" (the server binary is the same for both).
func (e Edition) Build() string {
	if e == EditionCommunity {
		return "community"
	}

	return "enterprise"
}

func (e Edition) composeFiles(cloudDir string) ([]string, error) {
	base := []string{"../docker-compose.yml"}

	switch e {
	case EditionCommunity:
	case EditionEnterprise:
		base = append(base, "../docker-compose.enterprise.yml")
	case EditionCloud:
		cloudCompose := filepath.Join(cloudDir, "docker-compose.yml")
		if _, err := os.Stat(cloudCompose); err != nil {
			return nil, fmt.Errorf("cloud edition requires %s: %w", cloudCompose, err)
		}

		base = append(base, "../docker-compose.enterprise.yml", cloudCompose)
	}

	base = append(base, "../docker-compose.test.yml", "../docker-compose.postgres.test.yml")

	if e != EditionCommunity {
		base = append(base, "../docker-compose.enterprise.test.yml")
	}

	return base, nil
}

func (e Edition) envFiles(cloudDir string) []string {
	files := []string{"../.env"}

	if e == EditionCommunity {
		return files
	}

	files = append(files, "../.env.enterprise")

	cloudEnv := filepath.Join(cloudDir, ".env")
	if _, err := os.Stat(cloudEnv); err == nil {
		files = append(files, cloudEnv)
	}

	return files
}

func (e Edition) envs(cloudDir string) (map[string]string, error) {
	cloudSrc := "."
	if e != EditionCommunity {
		abs, err := filepath.Abs(cloudDir)
		if err != nil {
			return nil, fmt.Errorf("resolving cloud dir: %w", err)
		}

		cloudSrc = abs
	}

	envs := map[string]string{
		"SHELLHUB_EDITION":       string(e),
		"SHELLHUB_BUILD_EDITION": e.Build(),
		"SHELLHUB_CLOUD_SRC":     cloudSrc,
		"SHELLHUB_ENV":           "production",
		"SHELLHUB_BIND_ADDRESS":  "127.0.0.1",
		"SHELLHUB_LOG_LEVEL":     "trace",
	}

	if e != EditionCommunity {
		envs["SHELLHUB_BILLING"] = "dummy"
		envs["SHELLHUB_EMAIL_PROVIDER"] = "dummy"
		envs["SHELLHUB_MAXMIND_MIRROR"] = ""
	}

	return envs, nil
}

func (e Edition) openapiSpec() string {
	switch e {
	case EditionCommunity:
		return "community-openapi.yaml"
	case EditionEnterprise:
		return "enterprise-openapi.yaml"
	case EditionCloud:
		return "cloud-openapi.yaml"
	default:
		return "community-openapi.yaml"
	}
}

// OpenAPIPath returns the output path for the bundled OpenAPI spec, relative to the repo
// root. Each edition writes to a separate file so concurrent stacks don't collide.
func (e Edition) OpenAPIPath() string {
	return filepath.Join("openapi", "static", "openapi-"+string(e)+".json")
}

// BundleOpenAPI runs redocly to bundle the edition's OpenAPI spec into [Edition.OpenAPIPath].
// repoRoot is the path to the repository root (e.g. ".." from tests/).
func BundleOpenAPI(ctx context.Context, edition Edition, repoRoot string) error {
	cmd := exec.CommandContext(ctx, //nolint:gosec // args are string literals
		"npx", "-y", "@redocly/cli@2.31.5", "bundle",
		filepath.Join("openapi", "spec", edition.openapiSpec()),
		"-o", edition.OpenAPIPath(),
	)
	cmd.Dir = repoRoot
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
