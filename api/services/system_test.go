package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/responses"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envmocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestGetSystemInfo(t *testing.T) {
	// Save and restore global hooks so tests don't leak.
	savedHooks := systemInfoHooks
	t.Cleanup(func() { systemInfoHooks = savedHooks })

	// Use a local env mock to avoid interference with the shared package-level envMock.
	localEnvMock := new(envmocks.Backend)
	savedEnvBackend := envs.DefaultBackend
	envs.DefaultBackend = localEnvMock
	t.Cleanup(func() { envs.DefaultBackend = savedEnvBackend })

	storeMock := &mocks.Store{}
	s := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	ctx := context.Background()

	cases := []struct {
		description    string
		req            *requests.GetSystemInfo
		setupMocks     func()
		setupHooks     func()
		expectedResult *responses.SystemInfo
		expectedErrMsg string
	}{
		{
			description: "store error propagates",
			req:         &requests.GetSystemInfo{Host: "localhost"},
			setupMocks: func() {
				storeMock.On("SystemGet", gomock.Anything).
					Return(nil, errors.New("db error")).Once()
			},
			setupHooks:     func() {},
			expectedResult: nil,
			expectedErrMsg: "db error",
		},
		{
			description: "returns local auth enabled, saml false when no hooks",
			req:         &requests.GetSystemInfo{Host: "192.168.1.1"},
			setupMocks: func() {
				localEnvMock.On("Get", "SHELLHUB_SSH_PORT").Return("22").Once()
				localEnvMock.On("Get", "SHELLHUB_VERSION").Return("v1.0.0").Once()
				storeMock.On("SystemGet", gomock.Anything).Return(&models.System{
					Setup: true,
					Authentication: &models.SystemAuthentication{
						Local: &models.SystemAuthenticationLocal{Enabled: true},
					},
				}, nil).Once()
			},
			setupHooks: func() {},
			expectedResult: &responses.SystemInfo{
				Version: "v1.0.0",
				Setup:   true,
				Endpoints: &responses.SystemEndpointsInfo{
					API: "192.168.1.1",
					SSH: "192.168.1.1:22",
				},
				Authentication: &responses.SystemAuthenticationInfo{
					Local: true,
					SAML:  false,
				},
			},
		},
		{
			description: "returns local auth disabled, saml false when no hooks",
			req:         &requests.GetSystemInfo{Host: "example.com"},
			setupMocks: func() {
				localEnvMock.On("Get", "SHELLHUB_SSH_PORT").Return("2222").Once()
				localEnvMock.On("Get", "SHELLHUB_VERSION").Return("v1.0.0").Once()
				storeMock.On("SystemGet", gomock.Anything).Return(&models.System{
					Setup: false,
					Authentication: &models.SystemAuthentication{
						Local: &models.SystemAuthenticationLocal{Enabled: false},
					},
				}, nil).Once()
			},
			setupHooks: func() {},
			expectedResult: &responses.SystemInfo{
				Version: "v1.0.0",
				Setup:   false,
				Endpoints: &responses.SystemEndpointsInfo{
					API: "example.com",
					SSH: "example.com:2222",
				},
				Authentication: &responses.SystemAuthenticationInfo{
					Local: false,
					SAML:  false,
				},
			},
		},
		{
			description: "hook sets saml true when enterprise has saml enabled",
			req:         &requests.GetSystemInfo{Host: "192.168.1.1"},
			setupMocks: func() {
				localEnvMock.On("Get", "SHELLHUB_SSH_PORT").Return("22").Once()
				localEnvMock.On("Get", "SHELLHUB_VERSION").Return("v1.0.0").Once()
				storeMock.On("SystemGet", gomock.Anything).Return(&models.System{
					Setup: true,
					Authentication: &models.SystemAuthentication{
						Local: &models.SystemAuthenticationLocal{Enabled: true},
					},
				}, nil).Once()
			},
			setupHooks: func() {
				OnGetSystemInfo(func(_ context.Context, info *responses.SystemInfo) error {
					info.Authentication.SAML = true

					return nil
				})
			},
			expectedResult: &responses.SystemInfo{
				Version: "v1.0.0",
				Setup:   true,
				Endpoints: &responses.SystemEndpointsInfo{
					API: "192.168.1.1",
					SSH: "192.168.1.1:22",
				},
				Authentication: &responses.SystemAuthenticationInfo{
					Local: true,
					SAML:  true,
				},
			},
		},
		{
			description: "hook error is returned",
			req:         &requests.GetSystemInfo{Host: "192.168.1.1"},
			setupMocks: func() {
				localEnvMock.On("Get", "SHELLHUB_SSH_PORT").Return("22").Once()
				localEnvMock.On("Get", "SHELLHUB_VERSION").Return("v1.0.0").Once()
				storeMock.On("SystemGet", gomock.Anything).Return(&models.System{
					Setup: false,
					Authentication: &models.SystemAuthentication{
						Local: &models.SystemAuthenticationLocal{Enabled: false},
					},
				}, nil).Once()
			},
			setupHooks: func() {
				OnGetSystemInfo(func(_ context.Context, _ *responses.SystemInfo) error {
					return errors.New("hook error")
				})
			},
			expectedResult: nil,
			expectedErrMsg: "system info hook failed: hook error",
		},
		{
			description: "port in request overrides host port in api endpoint",
			req:         &requests.GetSystemInfo{Host: "example.com", Port: 8080},
			setupMocks: func() {
				localEnvMock.On("Get", "SHELLHUB_SSH_PORT").Return("22").Once()
				localEnvMock.On("Get", "SHELLHUB_VERSION").Return("v1.0.0").Once()
				storeMock.On("SystemGet", gomock.Anything).Return(&models.System{
					Setup: false,
					Authentication: &models.SystemAuthentication{
						Local: &models.SystemAuthenticationLocal{Enabled: false},
					},
				}, nil).Once()
			},
			setupHooks: func() {},
			expectedResult: &responses.SystemInfo{
				Version: "v1.0.0",
				Setup:   false,
				Endpoints: &responses.SystemEndpointsInfo{
					API: "example.com:8080",
					SSH: "example.com:22",
				},
				Authentication: &responses.SystemAuthenticationInfo{
					Local: false,
					SAML:  false,
				},
			},
		},
		{
			description: "host with port is stripped for ssh endpoint",
			req:         &requests.GetSystemInfo{Host: "example.com:443"},
			setupMocks: func() {
				localEnvMock.On("Get", "SHELLHUB_SSH_PORT").Return("22").Once()
				localEnvMock.On("Get", "SHELLHUB_VERSION").Return("v1.0.0").Once()
				storeMock.On("SystemGet", gomock.Anything).Return(&models.System{
					Setup: false,
					Authentication: &models.SystemAuthentication{
						Local: &models.SystemAuthenticationLocal{Enabled: false},
					},
				}, nil).Once()
			},
			setupHooks: func() {},
			expectedResult: &responses.SystemInfo{
				Version: "v1.0.0",
				Setup:   false,
				Endpoints: &responses.SystemEndpointsInfo{
					API: "example.com:443",
					SSH: "example.com:22",
				},
				Authentication: &responses.SystemAuthenticationInfo{
					Local: false,
					SAML:  false,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			systemInfoHooks = nil
			tc.setupHooks()
			tc.setupMocks()

			result, err := s.GetSystemInfo(ctx, tc.req)
			assert.Equal(t, tc.expectedResult, result)

			if tc.expectedErrMsg != "" {
				assert.EqualError(t, err, tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	storeMock.AssertExpectations(t)
	localEnvMock.AssertExpectations(t)
}
