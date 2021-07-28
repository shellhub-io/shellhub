package mongo

import (
	"errors"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrWrongParamsType = errors.New("wrong parameters type")
	ErrDuplicateID     = errors.New("user already member of this namespace")
	ErrUserNotFound    = errors.New("user not found")
)

type Store struct {
	db    *mongo.Database
	cache cache.Cache

	store.Store
}

func NewStore(db *mongo.Database, cache cache.Cache) *Store {
	return &Store{db: db, cache: cache}
}
