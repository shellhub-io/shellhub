package routes

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	serviceMocks "github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetup(t *testing.T) {
	envMock := new(envMocks.Backend)
	envs.DefaultBackend = envMock

	envMock.On("Get", "SHELLHUB_CLOUD").Return("false")

	servicesMock := new(serviceMocks.Service)

	tests := []struct {
		description   string
		body          string
		requiredMocks func()
		expected      int
	}{
		{
			description:   "fail to parse the json body",
			body:          "",
			requiredMocks: func() {},
			expected:      http.StatusBadRequest,
		},
		{
			description: "fail to valid the json body",
			body: `{
                "name": "John Doe",
                "username": "john.doe",
                "email": "john.doe",
                "password": "123"
            }`,
			requiredMocks: func() {},
			expected:      http.StatusBadRequest,
		},
		{
			description: "fail to setup on service",
			body: `{
                "name": "John Doe",
                "username": "john.doe",
                "email": "john.doe@example.com",
                "password": "password"
            }`,
			requiredMocks: func() {
				servicesMock.On("Setup", mock.Anything, requests.Setup{
					Name:     "John Doe",
					Username: "john.doe",
					Email:    "john.doe@example.com",
					Password: "password",
				}).Return(errors.New("")).Once()
			},
			expected: http.StatusInternalServerError,
		},
		{
			description: "success to setup on service",
			body: `{
                "name": "John Doe",
                "username": "john.doe",
                "email": "john.doe@example.com",
                "password": "password"
            }`,
			requiredMocks: func() {
				servicesMock.On("Setup", mock.Anything, requests.Setup{
					Name:     "John Doe",
					Username: "john.doe",
					Email:    "john.doe@example.com",
					Password: "password",
				}).Return(nil).Once()
			},
			expected: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()

			req := httptest.NewRequest(http.MethodPost, "/api/setup", strings.NewReader(test.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router := NewRouter(servicesMock)
			router.ServeHTTP(rec, req)

			result := rec.Result()

			assert.Equal(t, test.expected, result.StatusCode)
		})
	}

	envMock.AssertExpectations(t)
}
