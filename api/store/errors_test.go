package store

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrInternal(t *testing.T) {
	t.Run("is non-nil", func(t *testing.T) {
		assert.NotNil(t, ErrInternal)
	})

	t.Run("has correct layer", func(t *testing.T) {
		var e errors.Error
		assert.True(t, errors.As(ErrInternal, &e))
		assert.Equal(t, ErrLayer, e.Layer)
	})

	t.Run("ErrCodeInternal does not collide with existing codes", func(t *testing.T) {
		assert.NotEqual(t, ErrCodeInternal, ErrCodeNoDocument)
		assert.NotEqual(t, ErrCodeInternal, ErrCodeDuplicated)
		assert.NotEqual(t, ErrCodeInternal, ErrCodeInvalid)
	})

	t.Run("ErrInternal carries ErrCodeInternal", func(t *testing.T) {
		var e errors.Error
		assert.True(t, errors.As(ErrInternal, &e))
		assert.Equal(t, ErrCodeInternal, e.Code)
	})
}
