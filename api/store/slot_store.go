package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type SlotStore interface {
	SlotsList(ctx context.Context, tenatn string) ([]models.Slot, error)
	SlotSet(ctx context.Context, tenant string, uid models.UID, status string) error
	SlotDelete(ctx context.Context, tenant string, uid models.UID) error
}
