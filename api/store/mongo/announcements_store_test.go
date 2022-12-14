package mongo

import (
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAnnouncementList(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	query := paginator.Query{
		Page:    1,
		PerPage: 3,
	}

	store := NewStore(db.Client().Database("test"), cache.NewNullCache())
	err := store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid0",
		Title:   "title0",
		Content: "content0",
	})
	assert.NoError(t, err)
	err = store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid1",
		Title:   "title1",
		Content: "content1",
	})
	assert.NoError(t, err)
	err = store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid2",
		Title:   "title2",
		Content: "content2",
	})
	assert.NoError(t, err)
	err = store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid3",
		Title:   "title3",
		Content: "content3",
	})
	assert.NoError(t, err)

	announcements, size, err := store.AnnouncementList(data.Context, query)
	assert.NoError(t, err)

	assert.Equal(t, announcements[0].Title, "title0")
	assert.Equal(t, announcements[1].Title, "title1")
	assert.Equal(t, announcements[2].Title, "title2")

	assert.Equal(t, 3, size)
}

func TestAnnouncementGet(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	store := NewStore(db.Client().Database("test"), cache.NewNullCache())
	err := store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid0",
		Title:   "title0",
		Content: "content0",
	})
	assert.NoError(t, err)

	announcement, err := store.AnnouncementGet(data.Context, "uuid0")
	assert.NoError(t, err)

	assert.Equal(t, "title0", announcement.Title)
	assert.Equal(t, "content0", announcement.Content)
}

func TestAnnouncementCreate(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	store := NewStore(db.Client().Database("test"), cache.NewNullCache())
	err := store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid0",
		Title:   "title0",
		Content: "content0",
	})
	assert.NoError(t, err)

	announcement, err := store.AnnouncementGet(data.Context, "uuid0")
	assert.NoError(t, err)

	assert.Equal(t, "title0", announcement.Title)
	assert.Equal(t, "content0", announcement.Content)
}

func TestAnnouncementUpdate(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	store := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid0",
		Title:   "title0",
		Content: "content0",
	})
	assert.NoError(t, err)

	err = store.AnnouncementUpdate(data.Context, &models.Announcement{
		UUID:    "uuid0",
		Title:   "title1",
		Content: "content1",
	})
	assert.NoError(t, err)

	announcement, err := store.AnnouncementGet(data.Context, "uuid0")
	assert.NoError(t, err)

	assert.Equal(t, "title1", announcement.Title)
	assert.Equal(t, "content1", announcement.Content)
}

func TestAnnouncementDelete(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	store := NewStore(db.Client().Database("test"), cache.NewNullCache())
	err := store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid0",
		Title:   "title0",
		Content: "content0",
	})
	assert.NoError(t, err)

	err = store.AnnouncementDelete(data.Context, "uuid0")
	assert.NoError(t, err)
}
