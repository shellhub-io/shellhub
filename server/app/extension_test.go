package app

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenExtensionsHandsEachTheStoreTheOneBeforeLeft(t *testing.T) {
	core := storemock.NewMockStore(t)
	wrapped := storemock.NewMockStore(t)

	var seen []store.Store

	wrap := func(_ context.Context, infra Infra) (Features, error) {
		seen = append(seen, infra.Store)

		return Features{Store: wrapped}, nil
	}

	keep := func(_ context.Context, infra Infra) (Features, error) {
		seen = append(seen, infra.Store)

		return Features{}, nil
	}

	_, final, err := openExtensions(context.Background(), []Extension{wrap, keep}, Infra{Store: core})
	require.NoError(t, err)

	assert.Equal(t, []store.Store{core, wrapped}, seen)
	assert.Same(t, wrapped, final)
}

func TestOpenExtensionsStopsAtTheFirstError(t *testing.T) {
	failure := errors.New("extension failed")
	opened := false

	fail := func(context.Context, Infra) (Features, error) { return Features{}, failure }
	after := func(context.Context, Infra) (Features, error) {
		opened = true

		return Features{}, nil
	}

	_, _, err := openExtensions(context.Background(), []Extension{fail, after}, Infra{})

	require.ErrorIs(t, err, failure)
	assert.False(t, opened)
}
