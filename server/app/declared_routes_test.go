package app

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/server/api/routes"
	"github.com/shellhub-io/shellhub/server/api/routes/middleware"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	sshhttp "github.com/shellhub-io/shellhub/server/ssh/http"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/webhandoff"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	"github.com/shellhub-io/shellhub/server/ssh/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func writeHostKey(t *testing.T) string {
	t.Helper()

	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	block, err := ssh.MarshalPrivateKey(key, "")
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "host.key")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(block), 0o600))

	return path
}

func assembledRouter(t *testing.T, edition envs.Edition) *echo.Echo {
	t.Helper()

	service := servicemocks.NewMockService(t)
	authn := middleware.NewAuthenticator(service)

	router, err := routes.NewRouter(service, routes.WithAuthentication(authn), routes.WithEdition(edition))
	require.NoError(t, err)

	sshhttp.Register(router, authn, dialer.NewDialer(nil, nil, clock.New()), service, session.NewRegistry(), &sshhttp.Config{})
	require.NoError(t, web.NewSSHServerBridge(router, authn, service, webhandoff.NewStore(), &web.Config{HostKeyFile: writeHostKey(t)}))

	return router
}

func TestRequireDeclaredRoutes(t *testing.T) {
	for _, edition := range []envs.Edition{envs.Community, envs.Enterprise, envs.Cloud} {
		t.Run(string(edition), func(t *testing.T) {
			router := assembledRouter(t, edition)

			require.NoError(t, requireDeclaredRoutes(router, edition))

			router.GET("/api/undeclared", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })

			assert.ErrorContains(t, requireDeclaredRoutes(router, edition), "GET /api/undeclared")
		})
	}
}
