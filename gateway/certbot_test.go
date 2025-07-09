package main

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestCertBot_generateProviderCredentialsFile(t *testing.T) {
	certbot := newCertBot(&Config{
		Tunnels: &Tunnels{
			Domain:   "localhost",
			Provider: "digitalocean",
			Token:    "test",
		},
	})
	certbot.fs = afero.NewMemMapFs()

	certbot.generateProviderCredentialsFile()

	buffer, err := afero.ReadFile(certbot.fs, "/etc/shellhub-gateway/digitalocean.ini")
	assert.NoError(t, err)

	assert.Equal(t, "dns_digitalocean_token = test", string(buffer))
}
