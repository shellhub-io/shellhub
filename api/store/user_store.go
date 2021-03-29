package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type UserStore interface {
	ListUsers(ctx context.Context, pagination paginator.Query, filters []models.Filter) ([]models.User, int, error)
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByTenant(ctx context.Context, tenant string) (*models.User, error)
	GetUserByID(ctx context.Context, ID string) (*models.User, error)
	UpdateUser(ctx context.Context, name, username, email, currentPassword, newPassword, ID string) error
	UpdateUserFromAdmin(ctx context.Context, name, username, email, password, ID string) error
	DeleteUser(ctx context.Context, ID string) error
}
