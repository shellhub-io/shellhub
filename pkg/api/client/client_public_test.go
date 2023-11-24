package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	mock "github.com/jarcoal/httpmock"
	reversermock "github.com/shellhub-io/shellhub/pkg/api/client/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/stretchr/testify/assert"
)

func TestGetInfo(t *testing.T) {
	type Expected struct {
		info *models.Info
		err  error
	}

	tests := []struct {
		description   string
		version       string
		requiredMocks func(client *http.Client)
		expected      Expected
	}{
		{
			description: "success to get info",
			version:     "v0.13.0",
			requiredMocks: func(client *http.Client) {
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
			requiredMocks: func(client *http.Client) {
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
			requiredMocks: func(client *http.Client) {
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
			requiredMocks: func(client *http.Client) {
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
			requiredMocks: func(client *http.Client) {
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
			assert.NoError(t, err)

			client, ok := cli.(*client)
			assert.True(t, ok)

			mock.ActivateNonDefault(client.http.GetClient())
			defer mock.DeactivateAndReset()

			test.requiredMocks(client.http.GetClient())

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
		requiredMocks func(client *http.Client)
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
			requiredMocks: func(client *http.Client) {
				responder, _ := mock.NewJsonResponder(200, models.DeviceAuthResponse{
					UID:       "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
					Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
					Name:      "83-18-77-25-78-0d",
					Namespace: "00000000-0000-4000-0000-000000000000",
				})

				mock.RegisterResponder("POST", "/api/devices/auth", responder)
			},
			expected: Expected{
				response: &models.DeviceAuthResponse{
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
			requiredMocks: func(client *http.Client) {
				fail, _ := mock.NewJsonResponder(404, nil)
				success, _ := mock.NewJsonResponder(200, models.DeviceAuthResponse{
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
				response: &models.DeviceAuthResponse{
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
			assert.NoError(t, err)

			client, ok := cli.(*client)
			assert.True(t, ok)

			mock.ActivateNonDefault(client.http.GetClient())
			defer mock.DeactivateAndReset()

			test.requiredMocks(client.http.GetClient())

			response, err := cli.AuthDevice(test.request)
			assert.Equal(t, test.expected, Expected{response, err})
		})
	}
}

func TestAuthPublicKey(t *testing.T) {
	// NOTICE: It was generated for tests only.
	/*
		-----BEGIN RSA PRIVATE KEY-----
		MIIEpQIBAAKCAQEAmQnNydiZxmr+qwxeibKE3UuSq4SlqrDaYlanynQ6FGRqgL+E
		LWlHDgsXFMeYajcTpYTXl8lcexh6Rld2yU7szZQ6LmBQRpZmQpOI5185/b6oZx1X
		eNpxd9fQzSoRrYRzllvdy3IquH8KUlNPQC7lyM05j/8oITvwibhHAYCJhRhKGMDJ
		msiMeNYnPphLWCaJVvryZKrVMcVK+G7evuMRaysUi+hHaR4wTlqdXnikIslHdsRE
		bD+wHfXu7Z0GKtLr0JOyfSf9oyK8GfdMLpScCnd9uTeuQ/Yw5zjdeuaBo548fIV2
		YYOUcQ4Np15p+qeSdCtoqykwwI7XTEcY3CqnfwIDAQABAoIBAQCJPRx5nS9Rm1Oj
		6qTA0QGLPiFahMrYBl7g/xrbWF1ZOi+0I4sR+i4JLvLU5KUO4zNQJI/+A0pW7T2B
		5R+WKZv8Q6flPOPgWTF9FLcZ3fCwV5DHX624qBRHR5ds1QQjWfsZ5GpC7q/W4b5M
		037nbOcNqtSrCENhglIT/thWRYGNgzEZ89hYPxzDeqrmA9DUoRD5wfnX2JQ7XefX
		OslgNgMuZEAXXmSWP4ViI0l/jX/zSsEekB0eTPGPWPPSu79gQuClYyF96cq04J1v
		elMW2vntUzCeY9FlCg1Dhl/TbUAxAgEyfd5jZhUAwrQtXUPR+ajflFTaqf6+/mO6
		AndsUdLBAoGBAOxk8xWskCxkTtR6jU9OWOJI0T/+/5vFNO6XsJQmE9kMdd+TodUv
		6MaF6LbGp1zCQxha99FOn7voMs7t5qR1ZaIeImAs9VJzimVJ26TKyrtRPEUUKo2K
		0MCC4SV4nyITTRro7JvOxvY9nIawktPvz9vl6xYWXZKBymE0gSnjm5zhAoGBAKW7
		EKWw1KbKh2Qvtomdp98NEdvbdp0QtwtrMvcy9VY9qqIQkxe1q0SM4kjHpAym34EM
		2kE/aKkv76eON5yayh76SQsL3sWEyl+1DfyqDZvG+CQMLNFfYwetn5SqhBiXk8wS
		l794POSP9INwmNotF0VDJWfYwbKv+xOgUQCI6XBfAoGBALvdgj3gVPeig+9isbis
		CAFVY05bpeGyeN2AmakzFaTxvR+AYSVenLxgoU/YUhtqmw3ZBYdG17Tq1K4U7K4Y
		SjvrfOA3+oKbJgpXBwJQ04rIBYDUhPjVsA6+NJWl5bmb4zQWlitAymfZEPMNkrJx
		rBwS/aAhCB4tnZtgsaGLrBihAoGAeHKB4MmpYybiS0/Z7B4meRc4mX7gr7oMLkJU
		og/o+iqLLom5PYwM0x9I/fgKJB/LZDEBFlaNLBjope/YPvrIVscz+tE/sHGX2wLt
		DWpRxZtPkHNAx1H52QS4bAWZR34yLih9HOy35K0y2awlirsmpHHuZ5DDBaicY2Eq
		PadDcSUCgYEAzvA5hDoPYGFxYw75nARwJ53YVrcE3+tQK+sIIQLzS+iqXdq0BbiF
		brF4wr6bHphEFa6WsJF0/3zJ1CsFUu6EsXKsHZaJ77MMuQZ/P/Uz1SynaCwXIdZQ
		tjgb9WIhnkq4PQvEj+84E2q0q14etm6Nl02KZ1R8vweDCmNrNhNkDbY=
		-----END RSA PRIVATE KEY-----

		ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCZCc3J2JnGav6rDF6JsoTdS5KrhKWqsNpiVqfKdDoUZGqAv4QtaUcOCxcUx5hqNxOlhNeXyVx7GHpGV3bJTuzNlDouYFBGlmZCk4jnXzn9vqhnHVd42nF319DNKhGthHOWW93Lciq4fwpSU09ALuXIzTmP/yghO/CJuEcBgImFGEoYwMmayIx41ic+mEtYJolW+vJkqtUxxUr4bt6+4xFrKxSL6EdpHjBOWp1eeKQiyUd2xERsP7Ad9e7tnQYq0uvQk7J9J/2jIrwZ90wulJwKd325N65D9jDnON165oGjnjx8hXZhg5RxDg2nXmn6p5J0K2irKTDAjtdMRxjcKqd/

		cd:8a:1b:73:03:47:15:3c:7c:2b:df:5d:b9:64:63:77*/

	type Signature struct {
		Username  string
		Namespace string
	}

	sig := &Signature{
		Username:  "test",
		Namespace: "namespace",
	}

	sigBytes, err := json.Marshal(sig)
	assert.NoError(t, err)

	sigString := string(sigBytes)
	fmt.Println(sigString)

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
			requiredMocks: func(client *http.Client) {
				responder, _ := mock.NewJsonResponder(401, nil)

				mock.RegisterResponder("POST", "/api/auth/ssh", responder)
			},
			expected: Expected{
				response: nil,
				err:      ErrUnauthorized,
			},
		},
		{
			description: "fail to auth public key when a request field is not set",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			request: &models.PublicKeyAuthRequest{
				Fingerprint: "cd:8a:1b:73:03:47:15:3c:7c:2b:df:5d:b9:64:63:77",
			},
			requiredMocks: func(client *http.Client) {
				responder, _ := mock.NewJsonResponder(400, nil)

				mock.RegisterResponder("POST", "/api/auth/ssh", responder)
			},
			expected: Expected{
				response: nil,
				err:      ErrBadRequest,
			},
		},
		{
			description: "fail to auth public key when the key is not found",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			request: &models.PublicKeyAuthRequest{
				Fingerprint: "cd:8a:1b:73:03:47:15:3c:7c:2b:df:5d:b9:64:63:77",
				Data:        `{"Username":"test","Namespace":"namespace"}`,
			},
			requiredMocks: func(client *http.Client) {
				responder, _ := mock.NewJsonResponder(404, nil)

				mock.RegisterResponder("POST", "/api/auth/ssh", responder)
			},
			expected: Expected{
				response: nil,
				err:      ErrNotFound,
			},
		},
		{
			description: "success to auth public key",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			request: &models.PublicKeyAuthRequest{
				Fingerprint: "cd:8a:1b:73:03:47:15:3c:7c:2b:df:5d:b9:64:63:77",
				Data:        `{"Username":"test","Namespace":"namespace"}`,
			},
			requiredMocks: func(client *http.Client) {
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
			assert.NoError(t, err)

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
	mock := new(reversermock.IReverser)

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
		{
			description: "fail when connot auth the agent on the SSH server",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			requiredMocks: func() {
				mock.On("Auth", context.Background(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c").Return(errors.New("")).Once()
			},
			expected: errors.New(""),
		},
		{
			description: "fail when connot create a new reverse listener",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			requiredMocks: func() {
				mock.On("Auth", context.Background(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c").Return(nil).Once()

				mock.On("NewListener").Return(nil, errors.New("")).Once()
			},
			expected: errors.New(""),
		},
		{
			description: "success to create a new reverse listener",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			requiredMocks: func() {
				mock.On("Auth", context.Background(), "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c").Return(nil).Once()

				mock.On("NewListener").Return(new(revdial.Listener), nil).Once()
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ctx := context.Background()

			cli, err := NewClient("https://www.cloud.shellhub.io/", WithReverser(mock))
			assert.NoError(t, err)

			test.requiredMocks()

			_, err = cli.NewReverseListener(ctx, test.token)
			assert.Equal(t, err, test.expected)
		})
	}
}
