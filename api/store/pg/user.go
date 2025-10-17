package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) UserList(ctx context.Context, opts ...store.QueryOption) ([]models.User, int, error)
func (pg *Pg) UserCreate(ctx context.Context, user *models.User) (insertedID string, err error)
func (pg *Pg) UserCreateInvited(ctx context.Context, email string) (insertedID string, err error)
func (pg *Pg) UserResolve(ctx context.Context, resolver store.UserResolver, value string, opts ...store.QueryOption) (*models.User, error)
func (pg *Pg) UserConflicts(ctx context.Context, target *models.UserConflicts) (conflicts []string, has bool, err error)
func (pg *Pg) UserUpdate(ctx context.Context, user *models.User) error
func (pg *Pg) UserGetInfo(ctx context.Context, id string) (userInfo *models.UserInfo, err error)
func (pg *Pg) UserDelete(ctx context.Context, user *models.User) error
