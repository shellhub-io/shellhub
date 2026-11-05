package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/binding"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type identityRequest struct {
	UserID   string          `identity:"id" json:"-"`
	TenantID string          `identity:"tenant" json:"-"`
	Role     authorizer.Role `identity:"role" json:"-"`
	IsAdmin  bool            `identity:"admin" json:"-"`
	Name     string          `json:"name"`
}

func bindAs(t *testing.T, caller *identity.Identity, pathTenant, body string, target any) error {
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

	return NewBinder().Bind(c, target)
}

func TestBinderFillsIdentityFields(t *testing.T) {
	caller := identity.Identity{ID: "user", TenantID: "tenant", Role: authorizer.RoleOwner, Admin: true}

	var req identityRequest
	require.NoError(t, bindAs(t, &caller, "path-tenant", `{"name":"kept"}`, &req))

	assert.Equal(t, identityRequest{UserID: "user", TenantID: "tenant", Role: authorizer.RoleOwner, IsAdmin: true, Name: "kept"}, req)
}

func TestBinderRefusesABodyNamingIdentityFields(t *testing.T) {
	caller := identity.Identity{ID: "user"}

	var req identityRequest
	err := bindAs(t, &caller, "", `{"UserID":"forged","IsAdmin":true}`, &req)

	require.ErrorContains(t, err, "IsAdmin, UserID")
	assert.Empty(t, req.UserID)
	assert.False(t, req.IsAdmin)
}

func TestBinderLeavesIdentityFieldsEmptyWhenAnonymous(t *testing.T) {
	var req identityRequest
	require.NoError(t, bindAs(t, nil, "", `{"name":"kept"}`, &req))

	assert.Empty(t, req.UserID)
	assert.False(t, req.IsAdmin)
}

func TestBinderRefusesAFieldReadFromThePathAndTheIdentity(t *testing.T) {
	var req struct {
		TenantID string `param:"tenant" identity:"tenant" json:"-"`
	}

	require.ErrorContains(t, bindAs(t, &identity.Identity{}, "path-tenant", `{}`, &req), "both param and identity")
}

func TestBinderRefusesAnIdentityTagNamingNoClaim(t *testing.T) {
	var req struct {
		Value string `identity:"no-such-value" json:"-"`
	}

	require.ErrorContains(t, bindAs(t, &identity.Identity{}, "", `{}`, &req), "names no claim")
}

func TestBinderFillsEveryIdentityClaim(t *testing.T) {
	type claims struct {
		ID      string          `identity:"id" json:"-"`
		Tenant  string          `identity:"tenant" json:"-"`
		Device  string          `identity:"device" json:"-"`
		Role    authorizer.Role `identity:"role" json:"-"`
		Admin   bool            `identity:"admin" json:"-"`
		AdminID string          `identity:"admin_id" json:"-"`
	}

	caller := identity.Identity{ID: "user", TenantID: "tenant", DeviceUID: "device", Role: authorizer.RoleOwner, Admin: true, AdminID: "admin"}

	var req claims
	require.NoError(t, bindAs(t, &caller, "", `{}`, &req))

	assert.Equal(t, claims{ID: "user", Tenant: "tenant", Device: "device", Role: authorizer.RoleOwner, Admin: true, AdminID: "admin"}, req)
}

func TestEveryIdentityClaimBindsAField(t *testing.T) {
	for _, claim := range binding.IdentityClaims {
		t.Run(claim, func(t *testing.T) {
			assert.NotPanics(t, func() { identityField(identity.Identity{}, claim) })
		})
	}
}
