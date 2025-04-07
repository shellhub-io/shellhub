package entity

import (
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type PrivateKey struct {
	bun.BaseModel     `bun:"table:private_keys"`
	models.PrivateKey `bun:"embed:"`
}
