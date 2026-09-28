package gateway_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const probeTenant = "00000000-0000-4000-0000-000000000000"

type probeRequest struct {
	Count int    `query:"count"`
	Name  string `query:"name" validate:"omitempty,min=3"`
	query.Paginator
	query.Sorter
	query.Filters
}

type probeCall struct {
	called bool
	scope  scope.Scope
	actor  gateway.Actor
	req    *probeRequest
	tenant string
}

func probeRouter(t *testing.T, withGatewayContext bool) *echo.Echo {
	t.Helper()

	e := echo.New()
	e.Binder = handlers.NewBinder()
	e.Validator = handlers.NewValidator()
	e.HTTPErrorHandler = handlers.NewErrors(nil)

	if withGatewayContext {
		e.Use(gateway.WithContext(nil))
	}

	return e
}

func probeHandler(call *probeCall, res []string, count int, err error) gateway.ListHandler[probeRequest, []string] {
	return func(ctx context.Context, sc scope.Scope, actor gateway.Actor, req *probeRequest) ([]string, int, error) {
		call.called = true
		call.scope = sc
		call.actor = actor
		call.req = req

		if tenant := gateway.TenantFromContext(ctx); tenant != nil {
			call.tenant = tenant.ID
		}

		return res, count, err
	}
}

