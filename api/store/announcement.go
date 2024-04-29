package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type AnnouncementsStore interface {
	AnnouncementCreate(ctx context.Context, announcement *models.Announcement) error
	AnnouncementUpdate(ctx context.Context, announcement *models.Announcement) error
	AnnouncementDelete(ctx context.Context, uuid string) error
}
