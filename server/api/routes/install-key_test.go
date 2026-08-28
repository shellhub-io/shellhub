package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/server/api/services"
	servicemock "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateInstallKey(t *testing.T) {
	type Expected struct {
		body   *responses.Error
		status int
	}

	svcMock := servicemock.NewMockService(t)

	cases := []struct {
		description   string
		headers       map[string]string
		body          map[string]any
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "surfaces the offending field on an invalid update",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			body: map[string]any{"usage_limit": 10},
			requiredMocks: func() {
				svcMock.On("UpdateInstallKey", mock.Anything, mock.Anything).
					Return(services.NewErrInstallKeyInvalidField(map[string]string{
						"usage_limit": "cannot be lower than the number of times the key was already used",
					})).
					Once()
			},
			expected: Expected{status: http.StatusBadRequest},
		},
		{
			description: "answers a message-only body for a non-field error",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			body: map[string]any{"revoked": true},
			requiredMocks: func() {
				svcMock.On("UpdateInstallKey", mock.Anything, mock.Anything).
					Return(services.NewErrInstallKeyForbidden()).
					Once()
			},
			expected: Expected{
				status: http.StatusForbidden,
				body:   &responses.Error{Message: "the legacy install key cannot be modified"},
			},
		},
		{
			description: "succeeds",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			body: map[string]any{"disabled": true},
			requiredMocks: func() {
				svcMock.On("UpdateInstallKey", mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/api/namespaces/install-key/ci", strings.NewReader(string(data)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(t, tc.expected.status, rec.Result().StatusCode)

			if tc.expected.body != nil {
				body := responses.Error{} //nolint:exhaustruct
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
				require.Equal(t, *tc.expected.body, body)
			}
		})
	}
}
