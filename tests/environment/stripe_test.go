package environment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStripeKeys(t *testing.T) {
	writeOverride := func(t *testing.T, content string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), ".env.override")
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

		return path
	}

	allKeys := "STRIPE_SECRET_KEY=sk_file\nSTRIPE_PRICE_ID=price_file\nSHELLHUB_STRIPE_PUBLISHABLE_KEY=pk_file\n"

	tests := []struct {
		name     string
		override string
		shell    map[string]string
		want     map[string]string
	}{
		{
			name:     "reads the keys from the override file",
			override: allKeys + "SHELLHUB_BILLING=dummy\n",
			want: map[string]string{
				"STRIPE_SECRET_KEY":               "sk_file",
				"STRIPE_PRICE_ID":                 "price_file",
				"SHELLHUB_STRIPE_PUBLISHABLE_KEY": "pk_file",
			},
		},
		{
			name:     "prefers the shell over the override file",
			override: allKeys,
			shell:    map[string]string{"STRIPE_SECRET_KEY": "sk_shell"},
			want: map[string]string{
				"STRIPE_SECRET_KEY":               "sk_shell",
				"STRIPE_PRICE_ID":                 "price_file",
				"SHELLHUB_STRIPE_PUBLISHABLE_KEY": "pk_file",
			},
		},
	}

	setShell := func(t *testing.T, shell map[string]string) {
		t.Helper()
		for _, name := range stripeKeyNames {
			t.Setenv(name, shell[name])
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setShell(t, tt.shell)

			keys, err := stripeKeys(writeOverride(t, tt.override))
			require.NoError(t, err)
			assert.Equal(t, tt.want, keys)
		})
	}

	t.Run("reads the keys from the shell without an override file", func(t *testing.T) {
		shell := map[string]string{
			"STRIPE_SECRET_KEY":               "sk_shell",
			"STRIPE_PRICE_ID":                 "price_shell",
			"SHELLHUB_STRIPE_PUBLISHABLE_KEY": "pk_shell",
		}
		setShell(t, shell)

		keys, err := stripeKeys(filepath.Join(t.TempDir(), ".env.override"))
		require.NoError(t, err)
		assert.Equal(t, shell, keys)
	})

	t.Run("names the missing key", func(t *testing.T) {
		setShell(t, nil)

		_, err := stripeKeys(writeOverride(t, "STRIPE_SECRET_KEY=sk_file\nSTRIPE_PRICE_ID=price_file\n"))
		require.ErrorContains(t, err, "set SHELLHUB_STRIPE_PUBLISHABLE_KEY")
	})
}

func TestStripeCLIImage(t *testing.T) {
	tests := []struct {
		name    string
		compose string
		want    string
		wantErr error
	}{
		{
			name:    "reads the stripe-cli service's image",
			compose: "services:\n  server:\n    image: server\n  stripe-cli:\n    image: stripe/stripe-cli:v9.9.9\n",
			want:    "stripe/stripe-cli:v9.9.9",
		},
		{
			name:    "fails without a stripe-cli service",
			compose: "services:\n  server:\n    image: server\n",
			wantErr: errNoStripeCLIImage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "docker-compose.yml")
			require.NoError(t, os.WriteFile(path, []byte(tt.compose), 0o600))

			image, err := stripeCLIImage(path)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, image)
		})
	}
}
