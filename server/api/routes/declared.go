package routes

import (
	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/server/api/routes/routetable"
)

// Declared returns the address of every route [NewRouter] mounts without [routetable.Mount], keyed
// to the reason it is served there. The core route table and the route extensions mount through
// [routetable.Mount] and are not included.
func Declared() map[string]string {
	return map[string]string{
		routetable.Address(echo.RouteAny, "/mcp"):   "the MCP server replays each tool call against the table's own routes",
		routetable.Address(echo.RouteAny, "/mcp/*"): "the MCP server replays each tool call against the table's own routes",
	}
}
