package services

import (
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestBillingEvaluate(t *testing.T) {
	type Expected struct {
		canAccept bool
		err       error
	}

	mock := new(mocks.Store)

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
				clientMock.On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").Return(&models.BillingEvaluation{CanAccept: true, CanConnect: true}, 0, nil).Once()
			},
			expected: Expected{canAccept: true, err: nil},
		},
		{
			description: "fails when client method fails",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				clientMock.On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").Return(&models.BillingEvaluation{CanAccept: true, CanConnect: true}, 0, ErrEvaluate).Once()
			},
			expected: Expected{canAccept: false, err: ErrEvaluate},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(&Keys{
				PrivateKey: privateKey,
				PublicKey:  &privateKey.PublicKey,
			}, store.Store(mock), cache.NewNullCache())

			canAccept, err := service.BillingEvaluate(clientMock, tc.tenant)
			assert.Equal(t, tc.expected, Expected{canAccept: canAccept, err: err})
		})
	}

	mock.AssertExpectations(t)
}

func TestBillingReport(t *testing.T) {
	mock := new(mocks.Store)

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
				clientMock.On("BillingReport", "00000000-0000-0000-0000-000000000000", "device_accept").Return(200, nil).Once()
			},
			expected: nil,
		},
		{
			description: "fails when client response status is 402",
			tenant:      "00000000-0000-0000-0000-000000000000",
			action:      "device_accept",
			requiredMocks: func() {
				clientMock.On("BillingReport", "00000000-0000-0000-0000-000000000000", "device_accept").Return(402, nil).Once()
			},
			expected: ErrPaymentRequired,
		},
		{
			description: "fails when client response status is other than 200 or 402",
			tenant:      "00000000-0000-0000-0000-000000000000",
			action:      "device_accept",
			requiredMocks: func() {
				clientMock.On("BillingReport", "00000000-0000-0000-0000-000000000000", "device_accept").Return(500, nil).Once()
			},
			expected: ErrReport,
		},
		{
			description: "fails when client returns an error",
			tenant:      "00000000-0000-0000-0000-000000000000",
			action:      "device_accept",
			requiredMocks: func() {
				clientMock.On("BillingReport", "00000000-0000-0000-0000-000000000000", "device_accept").Return(0, errors.New("error")).Once()
			},
			expected: errors.New("error"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(&Keys{
				PrivateKey: privateKey,
				PublicKey:  &privateKey.PublicKey,
			}, store.Store(mock), cache.NewNullCache())

			err := service.BillingReport(clientMock, tc.tenant, tc.action)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
