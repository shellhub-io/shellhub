package mongo

import (
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrWrongParamsType           = errors.New("wrong parameters type")
	ErrNamespaceDuplicatedMember = errors.New("this member is already in this namespace")
	ErrNamespaceMemberNotFound   = errors.New("this member does not exist in this namespace")
	ErrUserNotFound              = errors.New("user not found")
)

type Store struct {
	db    *mongo.Database
	cache cache.Cache
}

var _ store.Store = (*Store)(nil)

func NewStore(db *mongo.Database, cache cache.Cache) *Store {
	return &Store{db: db, cache: cache}
}

func (s *Store) Database() *mongo.Database {
	return s.db
}

func (s *Store) Cache() cache.Cache {
	return s.cache
}
