package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const PrivateKeyPath = "/var/run/secrets/api_private_key"

type SetupService interface {
	Setup(ctx context.Context, req requests.Setup) error
	SetupVerify(ctx context.Context, sign string) error
}

func (s *service) Setup(ctx context.Context, req requests.Setup) error {
	return nil
}

func (s *service) SetupVerify(_ context.Context, sign string) error {
	return nil
}
