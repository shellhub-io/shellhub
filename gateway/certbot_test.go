package main

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	gatewayMocks "github.com/shellhub-io/shellhub/gateway/mocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTunnelsCertificate_generateProviderCredentialsFile(t *testing.T) {
	certificate := TunnelsCertificate{
		Domain:   "localhost",
		Provider: "digitalocean",
		Token:    "test",
	}

	certificate.fs = afero.NewMemMapFs()

	certificate.generateProviderCredentialsFile()

	buffer, err := afero.ReadFile(certificate.fs, "/etc/shellhub-gateway/digitalocean.ini")
	assert.NoError(t, err)

	assert.Equal(t, "dns_digitalocean_token = test", string(buffer))
}

func TestTunnelsCertificate_generate(t *testing.T) {
	tests := []struct {
		name        string
		config      TunnelsCertificate
		staging     bool
		expected    error
		expectCalls func(*gatewayMocks.Executor)
	}{
		{
			name: "failed to run the command",
			config: TunnelsCertificate{
				Domain:   "localhost",
				Provider: "digitalocean",
				Token:    "test",
			},
			expectCalls: func(executorMock *gatewayMocks.Executor) {
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
			config: TunnelsCertificate{
				Domain:   "localhost",
				Provider: "digitalocean",
				Token:    "test",
			},
			expectCalls: func(executorMock *gatewayMocks.Executor) {
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
			config: TunnelsCertificate{
				Domain:   "localhost",
				Provider: "digitalocean",
				Token:    "test",
			},
			staging: true,
			expectCalls: func(executorMock *gatewayMocks.Executor) {
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
			executorMock := new(gatewayMocks.Executor)

			certificate := tc.config
			certificate.fs = afero.NewMemMapFs()
			certificate.ex = executorMock

			tc.expectCalls(executorMock)

			err := certificate.Generate(tc.staging)
			assert.Equal(tt, tc.expected, err)

			executorMock.AssertExpectations(t)
		})
	}
}

func TestCertBot_executeRenewCertificates(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expected    error
		expectCalls func(*gatewayMocks.Executor)
	}{
		{
			name: "failed to run the renew command",
			config: Config{
				Staging: false,
			},
			expectCalls: func(executorMock *gatewayMocks.Executor) {
				executorMock.On("Command", "certbot", "renew").Return(exec.Command("")).Once()
				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(errors.New("failed to run the renew command")).Once()
			},
			expected: errors.New("failed to run the renew command"),
		},
		{
			name: "successful renew command execution",
			config: Config{
				Staging: false,
			},
			expectCalls: func(executorMock *gatewayMocks.Executor) {
				executorMock.On("Command", "certbot", "renew").Return(exec.Command("")).Once()
				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(nil).Once()
			},
			expected: nil,
		},
		{
			name: "successful renew command execution in staging",
			config: Config{
				Staging: true,
			},
			expectCalls: func(executorMock *gatewayMocks.Executor) {
				executorMock.On("Command", "certbot", "renew", "--staging").Return(exec.Command("")).Once()
				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			executorMock := new(gatewayMocks.Executor)

			certbot := newCertBot(&tc.config)
			certbot.ex = executorMock

			tc.expectCalls(executorMock)

			err := certbot.executeRenewCertificates()
			assert.Equal(tt, tc.expected, err)

			executorMock.AssertExpectations(t)
		})
	}
}

func TestCertBot_renewCertificates(t *testing.T) {
	duration := 100 * time.Millisecond

	tests := []struct {
		name              string
		config            Config
		expectCalls       func(*gatewayMocks.Executor, *gatewayMocks.Ticker)
		shouldRenewCalled bool
	}{
		{
			name: "failed renewal",
			config: Config{
				Staging: false,
			},
			expectCalls: func(executorMock *gatewayMocks.Executor, tickerMock *gatewayMocks.Ticker) {
				tickerMock.On("Init", mock.Anything, mock.Anything).Once()
				tickerMock.On("Stop").Once()

				ch := make(chan time.Time, 1)
				ch <- time.Now()
				tickerMock.On("Tick").Return(ch).Once()

				executorMock.On("Command", "certbot",
					"renew",
				).Return(exec.Command("")).Once()

				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(errors.New("failed to renew")).Once()
			},
			shouldRenewCalled: false,
		},
		{
			name: "failed renewal more than run time",
			config: Config{
				Staging: false,
			},
			expectCalls: func(executorMock *gatewayMocks.Executor, tickerMock *gatewayMocks.Ticker) {
				tickerMock.On("Init", mock.Anything, mock.Anything).Once()
				tickerMock.On("Stop").Once()

				ch := make(chan time.Time, 2)
				ch <- time.Now()
				ch <- time.Now()
				tickerMock.On("Tick").Return(ch).Once()

				executorMock.On("Command", "certbot",
					"renew",
				).Return(exec.Command("")).Twice()

				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(errors.New("failed to renew")).Twice()
			},
			shouldRenewCalled: false,
		},
		{
			name: "success to renew after failure",
			config: Config{
				Staging: false,
			},
			expectCalls: func(executorMock *gatewayMocks.Executor, tickerMock *gatewayMocks.Ticker) {
				tickerMock.On("Init", mock.Anything, mock.Anything).Once()
				tickerMock.On("Stop").Once()

				ch := make(chan time.Time, 2)
				ch <- time.Now()
				tickerMock.On("Tick").Return(ch).Once()

				executorMock.On("Command", "certbot",
					"renew",
				).Return(exec.Command("")).Once()

				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(errors.New("failed to renew")).Once()

				ch <- time.Now()
				executorMock.On("Command", "certbot",
					"renew",
				).Return(exec.Command("")).Once()

				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(nil).Once()
			},
			shouldRenewCalled: true,
		},
		{
			name: "success to renew",
			config: Config{
				Staging: false,
			},
			expectCalls: func(executorMock *gatewayMocks.Executor, tickerMock *gatewayMocks.Ticker) {
				tickerMock.On("Init", mock.Anything, mock.Anything).Once()
				tickerMock.On("Stop").Once()

				ch := make(chan time.Time, 1)
				ch <- time.Now()
				tickerMock.On("Tick").Return(ch).Once()

				executorMock.On("Command", "certbot",
					"renew",
				).Return(exec.Command("")).Once()

				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(nil).Once()
			},
			shouldRenewCalled: true,
		},
		{
			name: "success to renew more than one time",
			config: Config{
				Staging: false,
			},
			expectCalls: func(executorMock *gatewayMocks.Executor, tickerMock *gatewayMocks.Ticker) {
				tickerMock.On("Init", mock.Anything, mock.Anything).Once()
				tickerMock.On("Stop").Once()

				ch := make(chan time.Time, 2)
				ch <- time.Now()
				ch <- time.Now()
				tickerMock.On("Tick").Return(ch).Once()

				executorMock.On("Command", "certbot",
					"renew",
				).Return(exec.Command("")).Twice()

				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(nil).Twice()
			},
			shouldRenewCalled: true,
		},
		{
			name: "success to renew on staging",
			config: Config{
				Staging: true,
			},
			expectCalls: func(executorMock *gatewayMocks.Executor, tickerMock *gatewayMocks.Ticker) {
				tickerMock.On("Init", mock.Anything, mock.Anything).Once()
				tickerMock.On("Stop").Once()

				ch := make(chan time.Time, 1)
				ch <- time.Now()
				tickerMock.On("Tick").Return(ch).Once()

				executorMock.On("Command", "certbot",
					"renew",
					"--staging",
				).Return(exec.Command("")).Once()

				executorMock.On("Run", mock.AnythingOfType("*exec.Cmd")).Return(nil).Once()
			},
			shouldRenewCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), duration)
			defer cancel()

			tickerMock := new(gatewayMocks.Ticker)
			executorMock := new(gatewayMocks.Executor)

			config := &tc.config

			renewWasCalled := false
			config.RenewedCallback = func() {
				renewWasCalled = true
			}

			certbot := newCertBot(config)
			certbot.tk = tickerMock
			certbot.ex = executorMock

			tc.expectCalls(executorMock, tickerMock)

			done := make(chan struct{})
			go func() {
				certbot.renewCertificates(ctx, duration)
				close(done)
			}()

			<-done

			assert.Equal(tt, tc.shouldRenewCalled, renewWasCalled)

			tickerMock.AssertExpectations(tt)
			executorMock.AssertExpectations(tt)
		})
	}
}
