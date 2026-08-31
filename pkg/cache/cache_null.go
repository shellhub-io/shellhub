package cache

import (
	"context"
	"time"
)

type nullCache struct{}

var _ Cache = &nullCache{}

// NewNullCache returns a cache that stores nothing and reports every read as a miss. It is what an
// instance without Redis runs, so a missing cache degrades to recomputing rather than to failing.
func NewNullCache() Cache {
	return &nullCache{}
}

func (*nullCache) Get(_ context.Context, _ string, _ any) error {
	return nil
}

func (*nullCache) Set(_ context.Context, _ string, _ any, _ time.Duration) error {
	return nil
}

func (*nullCache) SetNX(_ context.Context, _ string, _ any, _ time.Duration) (bool, error) {
	return true, nil
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
