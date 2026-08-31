package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	mock "github.com/jarcoal/httpmock"
	reversermock "github.com/shellhub-io/shellhub/pkg/api/client/reverser/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInfo(t *testing.T) {
	type Expected struct {
		info *models.Info
		err  error
	}

	tests := []struct {
		description   string
		version       string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "success to get info",
			version:     "v0.13.0",
			requiredMocks: func() {
				responder, _ := mock.NewJsonResponder(200, models.Info{
					Version: "v0.13.0",
					Endpoints: models.Endpoints{
						API: "cloud.shellhub.io:443",
						SSH: "cloud.shellhub.io:2222",
					},
				})

				mock.RegisterResponder("GET", "/info?agent_version=v0.13.0", responder)
			},
			expected: Expected{
				info: &models.Info{
					Version: "v0.13.0",
					Endpoints: models.Endpoints{
						API: "cloud.shellhub.io:443",
						SSH: "cloud.shellhub.io:2222",
					},
				},
				err: nil,
			},
		},
		{
			description: "success to get info after retry",
			version:     "v0.13.0",
			requiredMocks: func() {
				fail := mock.NewErrorResponder(errors.New("error on request"))
				success, _ := mock.NewJsonResponder(200, models.Info{
					Version: "v0.13.0",
					Endpoints: models.Endpoints{
						API: "cloud.shellhub.io:443",
						SSH: "cloud.shellhub.io:2222",
					},
				})

				responder := fail.
					Then(fail).
					Then(fail).
					Then(success)

				mock.RegisterResponder("GET", "/info?agent_version=v0.13.0", responder)
			},
			expected: Expected{
				info: &models.Info{
					Version: "v0.13.0",
					Endpoints: models.Endpoints{
						API: "cloud.shellhub.io:443",
						SSH: "cloud.shellhub.io:2222",
					},
				},
				err: nil,
			},
		},
		{
			description: "failed when resource is not found",
			version:     "v0.13.0",
			requiredMocks: func() {
				responder, _ := mock.NewJsonResponder(404, nil)

				mock.RegisterResponder("GET", "/info?agent_version=v0.13.0", responder)
			},
			expected: Expected{
				info: nil,
				err:  ErrNotFound,
			},
		},
		{
			description: "failed when request is missformated",
			version:     "v0.13.0",
			requiredMocks: func() {
				responder, _ := mock.NewJsonResponder(400, nil)

				mock.RegisterResponder("GET", "/info?agent_version=v0.13.0", responder)
			},
			expected: Expected{
				info: nil,
				err:  ErrBadRequest,
			},
		},
		{
			description: "failed when device request return an unmaped error code",
			version:     "v0.13.0",
			requiredMocks: func() {
				responder, _ := mock.NewJsonResponder(418, nil)

				mock.RegisterResponder("GET", "/info?agent_version=v0.13.0", responder)
			},
			expected: Expected{
				info: nil,
				err:  errors.Join(ErrUnknown, errors.New("418")),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cli, err := NewClient("https://www.cloud.shellhub.io/")
			require.NoError(t, err)

			client, ok := cli.(*client)
			assert.True(t, ok)

			mock.ActivateNonDefault(client.http.GetClient())
			defer mock.DeactivateAndReset()

			test.requiredMocks()

			info, err := cli.GetInfo(test.version)
			assert.Equal(t, test.expected, Expected{info, err})
		})
	}
}

