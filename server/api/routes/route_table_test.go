package routes

import (
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func unstatedClaims(declarations []gateway.Declaration) []string {
	unstated := make([]string, 0)

	for _, declaration := range declarations {
		if declaration.Unbounded && strings.TrimSpace(declaration.UnboundedReason) == "" {
			unstated = append(unstated, declaration.Handler+" reads across namespaces and states no reason")
		}

		if declaration.Anonymous && strings.TrimSpace(declaration.AnonymousReason) == "" {
			unstated = append(unstated, declaration.Handler+" requires no actor and states no reason")
		}
	}

	return unstated
}

func TestRouteTableStatesEveryClaim(t *testing.T) {
	authenticatedRouter(t)

	declarations := gateway.Declarations()
	require.NotEmpty(t, declarations, "the route table registered no wrapped route")

	assert.Empty(t, unstatedClaims(declarations))
}

func TestUnstatedClaimsRefusesAnEmptyReason(t *testing.T) {
	unstated := unstatedClaims([]gateway.Declaration{
		{Handler: "silent", Unbounded: true, Anonymous: true},
		{Handler: "stated", Unbounded: true, UnboundedReason: "because"},
	})

	require.Len(t, unstated, 2)
	for _, complaint := range unstated {
		assert.Contains(t, complaint, "silent")
	}
}

var wrapperExemptRoutes = []string{
	"GET /api/install",
	"POST /api/login",
	"POST /api/auth/user",
	"POST /api/tags",
	"POST /api/namespaces/:tenant/tags",
}

func TestWrapperExemptRoutesAreRegistered(t *testing.T) {
	router, _, _ := authenticatedRouter(t)

	registered := make(map[string]struct{})
	for _, route := range router.Router().Routes() {
		registered[route.Method+" "+route.Path] = struct{}{}
	}

	for _, exempt := range wrapperExemptRoutes {
		assert.Contains(t, registered, exempt, "the exempt set names %q but no such route is registered", exempt)
	}
}

var convertedRoutes = []struct {
	handler string
	shape   gateway.Shape
	route   string
}{
	{handler: "EvaluateHealth", shape: gateway.ShapeNone, route: "GET /api" + HealthCheckURL},
	{handler: "GetDevice", shape: gateway.ShapeOne, route: "GET /api" + GetDeviceURL},
	{handler: "GetDeviceList", shape: gateway.ShapeList, route: "GET /api" + GetDeviceListURL},
}

func methodName(qualified string) string {
	return strings.TrimSuffix(qualified[strings.LastIndex(qualified, ".")+1:], "-fm")
}

func TestAnonymousClaimsMatchTheAllowlist(t *testing.T) {
	_, authn, _ := authenticatedRouter(t)

	allowed := make(map[string]struct{})
	for _, route := range authn.AnonymousRoutes() {
		allowed[route] = struct{}{}
	}

	claimed := make(map[string]bool)
	for _, declaration := range gateway.Declarations() {
		claimed[methodName(declaration.Handler)] = declaration.Anonymous
	}

	for _, tc := range convertedRoutes {
		t.Run(tc.handler, func(t *testing.T) {
			_, inAllowlist := allowed[tc.route]

			assert.Equal(t, claimed[tc.handler], inAllowlist,
				"%s declares Anonymous=%v but the authenticator's allowlist says %v",
				tc.handler, claimed[tc.handler], inAllowlist)
		})
	}
}

func TestConvertedRoutesDeclareTheirShape(t *testing.T) {
	router, _, _ := authenticatedRouter(t)

	registered := make(map[string]struct{})
	for _, route := range router.Router().Routes() {
		registered[route.Method+" "+route.Path] = struct{}{}
	}

	shapes := make(map[string]gateway.Shape)
	for _, declaration := range gateway.Declarations() {
		shapes[methodName(declaration.Handler)] = declaration.Shape
	}

	for _, tc := range convertedRoutes {
		t.Run(tc.handler, func(tt *testing.T) {
			declared, found := shapes[tc.handler]

			require.True(tt, found, "%s is not registered through a gateway shape", tc.handler)
			assert.Equal(tt, tc.shape, declared, "%s answers with the wrong shape", tc.handler)
			assert.Contains(tt, registered, tc.route, "%s declares a shape but is not mounted", tc.handler)
		})
	}
}
