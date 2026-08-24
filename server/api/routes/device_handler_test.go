package routes

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	errs "github.com/shellhub-io/shellhub/server/api/routes/errors"
	svc "github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestGetDeviceHandler drives the handler as the pure function it is: no HTTP server, no status
// code, just the inputs the route table resolves and the service behind a mock.
func TestGetDeviceHandler(t *testing.T) {
	cases := []struct {
		description    string
		sc             scope.Scope
		uid            string
		requiredMocks  func(*mocks.MockService)
		expectedDevice *models.Device
		expectedErr    error
	}{
		{
			description: "passes the namespace scope it was given through to the service",
			sc:          scope.MustBounded("00000000-0000-4000-0000-000000000000"),
			uid:         "uid",
			requiredMocks: func(service *mocks.MockService) {
				service.
					On("GetDevice", gomock.Anything, scope.MustBounded("00000000-0000-4000-0000-000000000000"), models.UID("uid")).
					Return(&models.Device{UID: "uid"}, nil).
					Once()
			},
			expectedDevice: &models.Device{UID: "uid"},
		},
		{
			description: "reports the service's failure unchanged",
			sc:          scope.NewUnbounded("the admin console reads every namespace"),
			uid:         "missing",
			requiredMocks: func(service *mocks.MockService) {
				service.
					On("GetDevice", gomock.Anything, scope.NewUnbounded("the admin console reads every namespace"), models.UID("missing")).
					Return(nil, svc.ErrDeviceNotFound).
					Once()
			},
			expectedErr: svc.ErrDeviceNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			service := mocks.NewMockService(t)
			tc.requiredMocks(service)

			handler := NewHandler(service, nil)

			device, err := handler.GetDevice(t.Context(), tc.sc, gateway.Actor{ID: "user-id"},
				&requests.DeviceGet{DeviceParam: requests.DeviceParam{UID: tc.uid}})

			require.Equal(t, tc.expectedErr, err)
			require.Equal(t, tc.expectedDevice, device)
		})
	}
}

// TestGetDeviceListHandler covers what the device list handler still decides now that the wrapper
// owns the ceremony: the sort field, the caller's encoded filter, and nothing else.
func TestGetDeviceListHandler(t *testing.T) {
	const tenantID = "00000000-0000-4000-0000-000000000000"

	encode := func(t *testing.T, filters []query.Filter) string {
		t.Helper()

		raw, err := json.Marshal(filters)
		require.NoError(t, err)

		return base64.StdEncoding.EncodeToString(raw)
	}

	cases := []struct {
		description    string
		req            func(*testing.T) *requests.DeviceList
		requiredMocks  func(*mocks.MockService)
		expectedFields map[string]string
	}{
		{
			description: "refuses a sort field the device list does not accept",
			req: func(*testing.T) *requests.DeviceList {
				return &requests.DeviceList{Sorter: query.Sorter{By: "not_a_column", Order: query.OrderAsc}}
			},
			requiredMocks:  func(*mocks.MockService) {},
			expectedFields: map[string]string{"sort_by": "not_a_column"},
		},
		{
			description: "refuses a filter that is not valid base64",
			req: func(*testing.T) *requests.DeviceList {
				return &requests.DeviceList{Filters: query.Filters{Raw: "!!!not-base64!!!"}}
			},
			requiredMocks:  func(*mocks.MockService) {},
			expectedFields: map[string]string{"filter": "cannot be decoded"},
		},
		{
			description: "refuses a filter naming a field the device list does not know",
			req: func(t *testing.T) *requests.DeviceList {
				t.Helper()

				raw := encode(t, []query.Filter{{
					Type:   query.FilterTypeProperty,
					Params: &query.FilterProperty{Name: "nonexistent_field", Operator: "eq", Value: "foo"},
				}})

				return &requests.DeviceList{Filters: query.Filters{Raw: raw}}
			},
			requiredMocks:  func(*mocks.MockService) {},
			expectedFields: map[string]string{"filter": "is not valid"},
		},
		{
			description: "hands the service the decoded filter and the caller's connector intent",
			req: func(t *testing.T) *requests.DeviceList {
				t.Helper()

				raw := encode(t, []query.Filter{{
					Type:   query.FilterTypeProperty,
					Params: &query.FilterProperty{Name: "name", Operator: "contains", Value: "foo"},
				}})

				return &requests.DeviceList{TenantID: tenantID, Connector: true, Filters: query.Filters{Raw: raw}}
			},
			requiredMocks: func(service *mocks.MockService) {
				service.
					On("ListDevices", gomock.Anything, scope.MustBounded(tenantID), gomock.MatchedBy(func(req *requests.DeviceList) bool {
						if !req.Connector || len(req.Filters.Data) != 1 {
							return false
						}

						property, ok := req.Filters.Data[0].Params.(*query.FilterProperty)

						return ok && property.Name == "name" && property.Value == "foo"
					})).
					Return([]models.Device{}, 0, nil).
					Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			service := mocks.NewMockService(t)
			tc.requiredMocks(service)

			handler := NewHandler(service, nil)

			_, _, err := handler.GetDeviceList(t.Context(), scope.MustBounded(tenantID), gateway.Actor{ID: "user-id"}, tc.req(t))

			if tc.expectedFields != nil {
				require.Error(t, err)

				var wrapped errors.Error
				require.ErrorAs(t, err, &wrapped, "a refusal must be a ShellHub error")

				data, ok := wrapped.Data.(errs.ErrDataInvalidEntity)
				require.True(t, ok, "a refusal must carry the fields the caller can act on, got %v", wrapped.Data)
				assert.Equal(t, tc.expectedFields, data.Fields)

				return
			}

			require.NoError(t, err)
		})
	}
}