func TestWrapperCeremony(t *testing.T) {
	cases := []struct {
		description        string
		withGatewayContext bool
		headers            map[string]string
		target             string
		options            []gateway.RouteOption
		expectedStatus     int
		expectedCall       bool
		assert             func(*testing.T, *probeCall)
	}{
		{
			description:        "normalizes the paginator and the sorter before the handler sees them",
			withGatewayContext: true,
			headers:            map[string]string{"X-Tenant-ID": probeTenant, "X-ID": "user-id"},
			target:             "/probe?page=0&per_page=999&order_by=sideways",
			expectedStatus:     http.StatusOK,
			expectedCall:       true,
			assert: func(t *testing.T, call *probeCall) {
				t.Helper()

				assert.Equal(t, query.MinPage, call.req.Paginator.Page)
				assert.Equal(t, query.MaxPerPage, call.req.Paginator.PerPage)
				assert.Equal(t, query.OrderDesc, call.req.Sorter.Order)
			},
		},
		{
			description:        "refuses a request whose query cannot bind",
			withGatewayContext: true,
			headers:            map[string]string{"X-Tenant-ID": probeTenant, "X-ID": "user-id"},
			target:             "/probe?count=not-a-number",
			expectedStatus:     http.StatusUnprocessableEntity,
		},
		{
			description:        "refuses a request that fails validation",
			withGatewayContext: true,
			headers:            map[string]string{"X-Tenant-ID": probeTenant, "X-ID": "user-id"},
			target:             "/probe?name=ab",
			expectedStatus:     http.StatusBadRequest,
		},
		{
			description:        "bounds the handler to the namespace the caller carries",
			withGatewayContext: true,
			headers:            map[string]string{"X-Tenant-ID": probeTenant, "X-ID": "user-id"},
			target:             "/probe",
			expectedStatus:     http.StatusOK,
			expectedCall:       true,
			assert: func(t *testing.T, call *probeCall) {
				t.Helper()

				assert.Equal(t, scope.MustBounded(probeTenant), call.scope)
			},
		},
		{
			description:        "refuses a bounded route when the caller carries no namespace",
			withGatewayContext: true,
			headers:            map[string]string{"X-ID": "user-id"},
			target:             "/probe",
			expectedStatus:     http.StatusForbidden,
		},
		{
			description:        "hands an unbounded route the reason its registration stated",
			withGatewayContext: true,
			headers:            map[string]string{"X-ID": "user-id"},
			target:             "/probe",
			options:            []gateway.RouteOption{gateway.Unbounded("the probe reads every namespace")},
			expectedStatus:     http.StatusOK,
			expectedCall:       true,
			assert: func(t *testing.T, call *probeCall) {
				t.Helper()

				assert.False(t, call.scope.IsBounded())
				assert.Equal(t, "the probe reads every namespace", call.scope.Reason())
			},
		},
		{
			description:        "hands the handler the identity the request authenticated as",
			withGatewayContext: true,
			headers: map[string]string{
				"X-Tenant-ID": probeTenant,
				"X-ID":        "user-id",
				"X-Username":  "username",
			},
			target:         "/probe",
			expectedStatus: http.StatusOK,
			expectedCall:   true,
			assert: func(t *testing.T, call *probeCall) {
				t.Helper()

				assert.Equal(t, gateway.Actor{ID: "user-id", Username: "username"}, call.actor)
			},
		},
		{
			description:        "accepts an api key as the acting identity",
			withGatewayContext: true,
			headers:            map[string]string{"X-Tenant-ID": probeTenant, "X-API-Key": "key"},
			target:             "/probe",
			expectedStatus:     http.StatusOK,
			expectedCall:       true,
			assert: func(t *testing.T, call *probeCall) {
				t.Helper()

				assert.Equal(t, gateway.Actor{APIKey: "key"}, call.actor)
			},
		},
		{
			description:        "refuses a route that requires an actor when the request carries none",
			withGatewayContext: true,
			headers:            map[string]string{"X-Tenant-ID": probeTenant},
			target:             "/probe",
			expectedStatus:     http.StatusUnauthorized,
		},
		{
			description:        "runs an anonymous route with no actor at all",
			withGatewayContext: true,
			headers:            map[string]string{"X-Tenant-ID": probeTenant},
			target:             "/probe",
			options:            []gateway.RouteOption{gateway.Anonymous("the probe establishes the actor")},
			expectedStatus:     http.StatusOK,
			expectedCall:       true,
			assert: func(t *testing.T, call *probeCall) {
				t.Helper()

				assert.True(t, call.actor.IsZero())
			},
		},
		{
			description:        "refuses the request when the gateway context is not installed",
			withGatewayContext: false,
			headers:            map[string]string{"X-Tenant-ID": probeTenant, "X-ID": "user-id"},
			target:             "/probe",
			expectedStatus:     http.StatusInternalServerError,
		},
		{
			description:        "keeps the gateway context reachable from the request context",
			withGatewayContext: true,
			headers:            map[string]string{"X-Tenant-ID": probeTenant, "X-ID": "user-id"},
			target:             "/probe",
			expectedStatus:     http.StatusOK,
			expectedCall:       true,
			assert: func(t *testing.T, call *probeCall) {
				t.Helper()

				assert.Equal(t, probeTenant, call.tenant)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			call := new(probeCall)

			e := probeRouter(t, tc.withGatewayContext)
			gateway.GET(rootOf(e), "/probe", gateway.List(probeHandler(call, []string{"item"}, 1, nil)), tc.options...)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.target, nil)
			for name, value := range tc.headers {
				req.Header.Set(name, value)
			}

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code, rec.Body.String())
			require.Equal(t, tc.expectedCall, call.called)

			if tc.assert != nil {
				tc.assert(t, call)
			}
		})
	}
}

// TestListWritesTheTotalCountAfterTheErrorCheck pins the answer the twelve hand-written copies of
// this header disagreed on: the count the handler returned, and only once the call succeeded.
func TestListWritesTheTotalCountAfterTheErrorCheck(t *testing.T) {
	cases := []struct {
		description    string
		count          int
		err            error
		expectedStatus int
		expectedCount  string
	}{
		{
			description:    "writes the count the handler returned",
			count:          42,
			expectedStatus: http.StatusOK,
			expectedCount:  "42",
		},
		{
			description:    "writes no count when the handler failed",
			count:          42,
			err:            errors.New("boom", "route", 3),
			expectedStatus: http.StatusUnauthorized,
			expectedCount:  "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			call := new(probeCall)

			e := probeRouter(t, true)
			gateway.GET(rootOf(e), "/probe", gateway.List(probeHandler(call, []string{"item"}, tc.count, tc.err)))

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/probe", nil)
			req.Header.Set("X-Tenant-ID", probeTenant)
			req.Header.Set("X-ID", "user-id")

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code, rec.Body.String())
			assert.Equal(t, tc.expectedCount, rec.Header().Get("X-Total-Count"))
		})
	}
}

