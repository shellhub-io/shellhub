package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type AnnouncementsStore interface {
	AnnouncementList(ctx context.Context, paginator query.Paginator, sorter query.Sorter) ([]models.AnnouncementShort, int, error)
	AnnouncementGet(ctx context.Context, uuid string) (*models.Announcement, error)
	AnnouncementCreate(ctx context.Context, announcement *models.Announcement) error
	AnnouncementUpdate(ctx context.Context, announcement *models.Announcement) error
	AnnouncementDelete(ctx context.Context, uuid string) error
}
