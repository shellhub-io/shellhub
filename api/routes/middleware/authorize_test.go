package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/stretchr/testify/assert"
)

func TestRequiresTenant(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	cases := []struct {
		description string
		paramValue  string
		header      string
		admin       bool
		userID      string
		useGateway  bool
		expected    int
	}{
		{
			description: "allows when tenant header matches path param",
			paramValue:  tenant,
			header:      tenant,
			useGateway:  true,
			expected:    http.StatusOK,
		},
		{
			description: "blocks when tenant header does not match path param",
			paramValue:  tenant,
			header:      "7e7389a9-55be-4e14-8c47-817a1552774f",
			useGateway:  true,
			expected:    http.StatusForbidden,
		},
		{
			description: "blocks when tenant header is missing",
			paramValue:  tenant,
			useGateway:  true,
			expected:    http.StatusForbidden,
		},
		{
			description: "blocks when path param is empty",
			header:      tenant,
			useGateway:  true,
			expected:    http.StatusForbidden,
		},
		{
			description: "blocks when context is not a gateway.Context",
			paramValue:  tenant,
			header:      tenant,
			useGateway:  false,
			expected:    http.StatusForbidden,
		},
		{
			description: "allows admin panel proxy cross-tenant (no X-ID, X-Admin true)",
			paramValue:  "7e7389a9-55be-4e14-8c47-817a1552774f",
			header:      tenant,
			admin:       true,
			useGateway:  true,
			expected:    http.StatusOK,
		},
		{
			description: "allows admin panel proxy without tenant header",
			paramValue:  tenant,
			admin:       true,
			useGateway:  true,
			expected:    http.StatusOK,
		},
		{
			description: "blocks admin user on /api surface (X-ID set, X-Admin true)",
			paramValue:  "7e7389a9-55be-4e14-8c47-817a1552774f",
			header:      tenant,
			admin:       true,
			userID:      "admin-id",
			useGateway:  true,
			expected:    http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("X-Tenant-ID", tc.header)
			}
			if tc.admin {
				req.Header.Set("X-Admin", "true")
			}
			if tc.userID != "" {
				req.Header.Set("X-ID", tc.userID)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("tenant")
			c.SetParamValues(tc.paramValue)

			ctx := c
			if tc.useGateway {
				ctx = gateway.NewContext(nil, c)
			}

			next := func(echo.Context) error { return c.NoContent(http.StatusOK) }
			_ = RequiresTenant("tenant")(next)(ctx)
			assert.Equal(t, tc.expected, rec.Result().StatusCode)
		})
	}
}
