package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/api/services"
	servicemock "github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateInstallKey(t *testing.T) {
	type Expected struct {
		fields map[string]string
		status int
	}

	svcMock := servicemock.NewMockService(t)

	cases := []struct {
		description   string
		headers       map[string]string
		body          map[string]interface{}
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
			body: map[string]interface{}{"usage_limit": 10},
			requiredMocks: func() {
				svcMock.On("UpdateInstallKey", mock.Anything, mock.Anything).
					Return(services.NewErrInstallKeyInvalidField(map[string]string{
						"usage_limit": "cannot be lower than the number of times the key was already used",
					})).
					Once()
			},
			expected: Expected{
				status: http.StatusBadRequest,
				fields: map[string]string{
					"usage_limit": "cannot be lower than the number of times the key was already used",
				},
			},
		},
		{
			description: "answers a bare status for a non-field error",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			body: map[string]interface{}{"revoked": true},
			requiredMocks: func() {
				svcMock.On("UpdateInstallKey", mock.Anything, mock.Anything).
					Return(services.NewErrInstallKeyForbidden()).
					Once()
			},
			expected: Expected{status: http.StatusForbidden},
		},
		{
			description: "succeeds",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			body: map[string]interface{}{"disabled": true},
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

			req := httptest.NewRequest(http.MethodPatch, "/api/namespaces/install-key/ci", strings.NewReader(string(data)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(t, tc.expected.status, rec.Result().StatusCode)

			if tc.expected.fields != nil {
				var body struct {
					Fields map[string]string `json:"fields"`
				}
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
				require.Equal(t, tc.expected.fields, body.Fields)
			}
		})
	}
}
