package mocks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/stretchr/testify/assert"
)

func TestLicenseEvaluatorMock_CanAcceptDevice(t *testing.T) {
	t.Run("returns true when device can be accepted", func(t *testing.T) {
		m := new(mocks.LicenseEvaluator)
		m.On("CanAcceptDevice", context.Background()).Return(true, nil)

		ok, err := m.CanAcceptDevice(context.Background())
		assert.NoError(t, err)
		assert.True(t, ok)

		m.AssertExpectations(t)
	})

	t.Run("returns false when device limit reached", func(t *testing.T) {
		m := new(mocks.LicenseEvaluator)
		m.On("CanAcceptDevice", context.Background()).Return(false, nil)

		ok, err := m.CanAcceptDevice(context.Background())
		assert.NoError(t, err)
		assert.False(t, ok)

		m.AssertExpectations(t)
	})

	t.Run("returns error when check fails", func(t *testing.T) {
		m := new(mocks.LicenseEvaluator)
		expected := errors.New("license fetch error")
		m.On("CanAcceptDevice", context.Background()).Return(false, expected)

		ok, err := m.CanAcceptDevice(context.Background())
		assert.ErrorIs(t, err, expected)
		assert.False(t, ok)

		m.AssertExpectations(t)
	})
}
