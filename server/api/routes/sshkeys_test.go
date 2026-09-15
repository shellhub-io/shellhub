package routes

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	svc "github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetPublicKeys(t *testing.T) {
	mock := mocks.NewMockService(t)

	type Expected struct {
		expectedSession []models.PublicKey
		expectedStatus  int
	}
	cases := []struct {
		description   string
		paginator     query.Paginator
		headers       map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "success when try to list a publics keys exists",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			headers:     map[string]string{"X-ID": "000000000000000000000000", "X-Tenant-ID": "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func() {
				mock.
					On("ListPublicKeys", gomock.Anything, &requests.ListPublicKeys{Paginator: query.Paginator{Page: 1, PerPage: 10}, TenantID: "00000000-0000-4000-0000-000000000000"}).
					Return([]models.PublicKey{}, 1, nil).
					Once()
			},
			expected: Expected{
				expectedSession: []models.PublicKey{},
				expectedStatus:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.paginator)
			if err != nil {
				require.NoError(t, err)
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/sshkeys/public-keys", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session []models.PublicKey
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}
}

func TestGetPublicKeysBadFilter(t *testing.T) {
	cases := []struct {
		description string
		filter      string
	}{
		{
			description: "returns 400 when filter is not valid base64",
			filter:      "!!!not-base64!!!",
		},
		{
			description: "returns 400 when filter decodes to invalid JSON",
			filter:      "bm90LWpzb24=", // base64("not-json")
		},
		{
			description: "returns 400 when filter contains an unknown field",
			filter: func() string {
				filters := []query.Filter{
					{
						Type: query.FilterTypeProperty,
						Params: &query.FilterProperty{
							Name:     "nonexistent_field",
							Operator: "eq",
							Value:    "foo",
						},
					},
				}
				b, err := json.Marshal(filters)
				require.NoError(t, err)

				return base64.StdEncoding.EncodeToString(b)
			}(),
		},
		{
			description: "returns 400 when filter uses a disallowed operator for the field",
			filter: func() string {
				filters := []query.Filter{
					{
						Type: query.FilterTypeProperty,
						Params: &query.FilterProperty{
							Name:     "name",
							Operator: "regex", // disallowed: name only allows contains/eq/ne
							Value:    "foo",
						},
					},
				}
				b, err := json.Marshal(filters)
				require.NoError(t, err)

				return base64.StdEncoding.EncodeToString(b)
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := mocks.NewMockService(t)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/sshkeys/public-keys?filter="+tc.filter, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-Tenant-ID", "00000000-0000-4000-0000-000000000000")

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
			svcMock.AssertNotCalled(t, "ListPublicKeys")
		})
	}
}

func TestDeletePublicKey(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := mocks.NewMockService(t)

	cases := []struct {
		description   string
		fingerprint   string
		headers       map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when role is observer",
			fingerprint: "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when role is operator",
			fingerprint: "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when try to deleting a non existent public key",
			fingerprint: "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {
				svcMock.
					On("DeletePublicKey", gomock.Anything, "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8", "00000000-0000-4000-0000-000000000000").
					Return(svc.ErrNotFound).
					Once()
			},
			expected: Expected{status: http.StatusNotFound},
		},
		{
			description: "success when fingerprint is encoded",
			fingerprint: "8e%3Ab3%3Ae2%3Ace%3A3c%3A6c%3A27%3Aff%3A51%3Ac9%3A5d%3A77%3Aaf%3A92%3A2f%3Ad8",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {
				svcMock.
					On("DeletePublicKey", gomock.Anything, "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8", "00000000-0000-4000-0000-000000000000").
					Return(nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
		{
			description: "success when try to deleting an existing public key",
			fingerprint: "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {
				svcMock.
					On("DeletePublicKey", gomock.Anything, "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8", "00000000-0000-4000-0000-000000000000").
					Return(nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/sshkeys/public-keys/"+tc.fingerprint, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.status, rec.Result().StatusCode)
		})
	}
}
