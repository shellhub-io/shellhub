package mongo

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo/options"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var (
	ErrWrongParamsType           = errors.New("wrong parameters type")
	ErrNamespaceDuplicatedMember = errors.New("this member is already in this namespace")
	ErrNamespaceMemberNotFound   = errors.New("this member does not exist in this namespace")
	ErrUserNotFound              = errors.New("user not found")
	ErrStoreParseURI             = errors.New("fail to parse the Mongo URI")
	ErrStoreConnect              = errors.New("fail to connect to the database on Mongo URI")
	ErrStorePing                 = errors.New("fail to ping the Mongo database")
	ErrStoreApplyMigration       = errors.New("fail to apply Mongo migrations")
)

type Store struct {
	db    *mongo.Database
	cache cache.Cache
}

func Connect(ctx context.Context, uri string) (*mongo.Client, *mongo.Database, error) {
	client, err := mongo.Connect(ctx, mongooptions.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, errors.Join(ErrStoreConnect, err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, errors.Join(ErrStorePing, err)
	}

	connStr, err := connstring.ParseAndValidate(uri)
	if err != nil {
		return nil, nil, errors.Join(ErrStoreParseURI, err)
	}

	return client, client.Database(connStr.Database), nil
}

func NewStore(ctx context.Context, db *mongo.Database, cache cache.Cache, opts ...options.DatabaseOpt) (store.Store, error) {
	store := &Store{db: db, cache: cache}

	for _, opt := range opts {
		if err := opt(ctx, store.db); err != nil {
			return nil, err
		}
	}

	return store, nil
}
