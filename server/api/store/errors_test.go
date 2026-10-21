package store

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		assert.Empty(t, field)
	})

	t.Run("empty-field DuplicateFieldError returns empty string and false", func(t *testing.T) {
		err := errors.Join(ErrDuplicate, DuplicateFieldError{Field: ""})

		field, ok := DuplicatedField(err)

		assert.False(t, ok)
		assert.Empty(t, field)
	})
}
