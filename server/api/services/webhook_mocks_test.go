package services

import (
	"context"

	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/mock"
)

func ignoreWebhookEvents(storeMock *storemock.MockStore) {
	storeMock.On("WebhookEventRecord", mock.Anything, mock.Anything).Return(int64(0), nil).Maybe()
}

func runTransactionsInline(storeMock *storemock.MockStore) {
	storeMock.On("WithTransaction", mock.Anything, mock.AnythingOfType("store.TransactionCb")).
		Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).Maybe()
}
