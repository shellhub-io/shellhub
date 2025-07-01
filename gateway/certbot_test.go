package main

import (
	"errors"
	"os/exec"
	"testing"

	executorMock "github.com/shellhub-io/shellhub/gateway/mocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestCertBot_generateCertificateFromDNS(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expected    error
		expectCalls func(*executorMock.Executor)
	}{
		{
			name: "failed to run the command",
			config: Config{
				Tunnels: &Tunnels{
					Domain:   "localhost",
					Provider: "digitalocean",
					Token:    "test",
				},
			},
			expectCalls: func(executorMock *executorMock.Executor) {
				executorMock.On("Command", "certbot",
					"certonly",
					"--non-interactive",
					"--agree-tos",
					"--register-unsafely-without-email",
					"--cert-name",
					"*.localhost",
					"--dns-digitalocean",
					"--dns-digitalocean-credentials",
					"/etc/shellhub-gateway/digitalocean.ini",
					"-d",
					"*.localhost",
				).Return(exec.Command("")).Once()

				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(errors.New("failed to run the command")).Once()
			},
			expected: errors.New("failed to run the command"),
		},
		{
			name: "successful certificate generation",
			config: Config{
				Tunnels: &Tunnels{
					Domain:   "localhost",
					Provider: "digitalocean",
					Token:    "test",
				},
			},
			expectCalls: func(executorMock *executorMock.Executor) {
				executorMock.On("Command", "certbot",
					"certonly",
					"--non-interactive",
					"--agree-tos",
					"--register-unsafely-without-email",
					"--cert-name",
					"*.localhost",
					"--dns-digitalocean",
					"--dns-digitalocean-credentials",
					"/etc/shellhub-gateway/digitalocean.ini",
					"-d",
					"*.localhost",
				).Return(exec.Command("")).Once()

				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(nil).Once()
			},
			expected: nil,
		},
		{
			name: "successful certificate generation in staging",
			config: Config{
				Tunnels: &Tunnels{
					Domain:   "localhost",
					Provider: "digitalocean",
					Token:    "test",
				},
				Staging: true,
			},
			expectCalls: func(executorMock *executorMock.Executor) {
				executorMock.On("Command", "certbot",
					"certonly",
					"--non-interactive",
					"--agree-tos",
					"--register-unsafely-without-email",
					"--cert-name",
					"*.localhost",
					"--dns-digitalocean",
					"--dns-digitalocean-credentials",
					"/etc/shellhub-gateway/digitalocean.ini",
					"-d",
					"*.localhost",
					"--staging",
				).Return(exec.Command("")).Once()

				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			executorMock := new(executorMock.Executor)

			certbot := newCertBot(&tc.config)
			certbot.fs = afero.NewMemMapFs()
			certbot.ex = executorMock

			tc.expectCalls(executorMock)

			err := certbot.generateCertificateFromDNS()
			assert.Equal(tt, tc.expected, err)

			executorMock.AssertExpectations(t)
		})
	}
}
