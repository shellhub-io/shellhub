package clock_test

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/stretchr/testify/assert"
)

type sequence []time.Time

func (s *sequence) Now() time.Time {
	head := (*s)[0]
	*s = (*s)[1:]

	return head
}

func TestFreeze(t *testing.T) {
	frozen := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	t.Run("freezes Now until the test ends", func(t *testing.T) {
		clock.Freeze(t, frozen)

		assert.Equal(t, frozen, clock.Now())
		assert.Equal(t, frozen, clock.Now())
	})

	assert.NotEqual(t, frozen, clock.Now())
}

func TestSet(t *testing.T) {
	first := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	second := first.Add(time.Hour)

	t.Run("serves Now from the backend until the test ends", func(t *testing.T) {
		clock.Set(t, &sequence{first, second})

		assert.Equal(t, first, clock.Now())
		assert.Equal(t, second, clock.Now())
	})

	assert.NotEqual(t, first, clock.Now())
}

func TestFreezeNested(t *testing.T) {
	outer := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	inner := outer.Add(time.Hour)

	clock.Freeze(t, outer)

	t.Run("an inner freeze gives back the outer one", func(t *testing.T) {
		clock.Freeze(t, inner)

		assert.Equal(t, inner, clock.Now())
	})

	assert.Equal(t, outer, clock.Now())
}
