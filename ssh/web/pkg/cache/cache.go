package cache

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/ssh/web/pkg/token"
)

var instance cache.Cache

func getInstance() (cache.Cache, error) {
	if instance == nil {
		instance, err := cache.NewRedisCache("redis://redis:6379")

		return instance, err
	}

	return instance, nil
}

// CacheTokenTTL is the time to live of the token in the cache.
const CacheTokenTTL = time.Second * 30

type CachedToken struct {
	Token       string
	ID          string
	Device      string
	Username    string
	Password    string
	Fingerprint string
	Signature   string
	Data        interface{}
}

func Token(ctx context.Context, token *token.Token, data interface{}) (*CachedToken, error) {
	cache, err := getInstance()
	if err != nil {
		return nil, err
	}

	if err := cache.Set(ctx, token.ID, data, CacheTokenTTL); err != nil {
		return nil, err
	}

	return &CachedToken{
		Token: token.Token,
		ID:    token.ID,
		Data:  data,
	}, nil
}

func Restore(ctx context.Context, token *token.Token) (*CachedToken, error) {
	cache, err := getInstance()
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

	if err := cache.Get(ctx, token.ID, &value); err != nil {
		return nil, err
	}

	return &CachedToken{
		ID:          token.ID,
		Device:      value.Device,
		Username:    value.Username,
		Password:    value.Password,
		Fingerprint: value.Fingerprint,
		Signature:   value.Signature,
	}, nil
}
