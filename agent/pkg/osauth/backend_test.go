package osauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type expiry struct {
	Backend
	expired bool
}

func (e expiry) AccountExpired(string) bool {
	return e.expired
}

func TestSet(t *testing.T) {
	Set(t, expiry{expired: false})

	t.Run("serves AccountExpired from the backend until the test ends", func(t *testing.T) {
		Set(t, expiry{expired: true})

		assert.True(t, AccountExpired("nobody"))
	})

	assert.False(t, AccountExpired("nobody"))
}
