package mongo

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
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

var (
	ErrStoreParseURI       = errors.New("fail to parse the Mongo URI")
	ErrStoreConnect        = errors.New("fail to connect to the database on Mongo URI")
	ErrStorePing           = errors.New("fail to ping the Mongo database")
	ErrStoreApplyMigration = errors.New("fail to apply Mongo migrations")
)

func NewStoreMongo(ctx context.Context, cache cache.Cache, uri string) (store.Store, error) {
	connStr, err := connstring.ParseAndValidate(uri)
	if err != nil {
		return nil, errors.Join(ErrStoreParseURI, err)
	}

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, errors.Join(ErrStoreConnect, err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Join(ErrStorePing, err)
	}

	db := client.Database(connStr.Database)

	if err := ApplyMigrations(db); err != nil {
		return nil, errors.Join(ErrStoreApplyMigration, err)
	}

	return &Store{db: db, cache: cache}, nil
}
