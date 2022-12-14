package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type AnnouncementsStore interface {
	AnnouncementList(ctx context.Context, pagination paginator.Query) ([]models.AnnouncementShort, int, error)
	AnnouncementGet(ctx context.Context, uuid string) (*models.Announcement, error)
	AnnouncementCreate(ctx context.Context, announcement *models.Announcement) error
	AnnouncementUpdate(ctx context.Context, announcement *models.Announcement) error
	AnnouncementDelete(ctx context.Context, uuid string) error
}
