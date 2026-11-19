package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemocks "github.com/shellhub-io/shellhub/server/api/store/mocks"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthDeviceRefusesAHostTheLicenseDoesNotAllow(t *testing.T) {
	const host = "shellhub.other.com"
	ctx := context.Background()
	req := requests.DeviceAuth{
		TenantID:  "00000000-0000-4000-0000-000000000000",
		Hostname:  "device",
		Identity:  &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
		PublicKey: "public-key",
	}
	storeFailure := errors.New("store reached")

	t.Run("refuses before reaching the store when the evaluator refuses the host", func(t *testing.T) {
		evaluator := new(mockLicenseEvaluator)
		evaluator.On("AllowsDeviceHost", testifymock.Anything, host).Return(false, nil).Once()
		t.Cleanup(func() { evaluator.AssertExpectations(t) })
		svc := NewService(storemocks.NewMockStore(t), privateKey, &privateKey.PublicKey, nil, WithLicenseEvaluator(evaluator))

		_, err := svc.AuthDevice(ctx, req, "10.0.0.1", host)

		require.ErrorIs(t, err, ErrLicenseHostNotAllowed)
	})

	t.Run("returns the evaluator's error", func(t *testing.T) {
		failure := errors.New("license unreadable")
		evaluator := new(mockLicenseEvaluator)
		evaluator.On("AllowsDeviceHost", testifymock.Anything, host).Return(false, failure).Once()
		t.Cleanup(func() { evaluator.AssertExpectations(t) })
		svc := NewService(storemocks.NewMockStore(t), privateKey, &privateKey.PublicKey, nil, WithLicenseEvaluator(evaluator))

		_, err := svc.AuthDevice(ctx, req, "10.0.0.1", host)

		require.ErrorIs(t, err, failure)
	})

	t.Run("goes on to authenticate when the evaluator allows the host", func(t *testing.T) {
		evaluator := new(mockLicenseEvaluator)
		evaluator.On("AllowsDeviceHost", testifymock.Anything, host).Return(true, nil).Once()
		t.Cleanup(func() { evaluator.AssertExpectations(t) })
		st := storemocks.NewMockStore(t)
		st.EXPECT().NamespaceGetWithoutMembers(testifymock.Anything, req.TenantID).Return(nil, storeFailure).Once()
		svc := NewService(st, privateKey, &privateKey.PublicKey, nil, WithLicenseEvaluator(evaluator))

		_, err := svc.AuthDevice(ctx, req, "10.0.0.1", host)

		require.ErrorIs(t, err, storeFailure)
	})
	t.Run("never asks without an evaluator", func(t *testing.T) {
		st := storemocks.NewMockStore(t)
		st.EXPECT().NamespaceGetWithoutMembers(testifymock.Anything, req.TenantID).Return(nil, storeFailure).Once()
		svc := NewService(st, privateKey, &privateKey.PublicKey, nil)

		_, err := svc.AuthDevice(ctx, req, "10.0.0.1", host)

		require.ErrorIs(t, err, storeFailure)
	})
}

func TestAuthDeviceTokenRefusesAHostTheLicenseDoesNotAllow(t *testing.T) {
	const host = "shellhub.other.com"
	ctx := context.Background()
	claims := &authorizer.DeviceClaims{UID: "uid", TenantID: "00000000-0000-4000-0000-000000000000"}

	t.Run("refuses before reaching the store when the evaluator refuses the host", func(t *testing.T) {
		evaluator := new(mockLicenseEvaluator)
		evaluator.On("AllowsDeviceHost", testifymock.Anything, host).Return(false, nil).Once()
		t.Cleanup(func() { evaluator.AssertExpectations(t) })
		svc := NewService(storemocks.NewMockStore(t), privateKey, &privateKey.PublicKey, nil, WithLicenseEvaluator(evaluator))

		require.ErrorIs(t, svc.AuthDeviceToken(ctx, claims, host), ErrLicenseHostNotAllowed)
	})

	t.Run("returns the evaluator's error", func(t *testing.T) {
		failure := errors.New("license unreadable")
		evaluator := new(mockLicenseEvaluator)
		evaluator.On("AllowsDeviceHost", testifymock.Anything, host).Return(false, failure).Once()
		t.Cleanup(func() { evaluator.AssertExpectations(t) })
		svc := NewService(storemocks.NewMockStore(t), privateKey, &privateKey.PublicKey, nil, WithLicenseEvaluator(evaluator))

		require.ErrorIs(t, svc.AuthDeviceToken(ctx, claims, host), failure)
	})

	t.Run("goes on to resolve the device when the evaluator allows the host", func(t *testing.T) {
		evaluator := new(mockLicenseEvaluator)
		evaluator.On("AllowsDeviceHost", testifymock.Anything, host).Return(true, nil).Once()
		t.Cleanup(func() { evaluator.AssertExpectations(t) })
		st := storemocks.NewMockStore(t)
		st.EXPECT().DeviceResolve(testifymock.Anything, testifymock.Anything, store.DeviceUIDResolver, claims.UID).Return(nil, store.ErrNoDocuments).Once()
		svc := NewService(st, privateKey, &privateKey.PublicKey, nil, WithLicenseEvaluator(evaluator))

		require.ErrorIs(t, svc.AuthDeviceToken(ctx, claims, host), ErrAuthUnathorized)
	})

	t.Run("never asks without an evaluator", func(t *testing.T) {
		st := storemocks.NewMockStore(t)
		st.EXPECT().DeviceResolve(testifymock.Anything, testifymock.Anything, store.DeviceUIDResolver, claims.UID).Return(nil, store.ErrNoDocuments).Once()
		svc := NewService(st, privateKey, &privateKey.PublicKey, nil)

		require.ErrorIs(t, svc.AuthDeviceToken(ctx, claims, host), ErrAuthUnathorized)
	})
}
