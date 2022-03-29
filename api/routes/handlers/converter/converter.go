package converter

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/services"
)

func FromErrServiceToHTTPStatus(code int) int {
	switch code {
	case services.ErrCodeNotFound:
		return http.StatusNotFound
	case services.ErrCodeInvalid:
		return http.StatusBadRequest
	case services.ErrCodeLimit:
		return http.StatusForbidden
	case services.ErrCodePayment:
		return http.StatusPaymentRequired
	case services.ErrCodeDuplicated:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