func TestAuthDevice(t *testing.T) {
	type Expected struct {
		response *models.DeviceAuthResponse
		err      error
	}

	tests := []struct {
		description   string
		request       *models.DeviceAuthRequest
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "success to auth device",
			request: &models.DeviceAuthRequest{
				Info: &models.DeviceInfo{
					ID:         "manjaro",
					PrettyName: "Manjaro",
					Version:    "latest",
					Arch:       "amd64",
					Platform:   "docker",
				},
				DeviceAuth: &models.DeviceAuth{
					Hostname: "83-18-77-25-78-0d",
					Identity: &models.DeviceIdentity{
						MAC: "83:18:77:25:78:0d",
					},
					TenantID:  "00000000-0000-4000-0000-000000000000",
					PublicKey: "",
				},
			},
			requiredMocks: func() {
				responder, _ := mock.NewJsonResponder(200, models.DeviceAuthResponse{ //nolint:gosec // G101: test fixture
					UID:       "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
					Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
					Name:      "83-18-77-25-78-0d",
					Namespace: "00000000-0000-4000-0000-000000000000",
				})

				mock.RegisterResponder("POST", "/api/devices/auth", responder)
			},
			expected: Expected{
				response: &models.DeviceAuthResponse{ //nolint:gosec // G101: test fixture
					UID:       "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
					Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
					Name:      "83-18-77-25-78-0d",
					Namespace: "00000000-0000-4000-0000-000000000000",
				},
				err: nil,
			},
		},
		{
			description: "success to auth after any error",
			request: &models.DeviceAuthRequest{
				Info: &models.DeviceInfo{
					ID:         "manjaro",
					PrettyName: "Manjaro",
					Version:    "latest",
					Arch:       "amd64",
					Platform:   "docker",
				},
				DeviceAuth: &models.DeviceAuth{
					Hostname: "83-18-77-25-78-0d",
					Identity: &models.DeviceIdentity{
						MAC: "83:18:77:25:78:0d",
					},
					TenantID:  "00000000-0000-4000-0000-000000000000",
					PublicKey: "",
				},
			},
			requiredMocks: func() {
				fail, _ := mock.NewJsonResponder(404, nil)
				success, _ := mock.NewJsonResponder(200, models.DeviceAuthResponse{ //nolint:gosec // G101: test fixture
					UID:       "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
					Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
					Name:      "83-18-77-25-78-0d",
					Namespace: "00000000-0000-4000-0000-000000000000",
				})

				responder := fail.
					Then(fail).
					Then(fail).Then(success)

				mock.RegisterResponder("POST", "/api/devices/auth", responder)
			},
			expected: Expected{
				response: &models.DeviceAuthResponse{ //nolint:gosec // G101: test fixture
					UID:       "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
					Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
					Name:      "83-18-77-25-78-0d",
					Namespace: "00000000-0000-4000-0000-000000000000",
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cli, err := NewClient("https://www.cloud.shellhub.io/")
			require.NoError(t, err)

			client, ok := cli.(*client)
			assert.True(t, ok)

			mock.ActivateNonDefault(client.http.GetClient())
			defer mock.DeactivateAndReset()

			test.requiredMocks()

			response, err := cli.AuthDevice(test.request)
			assert.Equal(t, test.expected, Expected{response, err})
		})
	}
}

