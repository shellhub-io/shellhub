package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type identityRequest struct {
	UserID   string          `json:"user_id" identity:"id"`
	TenantID string          `param:"tenant" identity:"tenant"`
	Role     authorizer.Role `identity:"role"`
	IsAdmin  bool            `json:"is_admin" identity:"admin"`
	Name     string          `json:"name"`
}

func bindAs(t *testing.T, caller *identity.Identity, pathTenant, body string, target any) {
	t.Helper()

	ctx := context.Background()
	if caller != nil {
		ctx = identity.NewContext(ctx, *caller)
	}

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	c := echo.New().NewContext(req, httptest.NewRecorder())
	if pathTenant != "" {
		c.SetPathValues(echo.PathValues{{Name: "tenant", Value: pathTenant}})
	}

	require.NoError(t, NewBinder().Bind(c, target))
}

func TestBinderFillsIdentityFields(t *testing.T) {
	caller := identity.Identity{ID: "user", TenantID: "tenant", Role: authorizer.RoleOwner, Admin: true}

	var req identityRequest
	bindAs(t, &caller, "path-tenant", `{"name":"kept"}`, &req)

	assert.Equal(t, identityRequest{UserID: "user", TenantID: "tenant", Role: authorizer.RoleOwner, IsAdmin: true, Name: "kept"}, req)
}

func TestBinderKeepsThePathValueWhenTheIdentityHasNone(t *testing.T) {
	caller := identity.Identity{ID: "user"}

	var req identityRequest
	bindAs(t, &caller, "path-tenant", `{}`, &req)

	assert.Equal(t, "path-tenant", req.TenantID)
}
