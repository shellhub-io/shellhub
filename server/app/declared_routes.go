package app

import (
	"errors"
	"maps"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/server/api/routes"
	"github.com/shellhub-io/shellhub/server/api/routes/routetable"
	sshhttp "github.com/shellhub-io/shellhub/server/ssh/http"
	"github.com/shellhub-io/shellhub/server/ssh/web"
)

var errUndeclaredRoutes = errors.New("routes are mounted without a stated access")

var sshRoutes = map[string]string{
	routetable.Address(http.MethodGet, sshhttp.HandleConnectionV1Path): "the agent's connection, authenticated by the SSH server",
	routetable.Address(http.MethodGet, sshhttp.HandleConnectionV2Path): "the agent's connection, authenticated by the SSH server",
	routetable.Address(http.MethodGet, sshhttp.HandleRevdialPath):      "the agent's reverse dial, authenticated by the SSH server",
	routetable.Address(http.MethodPost, sshhttp.HandleSSHClosePath):    "the SSH server closes its own sessions",
	routetable.Address(http.MethodPost, web.WebSessionRoute):           "the web terminal's credentials, authenticated by the core authenticator",
	routetable.Address(http.MethodGet, web.WebsocketSSHBridgeRoute):    "the web terminal's socket, opened by the token its session minted",
}

func requireDeclaredRoutes(router *echo.Echo, edition envs.Edition) error {
	declared := routes.Declared(edition)
	maps.Copy(declared, sshRoutes)

	if undeclared := routetable.Undeclared(router.Router().Routes(), declared); len(undeclared) > 0 {
		return errors.Join(errUndeclaredRoutes, errors.New(strings.Join(undeclared, ", ")))
	}

	return nil
}
