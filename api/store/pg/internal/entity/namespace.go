package entity

import (
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type Namespace struct {
	bun.BaseModel    `bun:"table:namespaces"`
	models.Namespace `bun:"embed:"`
}
