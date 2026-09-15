package routes

import (
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	sshhttp "github.com/shellhub-io/shellhub/server/ssh/http"
	sshweb "github.com/shellhub-io/shellhub/server/ssh/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var gatewayExemptRoutes = map[string]string{
	"GET " + InternalMetricsURL: "the Prometheus registry answers a scrape, which carries no identity and reads no namespace",
	echo.RouteAny + " /mcp":     "the MCP transport dispatches to the routes below it, which is where the claims are",
	echo.RouteAny + " /mcp/*":   "the MCP transport dispatches to the routes below it, which is where the claims are",
}

var composedServerPaths = map[string]string{
	sshhttp.HandleConnectionV1Path: "the agent's tunnel is a websocket the SSH sidecar owns, and it answers to a device token rather than to a namespace member",
	sshhttp.HandleConnectionV2Path: "the agent's tunnel is a websocket the SSH sidecar owns, and it answers to a device token rather than to a namespace member",
	sshhttp.HandleRevdialPath:      "the reverse dial hands the connection to the dialer, and never reaches a handler that could bind a request",
	sshhttp.HandleSSHClosePath:     "the SSH package mounts it on the router root, so it carries an /api address without entering that group's chain: the authenticator installed at the root is what guards it",
	sshweb.WebSessionRoute:         "the web terminal's handoff is served by a raw http.Handler, outside echo's binder and validator",
	sshweb.WebsocketSSHBridgeRoute: "the web terminal's websocket is served by a raw http.Handler, outside echo's binder and validator",
	"/debug/pprof":                 "the profiler group is registered only in development, by the server rather than by the router",
}

func addresses(router *echo.Echo) map[string]struct{} {
	registered := make(map[string]struct{})
	for _, route := range router.Router().Routes() {
		registered[route.Method+" "+route.Path] = struct{}{}
	}

	return registered
}

func unstatedClaims(declarations []gateway.Declaration) []string {
	unstated := make([]string, 0)

	for _, declaration := range declarations {
		if declaration.Unbounded && strings.TrimSpace(declaration.UnboundedReason) == "" {
			unstated = append(unstated, declaration.Address()+" reads across namespaces and states no reason")
		}

		if declaration.Anonymous && strings.TrimSpace(declaration.AnonymousReason) == "" {
			unstated = append(unstated, declaration.Address()+" requires no actor and states no reason")
		}
	}

	return unstated
}

func unstatedAuthority(declarations []gateway.Declaration) []string {
	unstated := make([]string, 0)

	for _, declaration := range declarations {
		switch declaration.Authority {
		case gateway.AuthorityPermission:
			if len(declaration.Permissions) == 0 {
				unstated = append(unstated, declaration.Address()+" demands a permission and names none")
			}
		case gateway.AuthorityInHandler, gateway.AuthorityNone:
			if strings.TrimSpace(declaration.AuthorityReason) == "" {
				unstated = append(unstated, declaration.Address()+" leaves its authority to the handler or demands none, and states no reason")
			}
		case gateway.AuthorityUnstated:
			if !declaration.Anonymous {
				unstated = append(unstated, declaration.Address()+" demands no permission and states no reason")
			}
		}
	}

	sort.Strings(unstated)

	return unstated
}

func undeclaredRoutes(registered map[string]struct{}, declarations []gateway.Declaration, exempt map[string]string) []string {
	declared := make(map[string]struct{}, len(declarations))
	for _, declaration := range declarations {
		declared[declaration.Address()] = struct{}{}
	}

	undeclared := make([]string, 0)

	for address := range registered {
		if _, ok := declared[address]; ok {
			continue
		}

		if reason, ok := exempt[address]; ok && strings.TrimSpace(reason) != "" {
			continue
		}

		undeclared = append(undeclared, address+" is mounted but claims nothing")
	}

	sort.Strings(undeclared)

	return undeclared
}

func unmountedDeclarations(registered map[string]struct{}, declarations []gateway.Declaration) []string {
	unmounted := make([]string, 0)

	for _, declaration := range declarations {
		if _, ok := registered[declaration.Address()]; !ok {
			unmounted = append(unmounted, declaration.Address()+" is declared but no such route is mounted")
		}
	}

	return unmounted
}

func shadowedRoutes(declarations []gateway.Declaration) []string {
	seen := make(map[string]struct{}, len(declarations))
	shadowed := make([]string, 0)

	for _, declaration := range declarations {
		if _, ok := seen[declaration.Address()]; ok {
			shadowed = append(shadowed, declaration.Address()+" is mounted more than once")

			continue
		}

		seen[declaration.Address()] = struct{}{}
	}

	return shadowed
}

