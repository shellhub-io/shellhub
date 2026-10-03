package uuid_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

const fixed = "00000000-0000-4000-0000-000000000000"

type sequence []string

func (s *sequence) Generate() string {
	head := (*s)[0]
	*s = (*s)[1:]

	return head
}

func TestFix(t *testing.T) {
	t.Run("fixes Generate until the test ends", func(t *testing.T) {
		uuid.Fix(t, fixed)

		assert.Equal(t, fixed, uuid.Generate())
		assert.Equal(t, fixed, uuid.Generate())
	})

	assert.NotEqual(t, fixed, uuid.Generate())
}

func TestSet(t *testing.T) {
	const second = "00000000-0000-4000-0000-000000000001"

	t.Run("serves Generate from the backend until the test ends", func(t *testing.T) {
		uuid.Set(t, &sequence{fixed, second})

		assert.Equal(t, fixed, uuid.Generate())
		assert.Equal(t, second, uuid.Generate())
	})

	assert.NotEqual(t, fixed, uuid.Generate())
}
