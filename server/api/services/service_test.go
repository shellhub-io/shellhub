package services

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockLicenseEvaluator struct {
	mock.Mock
}

func (m *mockLicenseEvaluator) CanAcceptDevice(ctx context.Context) (bool, error) {
	args := m.Called(ctx)

	return args.Bool(0), args.Error(1)
}

func (m *mockLicenseEvaluator) CanConnectDevice(ctx context.Context) (bool, error) {
	args := m.Called(ctx)

	return args.Bool(0), args.Error(1)
}

func (m *mockLicenseEvaluator) AllowsDeviceHost(ctx context.Context, host string) (bool, error) {
	args := m.Called(ctx, host)

	return args.Bool(0), args.Error(1)
}
