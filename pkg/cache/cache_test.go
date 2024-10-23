package cache

import (
	"context"
	"errors"
	"testing"

	cacheMock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGet(t *testing.T) {
	type Test struct {
		Value bool
	}

	cache := new(cacheMock.Cache)

	type Expected struct {
		value *Test
		err   error
	}

	tests := []struct {
		description string
		key         string
		setup       func()
		expected    Expected
	}{
		{
			description: "fail to get the cache due a error",
			key:         "fail_with_error",
			setup: func() {
				cache.On("Get", mock.Anything, "fail_with_error", mock.Anything).
					Return(errors.New("failed with error")).
					Once()
			},
			expected: Expected{nil, errors.New("failed with error")},
		},
		{
			description: "fail to find this vaue from the cache",
			key:         "fail_when_not_found",
			setup: func() {
				cache.On("Get", mock.Anything, "fail_when_not_found", mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						value := args.Get(2).(**Test)
						(*value) = nil
					}).
					Once()
			},
			expected: Expected{nil, ErrGetNotFound},
		},
		{
			description: "success to get data from cache",
			key:         "success",
			setup: func() {
				cache.On("Get", mock.Anything, "success", mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						value := args.Get(2).(**Test)
						(*value) = &Test{
							Value: true,
						}
					}).Once()
			},
			expected: Expected{&Test{
				Value: true,
			}, nil},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ctx := context.Background()

			test.setup()

			value, err := Get[Test](ctx, cache, test.key)
			assert.Equal(t, test.expected, Expected{
				value: value,
				err:   err,
			})
		})
	}

	cache.AssertExpectations(t)
}
