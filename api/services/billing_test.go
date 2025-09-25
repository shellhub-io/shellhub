package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBillingEvaluate(t *testing.T) {
	type Expected struct {
		canAccept bool
		err       error
	}

	storeMock := new(mocks.Store)

	cases := []struct {
		description   string
		tenant        string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "succeeds when client method succeeds",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				clientMock.On("BillingEvaluate", mock.Anything, "00000000-0000-0000-0000-000000000000").Return(&models.BillingEvaluation{CanAccept: true, CanConnect: true}, 0, nil).Once()
			},
			expected: Expected{canAccept: true, err: nil},
		},
		{
			description: "fails when client method fails",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				clientMock.On("BillingEvaluate", mock.Anything, "00000000-0000-0000-0000-000000000000").Return(&models.BillingEvaluation{CanAccept: true, CanConnect: true}, 0, ErrEvaluate).Once()
			},
			expected: Expected{canAccept: false, err: ErrEvaluate},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, cache.NewNullCache(), clientMock)
			canAccept, err := service.BillingEvaluate(context.Background(), clientMock, tc.tenant)
			assert.Equal(t, tc.expected, Expected{canAccept: canAccept, err: err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestBillingReport(t *testing.T) {
	storeMock := new(mocks.Store)

	cases := []struct {
		description   string
		tenant        string
		action        string
		requiredMocks func()
		expected      error
	}{
		{
			description: "succeeds when client response status is 200",
			tenant:      "00000000-0000-0000-0000-000000000000",
			action:      "device_accept",
			requiredMocks: func() {
				clientMock.On("BillingReport", mock.Anything, "00000000-0000-0000-0000-000000000000", "device_accept").Return(200, nil).Once()
			},
			expected: nil,
		},
		{
			description: "fails when client response status is 402",
			tenant:      "00000000-0000-0000-0000-000000000000",
			action:      "device_accept",
			requiredMocks: func() {
				clientMock.On("BillingReport", mock.Anything, "00000000-0000-0000-0000-000000000000", "device_accept").Return(402, nil).Once()
			},
			expected: ErrPaymentRequired,
		},
		{
			description: "fails when client response status is other than 200 or 402",
			tenant:      "00000000-0000-0000-0000-000000000000",
			action:      "device_accept",
			requiredMocks: func() {
				clientMock.On("BillingReport", mock.Anything, "00000000-0000-0000-0000-000000000000", "device_accept").Return(500, nil).Once()
			},
			expected: ErrReport,
		},
		{
			description: "fails when client returns an error",
			tenant:      "00000000-0000-0000-0000-000000000000",
			action:      "device_accept",
			requiredMocks: func() {
				clientMock.On("BillingReport", mock.Anything, "00000000-0000-0000-0000-000000000000", "device_accept").Return(0, errors.New("error")).Once()
			},
			expected: errors.New("error"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, cache.NewNullCache(), clientMock)
			err := service.BillingReport(context.Background(), clientMock, tc.tenant, tc.action)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}
