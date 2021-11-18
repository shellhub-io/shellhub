package routes

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/shellhub-io/shellhub/api/apicontext"
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestUpdatePendingStatus(t *testing.T) {
	e := echo.New()
	mock := new(mocks.Service)
	ctx := context.TODO()
	a := NewEchoApp(e, mock)
	a.InitRoutes()
	h := a.RouteHandler()

	t.Run("Return payment required when the max count is reached", func(t *testing.T) {
		rec := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodPatch, "/devices/:uid/:status", bytes.NewBuffer([]byte{}))
		req.Header.Set("Content-Type", "application/json")
		echoContext := a.NewContext(req, rec)
		echoContext.SetParamNames("uid", "status")
		echoContext.SetParamValues("123", "pending")
		mock.On("UpdatePendingStatus", ctx, models.UID("123"), "pending", "", "").Return(svc.ErrMaxDeviceCountReached)
		fmt.Println(h)

		apictx := apicontext.NewContext(mock, echoContext)

		_ = h.UpdatePendingStatus(*apictx)

		assert.Equal(t, http.StatusPaymentRequired, rec.Code)
	})
}
