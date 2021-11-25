package services

import (
	"github.com/shellhub-io/shellhub/api/pkg/guard"
)

type MiddlewareService interface {
	CheckPermission(userType string, action int, service func() error) error
}

// CheckPermission checks if an user, through user's type, has permission, according to its type/role on namespace, to execute an action.
func (s *service) CheckPermission(userType string, action int, service func() error) error {
	if !guard.EvaluatePermission(userType, action) {
		return ErrForbidden
	}

	return service()
}
