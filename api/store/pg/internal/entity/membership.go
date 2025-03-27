package entity

import (
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type Membership struct {
	bun.BaseModel     `bun:"table:memberships"`
	models.Membership `bun:"embed:"`
}
