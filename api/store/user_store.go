package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type UserStore interface {
	UserList(ctx context.Context, pagination paginator.Query, filters []models.Filter) ([]models.User, int, error)
	UserCreate(ctx context.Context, user *models.User) error
	UserGetByUsername(ctx context.Context, username string) (*models.User, error)
	UserGetByEmail(ctx context.Context, email string) (*models.User, error)
	UserGetByTenant(ctx context.Context, tenant string) (*models.User, error)
	UserGetByID(ctx context.Context, ID string) (*models.User, error)
	UserDataUpdate(ctx context.Context, data *models.User, ID string) error
	UserPasswordUpdate(ctx context.Context, newPassword, ID string) error
	UserUpdateFromAdmin(ctx context.Context, name, username, email, password, ID string) error
	UserDelete(ctx context.Context, ID string) error
}
