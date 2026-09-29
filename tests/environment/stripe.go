package environment

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

var (
	errNoStripeCLIImage = errors.New("declares no stripe-cli image")
	errNoWebhookSecret  = errors.New("reading the Stripe webhook secret: stripe listen printed none")
)

var stripeKeyNames = []string{"STRIPE_SECRET_KEY", "STRIPE_PRICE_ID", "SHELLHUB_STRIPE_PUBLISHABLE_KEY"}

func stripeEnvs(ctx context.Context, cloudDir string) (map[string]string, error) {
	envs, err := stripeKeys("../.env.override")
	if err != nil {
		return nil, err
	}

	image, err := stripeCLIImage(filepath.Join(cloudDir, "docker-compose.yml"))
	if err != nil {
		return nil, err
	}

	secret, err := stripeWebhookSecret(ctx, image, envs["STRIPE_SECRET_KEY"])
	if err != nil {
		return nil, err
	}

	envs["STRIPE_WEBHOOK_SECRET"] = secret

	return envs, nil
}

func stripeKeys(overridePath string) (map[string]string, error) {
	override, err := godotenv.Read(overridePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("reading the Stripe keys from %s: %w", overridePath, err)
	}

	keys := make(map[string]string, len(stripeKeyNames))
	for _, name := range stripeKeyNames {
		value := os.Getenv(name)
		if value == "" {
			value = override[name]
		}

		if value == "" {
			return nil, fmt.Errorf("the cloud edition bills against Stripe test mode: set %s in the shell or %s", name, overridePath)
		}

		keys[name] = value
	}

	return keys, nil
}

func stripeCLIImage(composePath string) (string, error) {
	data, err := os.ReadFile(composePath) //nolint:gosec // the cloud compose file the stack itself loads
	if err != nil {
		return "", fmt.Errorf("reading the stripe-cli image: %w", err)
	}

	var compose struct {
		Services map[string]struct {
			Image string `yaml:"image"`
		} `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return "", fmt.Errorf("parsing %s: %w", composePath, err)
	}

	image := compose.Services["stripe-cli"].Image
	if image == "" {
		return "", fmt.Errorf("%s %w", composePath, errNoStripeCLIImage)
	}

	return image, nil
}

func stripeWebhookSecret(ctx context.Context, image, secretKey string) (string, error) {
	cmd := exec.CommandContext(ctx, //nolint:gosec // image comes from the cloud compose file the stack loads
		"docker", "run", "--rm", "-e", "STRIPE_API_KEY",
		image, "listen", "--print-secret")
	cmd.Env = append(os.Environ(), "STRIPE_API_KEY="+secretKey)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("reading the Stripe webhook secret: %w", err)
	}

	secret := strings.TrimSpace(string(out))
	if secret == "" {
		return "", errNoWebhookSecret
	}

	return secret, nil
}
