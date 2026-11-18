package errors

import "net/http"

// The kinds of failure an [Error]'s Code names. Every layer whose errors can reach the API uses
// these, so a kind answers with the same HTTP status whichever layer raised it.
const (
	CodeNotFound = iota + 1
	CodeDuplicated
	CodeLimit
	CodeInvalid
	CodePayment
	CodeUnauthorized
	CodeForbidden
	CodeStore
	CodeNotImplemented
	CodeConflict
	CodeLocked
	CodeInternal
	CodeUnprocessable
	CodeTooManyRequests
)

// HTTPStatus returns the status the API answers an error of code with. CodeStore, CodeInternal
// and a code outside the kinds above answer 500.
func HTTPStatus(code int) int {
	switch code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeDuplicated, CodeConflict:
		return http.StatusConflict
	case CodeLimit, CodeForbidden:
		return http.StatusForbidden
	case CodeInvalid:
		return http.StatusBadRequest
	case CodePayment:
		return http.StatusPaymentRequired
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeNotImplemented:
		return http.StatusNotImplemented
	case CodeLocked:
		return http.StatusLocked
	case CodeUnprocessable:
		return http.StatusUnprocessableEntity
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// InvalidFields is the Data of an error that names the request fields at fault and why, so a
// client can mark each one rather than show one message for the whole request.
type InvalidFields struct {
	Fields map[string]string `json:"fields"`
}
