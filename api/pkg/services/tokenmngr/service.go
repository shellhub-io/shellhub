package tokenmngr

import (
	"context"
	"time"

	"github.com/speps/go-hashids"

	"github.com/shellhub-io/shellhub/api/pkg/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Service interface {
	CreateToken(ctx context.Context, token models.Token, tenant string) (*models.Token, error)
	GetToken(ctx context.Context, id string) (*models.Token, error)
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) CreateToken(ctx context.Context, token models.Token, tenant string) (*models.Token, error) {
	hd := hashids.NewData()
	hd.MinLength = 6
	hd.Salt = tenant
	hd.Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	h, _ := hashids.NewWithData(hd)
	token.CreatedAt = time.Now()
	id, _ := h.Encode([]int{int(token.CreatedAt.Unix())})
	token.ID = id
	token.TenantID = tenant
	return s.store.CreateToken(ctx, token)
}

func (s *service) GetToken(ctx context.Context, id string) (*models.Token, error) {
	return s.store.GetToken(ctx, id)
}