func anonymityMismatches(declarations []gateway.Declaration, allowlist []string) []string {
	allowed := make(map[string]struct{}, len(allowlist))
	for _, entry := range allowlist {
		allowed[entry] = struct{}{}
	}

	mismatches := make([]string, 0)

	for _, declaration := range declarations {
		_, byMethod := allowed[declaration.Address()]
		_, byAnyMethod := allowed[routesmiddleware.AnyMethod+" "+declaration.Path]

		if declaration.Anonymous == (byMethod || byAnyMethod) {
			continue
		}

		if declaration.Anonymous {
			mismatches = append(mismatches, declaration.Address()+" claims no actor but the authenticator demands a credential")

			continue
		}

		mismatches = append(mismatches, declaration.Address()+" is reachable without a credential but claims an actor")
	}

	return mismatches
}

func anonymityDisagreements(declarations []gateway.Declaration) []string {
	byHandler := make(map[string][]gateway.Declaration, len(declarations))
	for _, declaration := range declarations {
		byHandler[declaration.Handler] = append(byHandler[declaration.Handler], declaration)
	}

	disagreements := make([]string, 0)

	for _, mounts := range byHandler {
		anonymous, credentialed := make([]string, 0), make([]string, 0)

		for _, declaration := range mounts {
			if declaration.Anonymous {
				anonymous = append(anonymous, declaration.Address())

				continue
			}

			credentialed = append(credentialed, declaration.Address())
		}

		if len(anonymous) == 0 || len(credentialed) == 0 {
			continue
		}

		sort.Strings(anonymous)
		sort.Strings(credentialed)

		disagreements = append(disagreements,
			strings.Join(credentialed, ", ")+" demand a credential but "+strings.Join(anonymous, ", ")+" serves the same handler without one")
	}

	sort.Strings(disagreements)

	return disagreements
}

func misplacedComposedRoutes(router *echo.Echo, composed map[string]string) []string {
	misplaced := make([]string, 0)

	for path, reason := range composed {
		if strings.TrimSpace(reason) == "" {
			misplaced = append(misplaced, path+" is named as the composed server's and states no reason")
		}

		for _, route := range router.Router().Routes() {
			if route.Path == path || strings.HasPrefix(route.Path, path+"/") {
				misplaced = append(misplaced, path+" is named as the composed server's but the router mounts it")

				break
			}
		}
	}

	sort.Strings(misplaced)

	return misplaced
}

func staleExemptions(registered map[string]struct{}, exempt map[string]string) []string {
	stale := make([]string, 0)

	for address, reason := range exempt {
		if _, ok := registered[address]; !ok {
			stale = append(stale, address+" is exempt but no such route is mounted")
		}

		if strings.TrimSpace(reason) == "" {
			stale = append(stale, address+" is exempt and states no reason")
		}
	}

	sort.Strings(stale)

	return stale
}

func unqueriedListRoutes(declarations []gateway.Declaration) []string {
	unqueried := make([]string, 0)

	for _, declaration := range declarations {
		if declaration.Shape == gateway.ShapeList && !declaration.AcceptsQuery {
			unqueried = append(unqueried, declaration.Address()+" serves a page and names no query contract")
		}
	}

	sort.Strings(unqueried)

	return unqueried
}

// TestRouteTableHoldsItsClaims reads the whole route table of a fully built router against the
// claims its registrations made. Every check is one predicate over that table, and each has a
// companion below feeding it a known-bad input, so a passing run means the predicate looked.
func TestRouteTableHoldsItsClaims(t *testing.T) {
	router, authn, _ := authenticatedRouter(t)

	declarations := gateway.Declarations(router)
	require.NotEmpty(t, declarations, "the route table registered no declaration")

	registered := addresses(router)

	t.Run("every claim that takes an exception states why", func(t *testing.T) {
		assert.Empty(t, unstatedClaims(declarations))
	})

	t.Run("every route states the authority it demands", func(t *testing.T) {
		assert.Empty(t, unstatedAuthority(declarations))
	})

	t.Run("every mounted route is declared or exempt", func(t *testing.T) {
		assert.Empty(t, undeclaredRoutes(registered, declarations, gatewayExemptRoutes))
	})

	t.Run("every declaration names a mounted route", func(t *testing.T) {
		assert.Empty(t, unmountedDeclarations(registered, declarations))
	})

	t.Run("no address is mounted twice", func(t *testing.T) {
		assert.Empty(t, shadowedRoutes(declarations))
	})

	t.Run("the anonymity claims and the allowlist agree", func(t *testing.T) {
		assert.Empty(t, anonymityMismatches(declarations, authn.AnonymousRoutes()))
	})

	t.Run("routes serving the same handler agree on whether it needs a credential", func(t *testing.T) {
		assert.Empty(t, anonymityDisagreements(declarations))
	})

	t.Run("every list route names a query contract", func(t *testing.T) {
		assert.Empty(t, unqueriedListRoutes(declarations))
	})

	t.Run("every exemption names a mounted route and states why", func(t *testing.T) {
		assert.Empty(t, staleExemptions(registered, gatewayExemptRoutes))
	})

	t.Run("the composed server's routes are named and are not this router's", func(t *testing.T) {
		assert.Empty(t, misplacedComposedRoutes(router, composedServerPaths))
	})
}