// TestWrapperNormalizesNothingForARequestCarryingNeither drives the case the accessors exist to
// distinguish: a request embedding no paginator and no sorter is passed through untouched, rather
// than normalized against values the wrapper invented.
func TestWrapperNormalizesNothingForARequestCarryingNeither(t *testing.T) {
	type plainRequest struct {
		UID string `query:"uid"`
	}

	var got *plainRequest

	e := probeRouter(t, true)
	gateway.GET(rootOf(e), "/probe", gateway.One(func(_ context.Context, _ scope.Scope, _ gateway.Actor, req *plainRequest) (string, error) {
		got = req

		return "ok", nil
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/probe?uid=device&page=0&per_page=999", nil)
	req.Header.Set("X-Tenant-ID", probeTenant)
	req.Header.Set("X-ID", "user-id")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotNil(t, got)
	assert.Equal(t, "device", got.UID)
}

func TestOneEncodesTheHandlerResult(t *testing.T) {
	e := probeRouter(t, true)
	gateway.GET(rootOf(e), "/probe", gateway.One(func(_ context.Context, _ scope.Scope, _ gateway.Actor, _ *probeRequest) (map[string]string, error) {
		return map[string]string{"name": "value"}, nil
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/probe", nil)
	req.Header.Set("X-Tenant-ID", probeTenant)
	req.Header.Set("X-ID", "user-id")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.JSONEq(t, `{"name":"value"}`, rec.Body.String())
	assert.Empty(t, rec.Header().Get("X-Total-Count"))
}

func TestNoneAnswersWithoutABody(t *testing.T) {
	e := probeRouter(t, true)
	gateway.GET(rootOf(e), "/probe", gateway.None(func(_ context.Context, _ scope.Scope, _ gateway.Actor, _ *probeRequest) error {
		return nil
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/probe", nil)
	req.Header.Set("X-Tenant-ID", probeTenant)
	req.Header.Set("X-ID", "user-id")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.String())
}

// TestDeclarationsRecordEveryClaim makes the claims a route table makes readable by a test: a route
// declaring breadth or anonymity must be able to show the reason it typed.
func TestDeclarationsRecordEveryClaim(t *testing.T) {
	const unboundedReason = "the declared probe reads every namespace"

	e := probeRouter(t, true)
	gateway.GET(rootOf(e), "/declared", gateway.List(probeHandler(new(probeCall), nil, 0, nil)),
		gateway.Unbounded(unboundedReason),
		gateway.Anonymous("the declared probe establishes the actor"))

	var found bool

	for _, declaration := range gateway.Declarations(e) {
		if declaration.UnboundedReason != unboundedReason {
			continue
		}

		found = true

		assert.Equal(t, gateway.ShapeList, declaration.Shape)
		assert.True(t, declaration.Anonymous)
		assert.Equal(t, "the declared probe establishes the actor", declaration.AnonymousReason)
		assert.NotEmpty(t, declaration.Handler)
	}

	assert.True(t, found, "the wrapper recorded no declaration for the probe route")
}

// TestListHoldsTheQueryToTheContractItsRegistrationNamed drives the contract mechanism once,
// through a mounted route: the wrapper is what decodes the filter and refuses a field the
// resource does not allow, so no handler opens with a validation preamble.
func TestListHoldsTheQueryToTheContractItsRegistrationNamed(t *testing.T) {
	contract := query.Contract{
		Filter:      query.NewFieldConstraints(map[string][]string{"name": {"contains"}}),
		Sort:        query.NewFieldSet("name", "created_at"),
		DefaultSort: query.Sorter{By: "created_at", Order: query.OrderAsc},
	}

	cases := []struct {
		description    string
		target         string
		expectedStatus int
		expectedFields map[string]string
		assert         func(*testing.T, *probeCall)
	}{
		{
			description:    "refuses a filter naming a field the contract does not allow",
			target:         "/probe?filter=" + encodeProbeFilter(t, "signature", "contains"),
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]string{"filter": "is not valid"},
		},
		{
			description:    "refuses a filter naming an operator the contract does not allow",
			target:         "/probe?filter=" + encodeProbeFilter(t, "name", "eq"),
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]string{"filter": "is not valid"},
		},
		{
			description:    "refuses a filter that is not base64",
			target:         "/probe?filter=not-base64!!",
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]string{"filter": "cannot be decoded"},
		},
		{
			description:    "refuses a filter larger than the cap",
			target:         "/probe?filter=" + strings.Repeat("A", query.MaxFilterRawBytes+1),
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]string{"filter": "cannot be decoded"},
		},
		{
			description:    "refuses a sort naming a field the contract does not allow",
			target:         "/probe?sort_by=secret",
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]string{"sort_by": "secret"},
		},
		{
			description:    "hands the handler a decoded filter, a normalized page and the contract's default sort",
			target:         "/probe?page=0&per_page=999&filter=" + encodeProbeFilter(t, "name", "contains"),
			expectedStatus: http.StatusOK,
			assert: func(t *testing.T, call *probeCall) {
				t.Helper()

				require.Len(t, call.req.Filters.Data, 1)
				assert.Equal(t, &query.FilterProperty{Name: "name", Operator: "contains", Value: "value"}, call.req.Filters.Data[0].Params)
				assert.Equal(t, query.MinPage, call.req.Paginator.Page)
				assert.Equal(t, query.MaxPerPage, call.req.Paginator.PerPage)
				assert.Equal(t, "created_at", call.req.Sorter.By)
				assert.Equal(t, query.OrderAsc, call.req.Sorter.Order)
			},
		},
		{
			description:    "leaves a sort the client asked for alone when the contract allows it",
			target:         "/probe?sort_by=name&order_by=asc",
			expectedStatus: http.StatusOK,
			assert: func(t *testing.T, call *probeCall) {
				t.Helper()

				assert.Equal(t, "name", call.req.Sorter.By)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			call := new(probeCall)

			e := probeRouter(t, true)
			gateway.GET(rootOf(e), "/probe", gateway.List(probeHandler(call, []string{"item"}, 1, nil)), gateway.Accepts(contract))

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.target, nil)
			req.Header.Set("X-Tenant-ID", probeTenant)
			req.Header.Set("X-ID", "user-id")

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code, rec.Body.String())
			require.Equal(t, tc.expectedStatus == http.StatusOK, call.called)

			if tc.expectedFields != nil {
				var body responses.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, tc.expectedFields, body.Fields)
			}

			if tc.assert != nil {
				tc.assert(t, call)
			}
		})
	}
}

// TestAcceptsRecordsTheContractOnTheDeclaration is what the route table's invariant reads: a claim
// the wrapper enforces still has to be visible to an audit of the declarations.
func TestAcceptsRecordsTheContractOnTheDeclaration(t *testing.T) {
	e := probeRouter(t, true)
	gateway.GET(rootOf(e), "/accepting", gateway.List(probeHandler(new(probeCall), nil, 0, nil)),
		gateway.Accepts(query.Contract{Sort: query.NewFieldSet("name")}))
	gateway.GET(rootOf(e), "/silent", gateway.List(probeHandler(new(probeCall), nil, 0, nil)))

	accepts := make(map[string]bool)
	for _, declaration := range gateway.Declarations(e) {
		accepts[declaration.Path] = declaration.AcceptsQuery
	}

	assert.True(t, accepts["/accepting"])
	assert.False(t, accepts["/silent"])
}

func encodeProbeFilter(t *testing.T, name, operator string) string {
	t.Helper()

	encoded, err := json.Marshal([]query.Filter{
		{Type: query.FilterTypeProperty, Params: &query.FilterProperty{Name: name, Operator: operator, Value: "value"}},
	})
	require.NoError(t, err)

	return base64.RawURLEncoding.EncodeToString(encoded)
}
