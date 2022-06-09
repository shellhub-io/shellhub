package routes

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/converter"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestUpdatePendingStatus(t *testing.T) {
	e := echo.New()
	e.Validator = handlers.NewValidator()
	mock := new(mocks.Service)
	ctx := context.TODO()
	h := NewHandler(mock)

	t.Run("Return payment required when the max count is reached", func(t *testing.T) {
		rec := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodPatch, "/devices/:uid/:status", bytes.NewBuffer([]byte{}))
		req.Header.Set("Content-Role", "application/json")
		req.Header.Set("X-Role", guard.RoleOwner)
		echoContext := e.NewContext(req, rec)
		echoContext.SetParamNames("uid", "status")
		echoContext.SetParamValues("123", "pending")
		mock.On("UpdatePendingStatus", ctx, models.UID("123"), "pending", "").Return(svc.ErrDeviceLimit)

		apictx := gateway.NewContext(mock, echoContext)

		output := h.UpdatePendingStatus(*apictx)
		err, ok := output.(errors.Error)
		assert.True(t, ok)

		assert.Equal(t, http.StatusPaymentRequired, converter.FromErrServiceToHTTPStatus(err.Code))
	})
}
