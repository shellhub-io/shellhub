package user

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
)

var ErrUnauthorized = errors.New("unauthorized")

type Service interface {
	UpdateDataUser(ctx context.Context, username, email, currentPassword, newPassword, tenant string) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) UpdateDataUser(ctx context.Context, username, email, currentPassword, newPassword, tenant string) error {
	user, err := s.store.GetUserByTenant(ctx, tenant)
	if err != nil {
		return err
	}
	if newPassword != "" && user.Password != currentPassword {
		return ErrUnauthorized
	}
	return s.store.UpdateUser(ctx, username, email, currentPassword, newPassword, tenant)
}
