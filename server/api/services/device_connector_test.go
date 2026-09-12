package services

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func connectorFilter(operator string) query.Filter {
	return query.Filter{
		Type:   query.FilterTypeProperty,
		Params: &query.FilterProperty{Name: "platform", Operator: operator, Value: "connector"},
	}
}

func andFilter() query.Filter {
	return query.Filter{
		Type:   query.FilterTypeOperator,
		Params: &query.FilterOperator{Name: "and"},
	}
}

func withoutConnectors(filters ...query.Filter) *query.Filters {
	return &query.Filters{Data: append(filters, andFilter(), connectorFilter("ne"))}
}

// TestListDevicesConnectorIntent asserts what the caller's intent means, not how it is spelled: the
// route test it replaces only checked that an operator preceded a property, and never which of
// them included or excluded connector devices.
func TestListDevicesConnectorIntent(t *testing.T) {
	const tenantID = "00000000-0000-4000-0000-000000000000"

	userFilter := query.Filter{
		Type:   query.FilterTypeProperty,
		Params: &query.FilterProperty{Name: "name", Operator: "contains", Value: "foo"},
	}

	cases := []struct {
		description     string
		connector       bool
		filters         []query.Filter
		expectedFilters []query.Filter
		expectedErr     error
	}{
		{
			description:     "excludes connector devices when the caller asked for none",
			connector:       false,
			filters:         nil,
			expectedFilters: []query.Filter{andFilter(), connectorFilter("ne")},
		},
		{
			description:     "narrows to connector devices when the caller asked for them",
			connector:       true,
			filters:         nil,
			expectedFilters: []query.Filter{andFilter(), connectorFilter("eq")},
		},
		{
			description:     "applies the intent after the caller's own filters",
			connector:       true,
			filters:         []query.Filter{userFilter},
			expectedFilters: []query.Filter{userFilter, andFilter(), connectorFilter("eq")},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			storeMock := storemock.NewMockStore(tt)
			queryOptionsMock := storemock.NewMockQueryOptions(tt)
			storeMock.On("Options").Return(queryOptionsMock).Maybe()

			queryOptionsMock.On("Match", &query.Filters{Data: tc.expectedFilters}).Return(nil).Once()
			queryOptionsMock.On("Sort", mock.Anything).Return(nil).Once()
			queryOptionsMock.On("Paginate", mock.Anything).Return(nil).Once()
			storeMock.On("NamespaceGetDeviceLimit", mock.Anything, tenantID).Return(models.NamespaceDeviceLimit{}, nil).Once()
			storeMock.
				On("DeviceList", mock.Anything, scope.MustBounded(tenantID), store.DeviceAcceptableIfNotAccepted, mock.Anything).
				Return([]models.Device{}, 0, nil).
				Once()

			service := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache())

			req := &requests.DeviceList{
				TenantID:  tenantID,
				Connector: tc.connector,
				Paginator: query.Paginator{Page: 1, PerPage: 10},
				Sorter:    query.Sorter{By: "created_at", Order: query.OrderAsc},
				Filters:   query.Filters{Data: tc.filters},
			}

			_, _, err := service.ListDevices(t.Context(), scope.MustBounded(tenantID), req)
			require.NoError(tt, err)
		})
	}
}

// TestListDevicesConnectorIntentRespectsTheFilterLimit pins that the pair the service appends is
// counted like any other filter. Appending it around the limit would let a caller buy two extra
// filter entries by asking for containers.
func TestListDevicesConnectorIntentRespectsTheFilterLimit(t *testing.T) {
	const tenantID = "00000000-0000-4000-0000-000000000000"

	filters := make([]query.Filter, 0, query.MaxFilterItems)
	for range query.MaxFilterItems {
		filters = append(filters, query.Filter{
			Type:   query.FilterTypeProperty,
			Params: &query.FilterProperty{Name: "name", Operator: "contains", Value: "foo"},
		})
	}

	storeMock := storemock.NewMockStore(t)
	service := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache())

	req := &requests.DeviceList{
		TenantID:  tenantID,
		Paginator: query.Paginator{Page: 1, PerPage: 10},
		Sorter:    query.Sorter{By: "created_at", Order: query.OrderAsc},
		Filters:   query.Filters{Data: filters},
	}

	_, _, err := service.ListDevices(t.Context(), scope.MustBounded(tenantID), req)
	require.ErrorIs(t, err, ErrDeviceFilterInvalid)
}
