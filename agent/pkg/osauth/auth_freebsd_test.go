//go:build freebsd

package osauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyPasswordHashEmptyHash(t *testing.T) {
	t.Setenv("SHELLHUB_PERMIT_EMPTY_PASSWORDS", "")
	assert.False(t, VerifyPasswordHash("", ""))

	t.Setenv("SHELLHUB_PERMIT_EMPTY_PASSWORDS", "true")
	assert.True(t, VerifyPasswordHash("", ""))
}
