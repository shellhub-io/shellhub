package mongo

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/api/order"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAnnouncementList(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	ordination := order.Query{
		OrderBy: order.Asc,
	}

	times := []time.Time{
		time.Now().Add(time.Hour * 2),
		time.Now().Add(time.Hour),
		time.Now().Add(time.Hour * 4),
		time.Now().Add(time.Hour * 3),
		time.Now(),
	}

	store := NewStore(db.Client().Database("test"), cache.NewNullCache())
	err := store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid0",
		Title:   "title0",
		Content: "content0",
		Date:    times[0],
	})
	assert.NoError(t, err)
	err = store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid1",
		Title:   "title1",
		Content: "content1",
		Date:    times[1],
	})
	assert.NoError(t, err)
	err = store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid2",
		Title:   "title2",
		Content: "content2",
		Date:    times[2],
	})
	assert.NoError(t, err)
	err = store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid3",
		Title:   "title3",
		Content: "content3",
		Date:    times[3],
	})
	assert.NoError(t, err)
	err = store.AnnouncementCreate(data.Context, &models.Announcement{
		UUID:    "uuid4",
		Title:   "title4",
		Content: "content4",
		Date:    times[4],
	})
	assert.NoError(t, err)

	announcements, size, err := store.AnnouncementList(data.Context, paginator.Query{
		Page:    1,
		PerPage: 2,
	}, ordination)
	assert.NoError(t, err)

	assert.Equal(t, announcements[0].Title, "title4")
	assert.Equal(t, announcements[1].Title, "title1")
	assert.Equal(t, 2, size)

	announcements, size, err = store.AnnouncementList(data.Context, paginator.Query{
		Page:    2,
		PerPage: 2,
	}, ordination)
	assert.NoError(t, err)

	assert.Equal(t, announcements[0].Title, "title0")
	assert.Equal(t, announcements[1].Title, "title3")
	assert.Equal(t, 2, size)

	announcements, size, err = store.AnnouncementList(data.Context, paginator.Query{
		Page:    3,
		PerPage: 2,
	}, ordination)
	assert.NoError(t, err)

	assert.Equal(t, announcements[0].Title, "title2")
	assert.Equal(t, 1, size)
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
