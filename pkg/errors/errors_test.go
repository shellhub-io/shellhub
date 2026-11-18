package errors

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithData(t *testing.T) {
	sentinel := New("not found", "service", CodeNotFound)

	t.Run("sets the data of the Error in the chain", func(t *testing.T) {
		got := WithData(Wrap(sentinel, nil), InvalidFields{Fields: map[string]string{"name": "invalid"}})
		assert.Equal(t, Error{Message: "not found", Layer: "service", Code: CodeNotFound, Data: InvalidFields{Fields: map[string]string{"name": "invalid"}}}, got)
	})

	t.Run("keeps a parent that holds no Error", func(t *testing.T) {
		plain := stderrors.New("connection reset")
		assert.Equal(t, plain, WithData(plain, nil))
	})

	t.Run("keeps a nil parent nil", func(t *testing.T) {
		assert.NoError(t, WithData(nil, nil))
	})
}
