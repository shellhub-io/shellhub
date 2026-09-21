package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderRefusal(t *testing.T, c *echo.Context, err error) {
	t.Helper()

	if err == nil {
		return
	}

	handlers.NewErrors(nil)(c, err)
}

func assertRefusalCarriesABody(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	body := responses.Error{} //nolint:exhaustruct // the decoder fills it from the response body
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), "a refusal must answer a JSON error body")
	assert.NotEmpty(t, body.Message, "a refusal's body must say why")
}
