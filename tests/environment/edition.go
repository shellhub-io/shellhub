package environment

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shellhub-io/shellhub/pkg/testimage"
)

// Edition selects which ShellHub compose overlays and environment a stack runs.
type Edition string

const (
	// EditionCommunity is the open-source, self-hosted edition.
	EditionCommunity Edition = "community"
	// EditionEnterprise adds the enterprise overlay (redis, minio, enterprise server image).
	EditionEnterprise Edition = "enterprise"
	// EditionCloud adds the cloud overlay on top of enterprise (billing, cloud endpoints) and
	// delivers its email over SMTP to a Mailpit in the stack.
	EditionCloud Edition = "cloud"
)

type editionProfile struct {
	paid        bool
	cloud       bool
	openapiSpec string
}

var editionProfiles = map[Edition]editionProfile{
	EditionCommunity:  {openapiSpec: "community-openapi.yaml"},
	EditionEnterprise: {paid: true, openapiSpec: "enterprise-openapi.yaml"},
	EditionCloud:      {paid: true, cloud: true, openapiSpec: "cloud-openapi.yaml"},
}

func (e Edition) profile() editionProfile {
	return editionProfiles[e]
}

// ParseEdition converts a string to an [Edition], returning an error for unknown values.
func ParseEdition(s string) (Edition, error) {
	if _, ok := editionProfiles[Edition(s)]; !ok {
		return "", fmt.Errorf("unknown edition %q: must be community, enterprise, or cloud", s)
	}

	return Edition(s), nil
}

// Build returns the build-time edition flag: community builds as "community", enterprise and
// cloud both build as "enterprise" (the server binary is the same for both).
func (e Edition) Build() string {
	if e.profile().paid {
		return "enterprise"
	}

	return "community"
}

func statCloudSource(cloudDir string) error {
	_, err := os.Stat(filepath.Join(cloudDir, "go.mod"))

	return err
}

func (e Edition) composeFiles(cloudDir string) ([]string, error) {
	profile := e.profile()
	files := []string{"../docker-compose.yml"}

	if profile.paid {
		if err := statCloudSource(cloudDir); err != nil {
			return nil, fmt.Errorf("the %s edition builds from the cloud source: %w", e, err)
		}

		files = append(files, "../docker-compose.enterprise.yml")
	}

	if profile.cloud {
		cloudCompose := filepath.Join(cloudDir, "docker-compose.yml")
		if _, err := os.Stat(cloudCompose); err != nil {
			return nil, fmt.Errorf("cloud edition requires %s: %w", cloudCompose, err)
		}

		files = append(files, cloudCompose)
	}

	files = append(files, "../docker-compose.test.yml", "../docker-compose.postgres.test.yml")

	if profile.paid {
		files = append(files, "../docker-compose.enterprise.test.yml")
	}

	if profile.cloud {
		files = append(files, "../docker-compose.cloud.test.yml")
	}

	return files, nil
}

func (e Edition) envFiles(cloudDir string) []string {
	files := []string{"../.env", "../" + testimage.FileName}

	if !e.profile().paid {
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
	if e.profile().paid {
		abs, err := filepath.Abs(cloudDir)
		if err != nil {
			return nil, fmt.Errorf("resolving cloud dir: %w", err)
		}

		cloudSrc = abs
	}

	envs := map[string]string{
		"SHELLHUB_EDITION":        string(e),
		"SHELLHUB_BUILD_EDITION":  e.Build(),
		"SHELLHUB_CLOUD_SRC":      cloudSrc,
		"SHELLHUB_ENV":            "production",
		"SHELLHUB_BIND_ADDRESS":   "127.0.0.1",
		"SHELLHUB_LOG_LEVEL":      "trace",
		"SHELLHUB_TELEMETRY":      "false",
		"SHELLHUB_ONBOARDING_URL": "",
		testimage.MirrorEnv:       testimage.Mirror(),
	}

	if e.profile().paid {
		envs["SHELLHUB_MAXMIND_MIRROR"] = ""
		envs["SHELLHUB_SAML_SECRET"] = buildSAMLSecret()
	}

	if e.profile().cloud {
		envs["SHELLHUB_BILLING"] = "stripe"
		envs["SHELLHUB_DISPOSABLE_EMAIL_CHECK"] = "false"
		envs["SHELLHUB_EMAIL_PROVIDER"] = "smtp"
		envs["SHELLHUB_EMAIL_SMTP_HOST"] = "mailpit"
		envs["SHELLHUB_EMAIL_SMTP_PORT"] = "1025"
	}

	return envs, nil
}

func buildSAMLSecret() string {
	key := make([]byte, 32)
	rand.Read(key)

	return base64.URLEncoding.EncodeToString(key)
}

// OpenAPIPath returns the output path for the bundled OpenAPI spec, relative to the repo
// root. Each edition writes to a separate file so concurrent stacks don't collide.
func (e Edition) OpenAPIPath() string {
	return filepath.Join("openapi", "static", "openapi-"+string(e)+".json")
}

// BundleOpenAPI runs redocly to bundle the edition's OpenAPI spec into [Edition.OpenAPIPath].
// repoRoot is the path to the repository root (e.g. ".." from tests/). It returns an error when
// versions.env there cannot be read, has a line that is not KEY=VALUE or does not pin
// REDOCLY_VERSION, or the error redocly exits with; cancelling ctx kills redocly.
func BundleOpenAPI(ctx context.Context, edition Edition, repoRoot string) error {
	redocly, err := testimage.Version(repoRoot, "REDOCLY_VERSION")
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, //nolint:gosec // the arguments are literals and a version read from versions.env
		"npx", "-y", "@redocly/cli@"+redocly, "bundle",
		filepath.Join("openapi", "spec", edition.profile().openapiSpec),
		"-o", edition.OpenAPIPath(),
	)
	cmd.Dir = repoRoot
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
