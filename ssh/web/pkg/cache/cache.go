package cache

import (
	"context"
	"errors"
	"time"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/ssh/web/pkg/token"
)

var instance cache.Cache

// TTL is the time to live of the token in the cache.
const TTL = time.Second * 30

type Token struct {
	Token       string
	ID          string
	Device      string
	Username    string
	Password    string
	Fingerprint string
	Signature   string
	Data        interface{}
}

type Data struct {
	Device      string
	Username    string
	Password    string
	Fingerprint string
	Signature   string
}

// ConnectRedis connects to redis to be used as cache system.
func ConnectRedis(uri string) error {
	if instance == nil {
		var err error
		instance, err = cache.NewRedisCache(uri)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func getConnection() (cache.Cache, error) { //nolint: ireturn
	if instance == nil {
		return nil, errors.New("cache was not connected")
	}

	return instance, nil
}

// Save saves a data set for TTL time using token as identifier.
func Save(ctx context.Context, token *token.Token, data *Data) (*Token, error) {
	connection, err := getConnection()
	if err != nil {
		return nil, err
	}

	if err := connection.Set(ctx, token.ID, data, TTL); err != nil {
		return nil, err
	}

	return &Token{ //nolint: exhaustruct
		Token: token.Token,
		ID:    token.ID,
		Data:  data,
	}, nil
}

// Restore restores a data set using token as identifier.
func Restore(ctx context.Context, token *token.Token) (*Token, error) {
	connection, err := getConnection()
	if err != nil {
		return nil, err
	}

	var value struct {
		Device      string
		Username    string
		Password    string
		Fingerprint string
		Signature   string
	}

	if err := connection.Get(ctx, token.ID, &value); err != nil {
		return nil, err
	}

	return &Token{ //nolint: exhaustruct
		ID:          token.ID,
		Device:      value.Device,
		Username:    value.Username,
		Password:    value.Password,
		Fingerprint: value.Fingerprint,
		Signature:   value.Signature,
	}, nil
}
