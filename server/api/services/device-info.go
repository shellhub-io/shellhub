package services

import (
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func deviceInfoFromRequest(info *requests.DeviceInfo) *models.DeviceInfo {
	if info == nil {
		return nil
	}

	return &models.DeviceInfo{
		ID:         info.ID,
		PrettyName: info.PrettyName,
		Version:    info.Version,
		Arch:       info.Arch,
		Platform:   info.Platform,
	}
}
