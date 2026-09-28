package routes

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

const ownerTestTenant = "00000000-0000-4000-0000-000000000000"

func TestMakeTeamDevice(t *testing.T) {
	cases := []struct {
		description string
		role        authorizer.Role
		mocks       func(*mocks.MockService)
		expected    int
	}{
		{
			description: "an operator cannot make a team device",
			role:        authorizer.RoleOperator,
			mocks:       func(*mocks.MockService) {},
			expected:    http.StatusForbidden,
		},
		{
			description: "an administrator makes a team device",
			role:        authorizer.RoleAdministrator,
			mocks: func(m *mocks.MockService) {
				m.On("MakeTeamDevice", gomock.Anything, ownerTestTenant, "uid").Return(nil).Once()
			},
			expected: http.StatusNoContent,
		},
		{
			description: "a device that is not accepted is not found",
			role:        authorizer.RoleOwner,
			mocks: func(m *mocks.MockService) {
				m.On("MakeTeamDevice", gomock.Anything, ownerTestTenant, "uid").Return(svc.NewErrDeviceNotFound(models.UID("uid"), nil)).Once()
			},
			expected: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			m := mocks.NewMockService(t)
			tc.mocks(m)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/devices/uid/owner", nil)
			req.Header.Set("X-Role", tc.role.String())
			req.Header.Set("X-Tenant-ID", ownerTestTenant)
			rec := httptest.NewRecorder()

			NewRouter(m).ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Result().StatusCode)
		})
	}
}

func TestGetDeviceListFiltersByOwner(t *testing.T) {
	encode := func(t *testing.T, value string) string {
		t.Helper()

		b, err := json.Marshal([]query.Filter{{
			Type:   query.FilterTypeProperty,
			Params: &query.FilterProperty{Name: "owner_id", Operator: "eq", Value: value},
		}})
		require.NoError(t, err)

		return base64.StdEncoding.EncodeToString(b)
	}

	list := func(t *testing.T, m *mocks.MockService, filter string) int {
		t.Helper()

		values := url.Values{}
		values.Set("filter", filter)

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/devices?"+values.Encode(), nil)
		req.Header.Set("X-Role", authorizer.RoleOwner.String())
		req.Header.Set("X-Tenant-ID", ownerTestTenant)
		rec := httptest.NewRecorder()

		NewRouter(m).ServeHTTP(rec, req)

		return rec.Result().StatusCode
	}

	t.Run("a user id passes through", func(t *testing.T) {
		m := mocks.NewMockService(t)
		m.On("ListDevices", gomock.Anything, gomock.Anything, gomock.Anything).Return([]models.Device{}, 0, nil).Once()

		assert.Equal(t, http.StatusOK, list(t, m, encode(t, "11111111-1111-4111-8111-111111111111")))
	})

	t.Run("a value that is not a user id is a bad request", func(t *testing.T) {
		m := mocks.NewMockService(t)

		assert.Equal(t, http.StatusBadRequest, list(t, m, encode(t, "not-a-uuid")))
		m.AssertNotCalled(t, "ListDevices")
	})
}

func TestMemberRequestsCarryTheDevicesToKeep(t *testing.T) {
	const memberID = "11111111-1111-4111-8111-111111111111"

	send := func(t *testing.T, m *mocks.MockService, method, query, body string) int {
		t.Helper()

		req := httptest.NewRequestWithContext(t.Context(), method, "/api/namespaces/"+ownerTestTenant+"/members/"+memberID+query, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Role", authorizer.RoleOwner.String())
		req.Header.Set("X-ID", "000000000000000000000000")
		req.Header.Set("X-Tenant-ID", ownerTestTenant)
		rec := httptest.NewRecorder()

		NewRouter(m).ServeHTTP(rec, req)

		return rec.Result().StatusCode
	}

	t.Run("removing a member", func(t *testing.T) {
		m := mocks.NewMockService(t)
		m.On("RemoveNamespaceMember", gomock.Anything, gomock.MatchedBy(func(req *requests.NamespaceRemoveMember) bool {
			return assert.ObjectsAreEqual([]string{"uid-1"}, req.KeepDevices)
		})).Return(&models.Namespace{TenantID: ownerTestTenant}, nil).Once()

		assert.Equal(t, http.StatusOK, send(t, m, http.MethodDelete, "?keep_devices=uid-1", ""))
	})

	t.Run("changing a member's role", func(t *testing.T) {
		m := mocks.NewMockService(t)
		m.On("UpdateNamespaceMember", gomock.Anything, gomock.MatchedBy(func(req *requests.NamespaceUpdateMember) bool {
			return assert.ObjectsAreEqual([]string{"uid-1"}, req.KeepDevices)
		})).Return(nil).Once()

		assert.Equal(t, http.StatusOK, send(t, m, http.MethodPatch, "", `{"role":"observer","keep_devices":["uid-1"]}`))
	})
}
