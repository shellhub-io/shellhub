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

func (n *nullCache) Get(_ context.Context, _ string, _ interface{}) error {
	return nil
}

func (n *nullCache) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
	return nil
}

func (n *nullCache) Delete(_ context.Context, _ string) error {
	return nil
}
