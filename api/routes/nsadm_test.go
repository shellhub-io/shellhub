package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/stretchr/testify/assert"
)

func TestEditNamespace(t *testing.T) {
	e := echo.New()
	mock := new(mocks.Service)
	ctx := context.TODO()
	h := NewHandler(mock)

	t.Run("Return bad request when the namespace is invalid", func(t *testing.T) {
		data := struct {
			Name string `json:"name"`
		}{"invalid.name"}

		bytesData, _ := json.Marshal(data)

		rec := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPut, "/namespaces/:id", bytes.NewBuffer(bytesData))
		req.Header.Set("X-ID", "ownerID")
		mock.On("EditNamespace", ctx, "123", data.Name, "ownerID").Return(nil, services.ErrInvalidFormat)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		echoContext := e.NewContext(req, rec)
		echoContext.SetParamNames("id")
		echoContext.SetParamValues("123")

		apictx := apicontext.NewContext(mock, echoContext)

		_ = h.EditNamespace(*apictx)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
