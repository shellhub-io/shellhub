package entity

import (
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type Device struct {
	bun.BaseModel `bun:"table:devices"`
	models.Device `bun:"embed:"`
}

type DeviceInfo struct {
	bun.BaseModel     `bun:"table:device_info"`
	models.DeviceInfo `bun:"embed:"`
}

type DevicePosition struct {
	bun.BaseModel         `bun:"table:device_position"`
	models.DevicePosition `bun:"embed:"`
}
