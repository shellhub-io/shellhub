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
	UserGetByTenant(ctx context.Context, tenantID string) (*models.User, error)
	UserGetByID(ctx context.Context, id string, ns bool) (*models.User, int, error)
	UserUpdateData(ctx context.Context, data *models.User, id string) error
	UserUpdatePassword(ctx context.Context, newPassword string, id string) error
	UserUpdateFromAdmin(ctx context.Context, name string, username string, email string, password string, id string) error
	UserCreateToken(ctx context.Context, token *models.UserTokenRecover) error
	UserGetToken(ctx context.Context, id string) (*models.UserTokenRecover, error)
	UserDeleteTokens(ctx context.Context, id string) error
	UserUpdateAccountStatus(ctx context.Context, id string) error
	UserDetachInfo(ctx context.Context, id string) (map[string][]*models.Namespace, error)
	UserDelete(ctx context.Context, id string) error
}
