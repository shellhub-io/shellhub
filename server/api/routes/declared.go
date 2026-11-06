package routes

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/server/api/routes/routetable"
)

// Declared returns the address of every route [NewRouter] mounts for edition without
// [routetable.Mount], keyed to the reason it is served there: the rows of the core route table,
// which state their access in that table, and the routes mounted beside it. Route extensions are
// not included; they mount through [routetable.Mount].
func Declared(edition envs.Edition) map[string]string {
	declared := map[string]string{
		routetable.Address(echo.RouteAny, "/mcp"):              "the MCP server replays each tool call against the table's own routes",
		routetable.Address(echo.RouteAny, "/mcp/*"):            "the MCP server replays each tool call against the table's own routes",
		routetable.Address(http.MethodGet, "/metrics"):         "Prometheus scrapes it, and authentication exempts it",
		routetable.Address(http.MethodGet, InternalMetricsURL): "Prometheus scrapes it from inside the cluster",
	}

	handler := NewHandler(nil)
	handler.edition = edition

	for _, r := range handler.publicRoutes() {
		declared[routetable.Address(r.method, publicAPIPrefix+r.path)] = "the core route table states its access"
	}

	return declared
}
