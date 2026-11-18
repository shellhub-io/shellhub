package services

import (
	"errors"
	"fmt"
	"testing"

	pkgerrors "github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewErrPublicKeyDataInvalid(t *testing.T) {
	cause := errors.New("invalid encoding")
	value := []byte("invalid-key")

	err := NewErrPublicKeyDataInvalid(value, cause)

	require.ErrorIs(t, err, ErrPublicKeyDataInvalid)
	require.ErrorIs(t, err, cause)
	var payload pkgerrors.Error
	require.ErrorAs(t, err, &payload)
	assert.Equal(t, ErrDataInvalid{Data: map[string]any{"Data": value}}, payload.Data)
}

// TestErrDeviceLicenseLimit asserts that ErrDeviceLicenseLimit is a distinct sentinel
// from ErrDeviceLimit: they share the same layer and code but differ by message,
// so errors.Is must return false when comparing one against the other.
func TestErrDeviceLicenseLimit(t *testing.T) {
	if errors.Is(ErrDeviceLicenseLimit, ErrDeviceLimit) {
		t.Error("ErrDeviceLicenseLimit must NOT match ErrDeviceLimit via errors.Is")
	}
}

func TestUserConflict(t *testing.T) {
	cases := []struct {
		description string
		err         error
		fields      []string
	}{
		{description: "names the field the store reports", err: fmt.Errorf("%w: %w", store.ErrDuplicate, store.DuplicateFieldError{Field: "email"}), fields: []string{"email"}},
		{description: "names no field when the store reports none", err: store.ErrDuplicate, fields: []string{}},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			fields, err := UserConflict(tc.err)
			assert.Equal(t, tc.fields, fields)
			require.ErrorIs(t, err, ErrUserDuplicated)
			require.ErrorIs(t, err, store.ErrDuplicate)
		})
	}

	t.Run("ignores any other error", func(t *testing.T) {
		fields, err := UserConflict(errors.New("connection reset"))
		assert.Nil(t, fields)
		assert.NoError(t, err)
	})
}
