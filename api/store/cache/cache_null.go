package cache

import (
	"context"
	"time"
)

type nullCache struct{}

var _ Cache = &nullCache{}

func NewNullCache() Cache {
	return &nullCache{}
}

func (n *nullCache) Get(ctx context.Context, key string, value interface{}) error {
	return nil
}

func (n *nullCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (n *nullCache) Delete(ctx context.Context, key string) error {
	return nil
}
