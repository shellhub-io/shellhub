package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
)

type deviceSelectorSourceFunc struct {
	permission authorizer.Permission
	list       func(ctx context.Context, sc scope.Scope) ([]DeviceSelector, error)
}

func (f deviceSelectorSourceFunc) EditPermission() authorizer.Permission {
	return f.permission
}

func (f deviceSelectorSourceFunc) DeviceSelectors(ctx context.Context, sc scope.Scope) ([]DeviceSelector, error) {
	return f.list(ctx, sc)
}
