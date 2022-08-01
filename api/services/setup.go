package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/pkg/api/request"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type SetupService interface {
	Setup(ctx context.Context, req request.Setup) error
}

func (s *service) Setup(ctx context.Context, req request.Setup) error {
	userData := models.UserData{
		Name:     req.Name,
		Email:    req.Email,
		Username: req.Username,
	}

	userPass := models.UserPassword{
		Password: validator.HashPassword(req.Password),
	}

	user := &models.User{
		UserData:     userData,
		UserPassword: userPass,
		Confirmed:    true,
		CreatedAt:    clock.Now(),
	}
	err := s.store.UserCreate(ctx, user)
	if err != nil {
		return NewErrUserDuplicated([]string{req.Username}, err)
	}

	namespace := &models.Namespace{
		Name:       req.Namespace,
		Owner:      user.ID,
		MaxDevices: 0,
		Members: []models.Member{
			{
				ID:   user.ID,
				Role: guard.RoleOwner,
			},
		},
		CreatedAt: clock.Now(),
	}

	_, err = s.store.NamespaceCreate(ctx, namespace)
	if err != nil {
		return NewErrNamespaceDuplicated(err)
	}

	return nil
}
