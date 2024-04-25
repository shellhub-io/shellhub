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

func (*nullCache) Get(_ context.Context, _ string, _ interface{}) error {
	return nil
}

func (*nullCache) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
	return nil
}

func (*nullCache) Delete(_ context.Context, _ string) error {
	return nil
}

func (*nullCache) HasAccountLockout(_ context.Context, _, _ string) (int64, int, error) {
	return 0, 0, nil
}

func (*nullCache) StoreLoginAttempt(_ context.Context, _, _ string) (int64, int, error) {
	return 0, 0, nil
}

func (*nullCache) ResetLoginAttempts(_ context.Context, _, _ string) error {
	return nil
}
