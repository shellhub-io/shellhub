package routes

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/identity/identitytest"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const locationsTestTenant = "00000000-0000-4000-0000-000000000000"

func encodeLocationFilter(t *testing.T, filters []query.Filter) string {
	t.Helper()

	b, err := json.Marshal(filters)
	require.NoError(t, err)

	return base64.RawURLEncoding.EncodeToString(b)
}

func newLocationsRequest(t *testing.T, filter string, tenant string) *http.Request {
	t.Helper()

	target := "/api/devices/locations"
	if filter != "" {
		target += "?" + url.Values{"filter": {filter}}.Encode()
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)

	return identitytest.As(req, identity.Identity{
		ID:       "000000000000000000000000",
		Role:     authorizer.RoleOwner,
		TenantID: tenant,
	})
}

func TestGetDeviceLocations(t *testing.T) {
	t.Run("serves the locations the service counts in the request's namespace", func(t *testing.T) {
		mock := mocks.NewMockService(t)
		want := []models.DeviceLocation{
			{Latitude: 47.38, Longitude: 8.54, Devices: 128, Online: 97},
			{Latitude: 52.52, Longitude: 13.40, Devices: 3, Online: 0},
		}

		filter := encodeLocationFilter(t, []query.Filter{{
			Type:   query.FilterTypeProperty,
			Params: &query.FilterProperty{Name: "latitude", Operator: "gt", Value: "47"},
		}})

		mock.
			On("ListDeviceLocations",
				gomock.Anything,
				gomock.MatchedBy(func(sc scope.Scope) bool { return sc.IsBounded() && sc.TenantID() == locationsTestTenant }),
				gomock.MatchedBy(func(req *requests.DeviceLocationList) bool {
					if len(req.Filters.Data) != 1 {
						return false
					}

					property, ok := req.Filters.Data[0].Params.(*query.FilterProperty)

					return ok && property.Name == "latitude" && property.Operator == "gt"
				}),
			).
			Return(want, nil).
			Once()

		rec := httptest.NewRecorder()
		newTestRouter(t, mock).ServeHTTP(rec, newLocationsRequest(t, filter, locationsTestTenant))

		require.Equal(t, http.StatusOK, rec.Result().StatusCode)

		var got []models.DeviceLocation
		require.NoError(t, json.NewDecoder(rec.Result().Body).Decode(&got))
		assert.Equal(t, want, got)
	})

	t.Run("refuses a request with no namespace", func(t *testing.T) {
		mock := mocks.NewMockService(t)

		rec := httptest.NewRecorder()
		newTestRouter(t, mock).ServeHTTP(rec, newLocationsRequest(t, "", ""))

		assert.Equal(t, http.StatusForbidden, rec.Result().StatusCode)
		mock.AssertNotCalled(t, "ListDeviceLocations")
	})

	t.Run("answers 500 when the service fails", func(t *testing.T) {
		mock := mocks.NewMockService(t)
		mock.
			On("ListDeviceLocations", gomock.Anything, gomock.Anything, gomock.Anything).
			Return(nil, errors.New("store failure")).
			Once()

		rec := httptest.NewRecorder()
		newTestRouter(t, mock).ServeHTTP(rec, newLocationsRequest(t, "", locationsTestTenant))

		assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
	})
}

func TestGetDeviceLocationsBadFilter(t *testing.T) {
	cases := []struct {
		description string
		filter      func(t *testing.T) string
	}{
		{
			description: "a filter that is not base64",
			filter:      func(*testing.T) string { return "!!!not-base64!!!" },
		},
		{
			description: "a field the device list does not filter on",
			filter: func(t *testing.T) string {
				t.Helper()

				return encodeLocationFilter(t, []query.Filter{{
					Type:   query.FilterTypeProperty,
					Params: &query.FilterProperty{Name: "nonexistent_field", Operator: "eq", Value: "foo"},
				}})
			},
		},
		{
			description: "an operator latitude does not take",
			filter: func(t *testing.T) string {
				t.Helper()

				return encodeLocationFilter(t, []query.Filter{{
					Type:   query.FilterTypeProperty,
					Params: &query.FilterProperty{Name: "latitude", Operator: "contains", Value: "47"},
				}})
			},
		},
		{
			description: "a latitude that is not a number",
			filter: func(t *testing.T) string {
				t.Helper()

				return encodeLocationFilter(t, []query.Filter{{
					Type:   query.FilterTypeProperty,
					Params: &query.FilterProperty{Name: "latitude", Operator: "gt", Value: "north"},
				}})
			},
		},
		{
			description: "a located that is not a bool",
			filter: func(t *testing.T) string {
				t.Helper()

				return encodeLocationFilter(t, []query.Filter{{
					Type:   query.FilterTypeProperty,
					Params: &query.FilterProperty{Name: "located", Operator: "eq", Value: "maybe"},
				}})
			},
		},
	}

	for _, tc := range cases {
		t.Run("answers 400 to "+tc.description, func(t *testing.T) {
			mock := mocks.NewMockService(t)

			rec := httptest.NewRecorder()
			newTestRouter(t, mock).ServeHTTP(rec, newLocationsRequest(t, tc.filter(t), locationsTestTenant))

			assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
			mock.AssertNotCalled(t, "ListDeviceLocations")
		})
	}
}

func TestGetDeviceLocationsTakesLocatedAsABoolUnderEq(t *testing.T) {
	mock := mocks.NewMockService(t)
	mock.
		On("ListDeviceLocations", gomock.Anything, gomock.Anything, gomock.Anything).
		Return([]models.DeviceLocation{}, nil).
		Once()

	filter := encodeLocationFilter(t, []query.Filter{{
		Type:   query.FilterTypeProperty,
		Params: &query.FilterProperty{Name: "located", Operator: "eq", Value: true},
	}})

	rec := httptest.NewRecorder()
	newTestRouter(t, mock).ServeHTTP(rec, newLocationsRequest(t, filter, locationsTestTenant))

	assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
}

func TestGetDeviceLocationsReportsTheGeoIPSource(t *testing.T) {
	cases := []struct {
		description string
		options     []Option
		header      string
	}{
		{"names DB-IP when the server locates with it", []Option{WithGeoIPSource("dbip")}, "dbip"},
		{"names MaxMind when the server locates with it", []Option{WithGeoIPSource("maxmind")}, "maxmind"},
		{"names no source when the server opened no database", []Option{WithGeoIPSource("")}, ""},
		{"names no source when the server was told none", nil, ""},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			mock := mocks.NewMockService(t)
			mock.
				On("ListDeviceLocations", gomock.Anything, gomock.Anything, gomock.Anything).
				Return([]models.DeviceLocation{}, nil).
				Once()

			rec := httptest.NewRecorder()
			newTestRouter(t, mock, tc.options...).ServeHTTP(rec, newLocationsRequest(t, "", locationsTestTenant))

			require.Equal(t, http.StatusOK, rec.Result().StatusCode)
			values, present := rec.Result().Header["X-Geoip-Source"]
			if tc.header == "" {
				assert.False(t, present, "header %v", values)

				return
			}

			assert.Equal(t, tc.header, rec.Result().Header.Get("X-GeoIP-Source"))
		})
	}
}
