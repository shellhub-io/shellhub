package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func apiKeyRequestTo(e *echo.Echo, path, key string) *echo.Context {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
	req.Header.Set("X-API-Key", key)
	c := e.NewContext(req, httptest.NewRecorder())
	c.SetPath(path)

	return c
}

func apiKeyRequest(e *echo.Echo, key string) *echo.Context {
	return apiKeyRequestTo(e, "/admin/api/users", key)
}

func TestAuthenticatorResolveInstanceAPIKey(t *testing.T) {
	const instanceKey = models.InstanceAPIKeyPrefix + "cdfd3cb0-c44e-4e54-b931-6d57713ad159"

	cases := []struct {
		description   string
		key           string
		requiredMocks func(service *mocks.MockService)
		expected      *gateway.Identity
	}{
		{
			description: "resolves a prefixed key to the instance administrator identity",
			key:         instanceKey,
			requiredMocks: func(service *mocks.MockService) {
				service.On("AuthInstanceAPIKey", mock.Anything, instanceKey).
					Return(&models.InstanceAPIKey{ID: "digest", Name: "license-sync", CreatedBy: testUserID}, nil).
					Once()
			},
			expected: &gateway.Identity{
				APIKey: instanceKey,
				Admin:  true,
			},
		},
		{
			description: "yields no identity when the creator is no longer an instance administrator",
			key:         instanceKey,
			requiredMocks: func(service *mocks.MockService) {
				service.On("AuthInstanceAPIKey", mock.Anything, instanceKey).
					Return(nil, errors.New("unauthorized")).
					Once()
			},
			expected: nil,
		},
		{
			description: "yields no identity when the prefixed key is unknown",
			key:         models.InstanceAPIKeyPrefix + "00000000-0000-4000-0000-000000000000",
			requiredMocks: func(service *mocks.MockService) {
				service.On("AuthInstanceAPIKey", mock.Anything, models.InstanceAPIKeyPrefix+"00000000-0000-4000-0000-000000000000").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: nil,
		},
		{
			description: "yields no identity when the prefix carries no key at all",
			key:         models.InstanceAPIKeyPrefix,
			requiredMocks: func(service *mocks.MockService) {
				service.On("AuthInstanceAPIKey", mock.Anything, models.InstanceAPIKeyPrefix).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			service := new(mocks.MockService)
			tc.requiredMocks(service)

			identity, err := NewAuthenticator(service).Resolve(apiKeyRequest(echo.New(), tc.key))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, identity)

			service.AssertExpectations(t)
		})
	}
}

func TestAuthenticatorResolveInstanceAPIKeyOffAdminSurface(t *testing.T) {
	const instanceKey = models.InstanceAPIKeyPrefix + "cdfd3cb0-c44e-4e54-b931-6d57713ad159"

	for _, path := range []string{
		"/api/devices",
		"/api/namespaces/:tenant",
		"/api/users/security",
	} {
		t.Run(path, func(t *testing.T) {
			service := new(mocks.MockService)

			identity, err := NewAuthenticator(service).Resolve(
				apiKeyRequestTo(echo.New(), path, instanceKey),
			)
			require.NoError(t, err)
			assert.Nil(t, identity)

			service.AssertNotCalled(t, "AuthInstanceAPIKey", mock.Anything, mock.Anything)
			service.AssertExpectations(t)
		})
	}
}

func TestAuthenticatorResolveNamespaceAPIKeyIsNeverAdmin(t *testing.T) {
	const namespaceKey = "cdfd3cb0-c44e-4e54-b931-6d57713ad159"

	for _, role := range []authorizer.Role{
		authorizer.RoleOwner,
		authorizer.RoleAdministrator,
		authorizer.RoleOperator,
		authorizer.RoleObserver,
	} {
		t.Run("role "+role.String(), func(t *testing.T) {
			service := new(mocks.MockService)
			service.On("AuthAPIKey", mock.Anything, namespaceKey).
				Return(&models.APIKey{ID: "digest", Name: "ci", TenantID: testTenant, Role: role}, nil).
				Once()

			identity, err := NewAuthenticator(service).Resolve(apiKeyRequest(echo.New(), namespaceKey))
			require.NoError(t, err)
			require.NotNil(t, identity)

			assert.False(t, identity.Admin, "a namespace API key must never carry the admin flag")
			assert.Equal(t, testTenant, identity.TenantID)
			assert.Equal(t, role, identity.Role)

			service.AssertNotCalled(t, "AuthInstanceAPIKey", mock.Anything, mock.Anything)
			service.AssertExpectations(t)
		})
	}
}