// TestUnstatedClaimsRefusesAnEmptyReason proves the check above bites, rather than passing because
// it looks at nothing.
func TestUnstatedClaimsRefusesAnEmptyReason(t *testing.T) {
	unstated := unstatedClaims([]gateway.Declaration{
		{Method: "GET", Path: "/silent", Unbounded: true, Anonymous: true},
		{Method: "GET", Path: "/stated", Unbounded: true, UnboundedReason: "because"},
	})

	require.Len(t, unstated, 2)
	for _, complaint := range unstated {
		assert.Contains(t, complaint, "/silent")
	}
}

// TestUnstatedAuthorityRefusesARouteThatDemandsNothingSilently proves the check above bites. Its
// three complaints are the three ways a route can leave a reader unable to tell a decision from an
// omission: no claim at all, a claim whose reason nobody typed, and a permission set naming
// nothing.
func TestUnstatedAuthorityRefusesARouteThatDemandsNothingSilently(t *testing.T) {
	unstated := unstatedAuthority([]gateway.Declaration{
		{Method: "GET", Path: "/anonymous", Anonymous: true},
		{Method: "GET", Path: "/empty", Authority: gateway.AuthorityPermission},
		{Method: "GET", Path: "/guarded", Authority: gateway.AuthorityPermission, Permissions: []authorizer.Permission{authorizer.DeviceRemove}},
		{Method: "GET", Path: "/handler", Authority: gateway.AuthorityInHandler, AuthorityReason: "the path names the namespace"},
		{Method: "GET", Path: "/silent"},
		{Method: "GET", Path: "/unreasoned", Authority: gateway.AuthorityNone, AuthorityReason: "  "},
	})

	assert.Equal(t, []string{
		"GET /empty demands a permission and names none",
		"GET /silent demands no permission and states no reason",
		"GET /unreasoned leaves its authority to the handler or demands none, and states no reason",
	}, unstated)
}

// TestUndeclaredRoutesCatchesARouteThatClaimsNothing drives the case the invariant exists for: a
// route mounted around the gateway, which no audit would otherwise reach.
func TestUndeclaredRoutesCatchesARouteThatClaimsNothing(t *testing.T) {
	registered := map[string]struct{}{
		"GET /declared":   {},
		"GET /exempt":     {},
		"GET /smuggled":   {},
		"GET /unreasoned": {},
	}

	undeclared := undeclaredRoutes(
		registered,
		[]gateway.Declaration{{Method: "GET", Path: "/declared"}},
		map[string]string{"GET /exempt": "a reason", "GET /unreasoned": "  "},
	)

	assert.Equal(t, []string{
		"GET /smuggled is mounted but claims nothing",
		"GET /unreasoned is mounted but claims nothing",
	}, undeclared)
}

// TestUnmountedDeclarationsCatchesAStaleClaim is the other direction: a claim recorded for a route
// that was renamed or removed is believed by every reader until something checks it.
func TestUnmountedDeclarationsCatchesAStaleClaim(t *testing.T) {
	unmounted := unmountedDeclarations(
		map[string]struct{}{"GET /mounted": {}},
		[]gateway.Declaration{{Method: "GET", Path: "/mounted"}, {Method: "GET", Path: "/gone"}},
	)

	require.Len(t, unmounted, 1)
	assert.Contains(t, unmounted[0], "/gone")
}

// TestShadowedRoutesCatchesADoubleMount matters because echo overwrites silently: the second
// registration wins, and the first route's guards simply stop running.
func TestShadowedRoutesCatchesADoubleMount(t *testing.T) {
	shadowed := shadowedRoutes([]gateway.Declaration{
		{Method: "GET", Path: "/devices"},
		{Method: "POST", Path: "/devices"},
		{Method: "GET", Path: "/devices"},
	})

	require.Len(t, shadowed, 1)
	assert.Contains(t, shadowed[0], "GET /devices")
}

