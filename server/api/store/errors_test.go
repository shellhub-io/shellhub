package store

import (
	"errors"
	"testing"

	pkgerrors "github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrInternal(t *testing.T) {
	t.Run("is non-nil", func(t *testing.T) {
		assert.NotNil(t, ErrInternal)
	})

	t.Run("has correct layer", func(t *testing.T) {
		var e pkgerrors.Error
		assert.True(t, pkgerrors.As(ErrInternal, &e))
		assert.Equal(t, ErrLayer, e.Layer)
	})

	t.Run("ErrCodeInternal does not collide with existing codes", func(t *testing.T) {
		assert.NotEqual(t, ErrCodeInternal, ErrCodeNoDocument)
		assert.NotEqual(t, ErrCodeInternal, ErrCodeDuplicated)
		assert.NotEqual(t, ErrCodeInternal, ErrCodeInvalid)
	})

	t.Run("ErrInternal carries ErrCodeInternal", func(t *testing.T) {
		var e pkgerrors.Error
		assert.True(t, pkgerrors.As(ErrInternal, &e))
		assert.Equal(t, ErrCodeInternal, e.Code)
	})
}

func TestDuplicatedField(t *testing.T) {
	t.Run("joined with ErrDuplicate returns field and true", func(t *testing.T) {
		err := errors.Join(ErrDuplicate, DuplicateFieldError{Field: "email"})

		field, ok := DuplicatedField(err)

		assert.True(t, ok)
		assert.Equal(t, "email", field)
	})

	t.Run("bare ErrDuplicate returns empty string and false", func(t *testing.T) {
		field, ok := DuplicatedField(ErrDuplicate)

		assert.False(t, ok)
		assert.Equal(t, "", field)
	})

	t.Run("empty-field DuplicateFieldError returns empty string and false", func(t *testing.T) {
		err := errors.Join(ErrDuplicate, DuplicateFieldError{Field: ""})

		field, ok := DuplicatedField(err)

		assert.False(t, ok)
		assert.Equal(t, "", field)
	})
}
