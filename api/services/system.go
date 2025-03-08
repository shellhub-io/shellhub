package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/responses"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

type SystemService interface {
	// GetSystemInfo retrieves the instance's information
	GetSystemInfo(ctx context.Context, req *requests.GetSystemInfo) (*responses.SystemInfo, error)

	SystemDownloadInstallScript(ctx context.Context) (string, error)
}

func (s *service) GetSystemInfo(ctx context.Context, req *requests.GetSystemInfo) (*responses.SystemInfo, error) {
	return nil, nil
}

func (s *service) SystemDownloadInstallScript(_ context.Context) (string, error) {
	return "", nil
}
