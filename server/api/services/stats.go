package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type StatsService interface {
	GetStats(ctx context.Context, req *requests.GetStats) (*models.Stats, error)
}

func (s *service) GetStats(ctx context.Context, req *requests.GetStats) (*models.Stats, error) {
	// An absent tenant is the instance-wide statistics request from the admin surface, which counts
	// devices and sessions across every namespace by design.
	sc := scope.NewUnbounded("instance-wide statistics deliberately aggregate every namespace")
	if req.TenantID != "" {
		var err error
		if sc, err = BoundTo(req.TenantID); err != nil {
			return nil, err
		}
	}

	return s.store.GetStats(ctx, sc)
}