// TestAnonymityMismatchesCatchesBothDirections joins the two places a route's anonymity is stated.
// The gateway claim frees the handler from needing an actor; the authenticator's allowlist is what
// lets the request past the credential check. A route carrying one without the other is either
// unreachable or reachable without a credential.
func TestAnonymityMismatchesCatchesBothDirections(t *testing.T) {
	mismatches := anonymityMismatches(
		[]gateway.Declaration{
			{Method: "GET", Path: "/agreed", Anonymous: true},
			{Method: "GET", Path: "/guarded"},
			{Method: "GET", Path: "/unreachable", Anonymous: true},
			{Method: "GET", Path: "/open"},
		},
		[]string{"GET /agreed", "GET /open"},
	)

	require.Len(t, mismatches, 2)
	assert.Contains(t, mismatches[0], "/unreachable")
	assert.Contains(t, mismatches[1], "/open")
}

// TestAnonymityDisagreementsCatchesASecondSpellingThatForgot is the direction anonymityMismatches
// cannot see: a route claiming nothing and allowlisted nowhere reads as consistent to it, because
// that is what every authenticated route looks like. What distinguishes the bug is the handler —
// a second spelling of an endpoint mounted beside the first and given none of its claims answers
// 401 to the very clients it exists for, and only its sibling says so.
func TestAnonymityDisagreementsCatchesASecondSpellingThatForgot(t *testing.T) {
	disagreements := anonymityDisagreements([]gateway.Declaration{
		{Method: "POST", Path: "/devices/auth", Handler: "AuthDevice", Anonymous: true},
		{Method: "POST", Path: "/auth/device", Handler: "AuthDevice"},
		{Method: "GET", Path: "/tags", Handler: "GetTags"},
		{Method: "GET", Path: "/namespaces/:tenant/tags", Handler: "GetTags"},
		{Method: "GET", Path: "/info", Handler: "GetSystemInfo", Anonymous: true},
	})

	assert.Equal(t, []string{
		"POST /auth/device demand a credential but POST /devices/auth serves the same handler without one",
	}, disagreements)
}

// TestMisplacedComposedRoutesCatchesOneThatMoved keeps the naming honest in the only direction
// this seam can check. A route named as the composed server's that turns up on a router [NewRouter]
// built is no longer outside the coverage invariant's reach, and belongs in the exempt set where
// its exemption is checked both ways.
func TestMisplacedComposedRoutesCatchesOneThatMoved(t *testing.T) {
	router := echo.New()
	router.GET("/moved", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })
	router.GET("/nested/deep", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })

	misplaced := misplacedComposedRoutes(router, map[string]string{
		"/elsewhere": "a reason",
		"/moved":     "a reason",
		"/nested":    "a reason",
		"/silent":    "",
	})

	assert.Equal(t, []string{
		"/moved is named as the composed server's but the router mounts it",
		"/nested is named as the composed server's but the router mounts it",
		"/silent is named as the composed server's and states no reason",
	}, misplaced)
}

// TestStaleExemptionsCatchesAnExemptionNothingMounts keeps the exempt set from outliving the
// routes it excuses, and from excusing one without saying why.
func TestStaleExemptionsCatchesAnExemptionNothingMounts(t *testing.T) {
	stale := staleExemptions(
		map[string]struct{}{"GET /kept": {}, "GET /silent": {}},
		map[string]string{"GET /kept": "a reason", "GET /gone": "a reason", "GET /silent": ""},
	)

	assert.Equal(t, []string{
		"GET /gone is exempt but no such route is mounted",
		"GET /silent is exempt and states no reason",
	}, stale)
}

// TestUnqueriedListRoutesCatchesAListThatNamesNoContract replaces the go/ast test that used to
// enforce this rule by reading the requests and routes packages as source text. That test matched a
// handler by its gateway-context parameter, so a route dropped out of its coverage the moment its
// handler changed shape — and it kept passing. A route is in the table whatever shape its handler
// has, so a list route cannot leave this check by changing its signature. A route that still writes
// its own response is a ShapeLegacy one, and is covered when it converts.
func TestUnqueriedListRoutesCatchesAListThatNamesNoContract(t *testing.T) {
	unqueried := unqueriedListRoutes([]gateway.Declaration{
		{Method: "GET", Path: "/named", Shape: gateway.ShapeList, AcceptsQuery: true},
		{Method: "GET", Path: "/silent", Shape: gateway.ShapeList},
		{Method: "GET", Path: "/single", Shape: gateway.ShapeOne},
		{Method: "GET", Path: "/legacy", Shape: gateway.ShapeLegacy},
	})

	assert.Equal(t, []string{"GET /silent serves a page and names no query contract"}, unqueried)
}