func TestAuthPublicKey(t *testing.T) {
	type Signature struct {
		Username  string `json:"Username"`
		Namespace string `json:"Namespace"`
	}

	sig := &Signature{
		Username:  "test",
		Namespace: "namespace",
	}

	sigBytes, err := json.Marshal(sig)
	require.NoError(t, err)

	sigString := string(sigBytes)
	t.Log(sigString)

	type Expected struct {
		response *models.PublicKeyAuthResponse
		err      error
	}

	tests := []struct {
		description   string
		request       *models.PublicKeyAuthRequest
		token         string
		requiredMocks func(client *http.Client)
		expected      Expected
	}{
		{
			description: "fail to auth public key when token is empty",
			token:       "",
			request: &models.PublicKeyAuthRequest{
				Fingerprint: "cd:8a:1b:73:03:47:15:3c:7c:2b:df:5d:b9:64:63:77",
				Data:        `{"Username":"test","Namespace":"namespace"}`,
			},
			requiredMocks: func(_ *http.Client) {
				responder, _ := mock.NewJsonResponder(401, nil)

				mock.RegisterResponder("POST", "/api/auth/ssh", responder)
			},
			expected: Expected{
				response: nil,
				err:      ErrUnauthorized,
			},
		},
		{ //nolint:gosec // G101: test fixture
			description: "fail to auth public key when a request field is not set",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			request: &models.PublicKeyAuthRequest{
				Fingerprint: "cd:8a:1b:73:03:47:15:3c:7c:2b:df:5d:b9:64:63:77",
			},
			requiredMocks: func(_ *http.Client) {
				responder, _ := mock.NewJsonResponder(400, nil)

				mock.RegisterResponder("POST", "/api/auth/ssh", responder)
			},
			expected: Expected{
				response: nil,
				err:      ErrBadRequest,
			},
		},
		{ //nolint:gosec // G101: test fixture
			description: "fail to auth public key when the key is not found",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			request: &models.PublicKeyAuthRequest{
				Fingerprint: "cd:8a:1b:73:03:47:15:3c:7c:2b:df:5d:b9:64:63:77",
				Data:        `{"Username":"test","Namespace":"namespace"}`,
			},
			requiredMocks: func(_ *http.Client) {
				responder, _ := mock.NewJsonResponder(404, nil)

				mock.RegisterResponder("POST", "/api/auth/ssh", responder)
			},
			expected: Expected{
				response: nil,
				err:      ErrNotFound,
			},
		},
		{ //nolint:gosec // G101: test fixture
			description: "success to auth public key",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			request: &models.PublicKeyAuthRequest{
				Fingerprint: "cd:8a:1b:73:03:47:15:3c:7c:2b:df:5d:b9:64:63:77",
				Data:        `{"Username":"test","Namespace":"namespace"}`,
			},
			requiredMocks: func(_ *http.Client) {
				responder, _ := mock.NewJsonResponder(200, models.PublicKeyAuthResponse{
					Signature: "hgihFKUWAr1QB87Eb7FiBu9pMjTLLBgLXEqNIYd4S+UoOZ7xqozEMds9EvwB1TwCjMa+uAmZsB7qtARVvoVPrUNp/OBQ7iKzV2+GpIpRFfEqa0ugQBf+XQBfo/irDnH/wAixgoqC3KUyIk+nQxwz7wvgVDB0WTxD2eK9TzyD3WIOSVGlPWNytx7HTP0TTN5EJ0tjj/H4v1F9t+8Nd3ZGUz0z73rZ3qKbzXBJBkRyyDDtWo9lGIOcz5e4LcgojNVxznsHDXJ/2gBnRL6JHZQm6v3gCpzZRRXA+cagSSuJzWQwwDmwydfiAJsbSPeen4+X+IEkfrXBW1KHMRsZh1AtTw==",
				})

				mock.RegisterResponder("POST", "/api/auth/ssh", responder)
			},
			expected: Expected{
				response: &models.PublicKeyAuthResponse{
					Signature: "hgihFKUWAr1QB87Eb7FiBu9pMjTLLBgLXEqNIYd4S+UoOZ7xqozEMds9EvwB1TwCjMa+uAmZsB7qtARVvoVPrUNp/OBQ7iKzV2+GpIpRFfEqa0ugQBf+XQBfo/irDnH/wAixgoqC3KUyIk+nQxwz7wvgVDB0WTxD2eK9TzyD3WIOSVGlPWNytx7HTP0TTN5EJ0tjj/H4v1F9t+8Nd3ZGUz0z73rZ3qKbzXBJBkRyyDDtWo9lGIOcz5e4LcgojNVxznsHDXJ/2gBnRL6JHZQm6v3gCpzZRRXA+cagSSuJzWQwwDmwydfiAJsbSPeen4+X+IEkfrXBW1KHMRsZh1AtTw==",
				},
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cli, err := NewClient("https://www.cloud.shellhub.io/")
			require.NoError(t, err)

			client, ok := cli.(*client)
			assert.True(t, ok)

			mock.ActivateNonDefault(client.http.GetClient())
			defer mock.DeactivateAndReset()

			test.requiredMocks(client.http.GetClient())

			response, err := cli.AuthPublicKey(test.request, test.token)
			assert.Equal(t, test.expected, Expected{response, err})
		})
	}
}

func TestReverseListener(t *testing.T) {
	mock := reversermock.NewMockReverser(t)

	tests := []struct {
		description   string
		token         string
		requiredMocks func()
		expected      error
	}{
		{
			description:   "fail when token is empty",
			token:         "",
			requiredMocks: func() {},
			expected:      errors.New("token is empty"),
		},
		{ //nolint:gosec // G101: well-known jwt.io example token, not a real credential
			description: "fail when cannot auth the agent on the SSH server",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			requiredMocks: func() {
				mock.On("Auth", context.Background(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", "").Return(errors.New("")).Once()
			},
			expected: errors.New(""),
		},
		{ //nolint:gosec // G101: well-known jwt.io example token, not a real credential
			description: "fail when cannot create a new reverse listener",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			requiredMocks: func() {
				mock.On("Auth", context.Background(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", "").Return(nil).Once()

				mock.On("NewListener").Return(nil, errors.New("")).Once()
			},
			expected: errors.New(""),
		},
		{ //nolint:gosec // G101: well-known jwt.io example token, not a real credential
			description: "success to create a new reverse listener",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			requiredMocks: func() {
				mock.On("Auth", context.Background(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", "").Return(nil).Once()

				mock.On("NewListener").Return(new(revdial.Listener), nil).Once()
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ctx := context.Background()

			cli, err := NewClient("https://www.cloud.shellhub.io/", WithReverser(mock))
			require.NoError(t, err)

			test.requiredMocks()

			_, err = cli.NewReverseListenerV1(ctx, test.token, "")
			assert.Equal(t, test.expected, err)
		})
	}
}
