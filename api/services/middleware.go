package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
)

type MiddlewareService interface {
	CheckPermission(context context.Context, tenantID, userID string, action int, service func() error) error
}

// CheckPermission checks if an user, through userID, has the permission, according to its type/role on namespace, to execute an action.
func (s *service) CheckPermission(context context.Context, tenantID, userID string, action int, service func() error) error {
	if !guard.EvaluatePermission(context, s.store, tenantID, userID, action) {
		return ErrForbidden
	}

	return service()
}
