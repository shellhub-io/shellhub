// Package gatewaytest builds gateway endpoints for tests that mount a route table without its real
// handlers.
package gatewaytest

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// MemberRequest is what a [Recording] endpoint taking a [gateway.Member] binds: the acting user's ID,
// required, as the real member handlers' requests declare it.
type MemberRequest struct {
	UserID string `identity:"id" json:"-" validate:"required"`
}

// Recording returns an endpoint whose handler takes the caller kind names, sets *reached when it
// runs and answers 200 with no body. For [gateway.CallerUnstated] it returns a [gateway.Raw]
// endpoint, which every access admits; for any other kind it returns a typed [gateway.None], so the
// gateway still refuses a caller of another kind before the handler runs. The member endpoint binds
// a [MemberRequest], so an API key the route table let through answers 400 from validation, as on a
// real member route, and only a guard in front of the endpoint answers it 403.
func Recording(kind gateway.CallerKind, reached *bool) gateway.Endpoint {
	switch kind {
	case gateway.CallerAnonymous:
		return gateway.None(func(context.Context, scope.Scope, gateway.Anonymous, *struct{}) error {
			*reached = true

			return nil
		})
	case gateway.CallerPrincipal:
		return gateway.None(func(context.Context, scope.Scope, gateway.Principal, *struct{}) error {
			*reached = true

			return nil
		})
	case gateway.CallerMember:
		return gateway.None(func(context.Context, scope.Scope, gateway.Member, *MemberRequest) error {
			*reached = true

			return nil
		})
	case gateway.CallerAdmin:
		return gateway.None(func(context.Context, scope.Scope, gateway.Admin, *struct{}) error {
			*reached = true

			return nil
		})
	case gateway.CallerUnstated:
	}

	return gateway.Raw(func(c *echo.Context) error {
		*reached = true

		return c.NoContent(http.StatusOK)
	})
}
