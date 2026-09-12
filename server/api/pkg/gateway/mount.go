package gateway

import (
	"net/http"
	"sort"
	"sync"

	"github.com/labstack/echo/v5"
)

// Target is the part of echo's routing API mounting needs. Both *echo.Echo and *echo.Group
// satisfy it, and both answer with the address the route ended up at — the group prefix already
// applied — which is the address the router reports and the authenticator matches on.
type Target interface {
	Add(method, path string, handler echo.HandlerFunc, middleware ...echo.MiddlewareFunc) echo.RouteInfo
}

// Mounter mounts routes onto one target and declares them against one router. Mounting and
// declaring are the same act: a declaration cannot be recorded for a route nobody mounted, and a
// route mounted through a mounter cannot escape the declaration.
//
// The router is carried separately from the target because a group does not name the router it
// was carved from, and it is the router a claim belongs to.
type Mounter struct {
	router *echo.Echo
	target Target
}

// MountOn returns a mounter adding routes to target and declaring them against router. Pass
// router as target to mount on its root; pass a group of it to mount under that group's prefix.
func MountOn(router *echo.Echo, target Target) *Mounter {
	return &Mounter{router: router, target: target}
}

// GET mounts route at path answering GET, declares it against the mounter's router, and returns
// the address it was mounted at — which is the address the declaration then carries. Each option's
// guard wraps the handler in the order the option is written, so the guard written first is the
// one that answers first.
//
// [POST], [PUT], [PATCH] and [DELETE] do the same for their methods.
func GET(m *Mounter, path string, route Route, options ...RouteOption) echo.RouteInfo {
	return mount(m, http.MethodGet, path, route, options)
}

// POST mounts route at path, answering POST. See [GET] for what mounting declares.
func POST(m *Mounter, path string, route Route, options ...RouteOption) echo.RouteInfo {
	return mount(m, http.MethodPost, path, route, options)
}

// PUT mounts route at path, answering PUT. See [GET] for what mounting declares.
func PUT(m *Mounter, path string, route Route, options ...RouteOption) echo.RouteInfo {
	return mount(m, http.MethodPut, path, route, options)
}

// PATCH mounts route at path, answering PATCH. See [GET] for what mounting declares.
func PATCH(m *Mounter, path string, route Route, options ...RouteOption) echo.RouteInfo {
	return mount(m, http.MethodPatch, path, route, options)
}

// DELETE mounts route at path, answering DELETE. See [GET] for what mounting declares.
func DELETE(m *Mounter, path string, route Route, options ...RouteOption) echo.RouteInfo {
	return mount(m, http.MethodDelete, path, route, options)
}

func mount(m *Mounter, method, path string, route Route, options []RouteOption) echo.RouteInfo {
	declaration := route.declaration

	guards := make([]echo.MiddlewareFunc, 0, len(options))

	for _, option := range options {
		if guard := option(&declaration); guard != nil {
			guards = append(guards, guard)
		}
	}

	info := m.target.Add(method, path, route.build(declaration), guards...)

	declaration.Method, declaration.Path = info.Method, info.Path

	declare(m.router, declaration)

	return info
}

var tables = struct {
	sync.Mutex
	byRouter map[*echo.Echo][]Declaration
}{byRouter: make(map[*echo.Echo][]Declaration)}

func declare(router *echo.Echo, declaration Declaration) {
	tables.Lock()
	defer tables.Unlock()

	tables.byRouter[router] = append(tables.byRouter[router], declaration)
}

// Declarations returns what every route mounted on router claims, ordered by address. The table
// belongs to the router rather than to the process, so an invariant over it holds regardless of
// which other routers — under which other editions — a neighbouring test built.
//
// Repeats are kept: two declarations sharing an address is how a shadowed route shows up, and
// collapsing them here would hide it.
func Declarations(router *echo.Echo) []Declaration {
	tables.Lock()
	defer tables.Unlock()

	all := append([]Declaration(nil), tables.byRouter[router]...)

	sort.Slice(all, func(i, j int) bool {
		if all[i].Path != all[j].Path {
			return all[i].Path < all[j].Path
		}

		return all[i].Method < all[j].Method
	})

	return all
}
