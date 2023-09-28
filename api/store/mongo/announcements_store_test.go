package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/order"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAnnouncementList(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		ann []models.AnnouncementShort
		len int
		err error
	}

	cases := []struct {
		description string
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds when announcement list is empty",
			setup: func() error {
				return nil
			},
			expected: Expected{
				ann: nil,
				len: 0,
				err: nil,
			},
		},
		{
			description: "succeeds when announcement list is not empty",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Announcement)
			},
			expected: Expected{
				ann: []models.AnnouncementShort{
					{
						Date:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UUID:  "00000000-0000-4000-0000-000000000000",
						Title: "title0",
					},
				},
				len: 1,
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			ann, count, err := mongostore.AnnouncementList(ctx, paginator.Query{Page: -1, PerPage: -1}, order.Query{OrderBy: order.Asc})
			assert.Equal(t, tc.expected, Expected{ann: ann, len: count, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestAnnouncementGet(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		ann *models.Announcement
		err error
	}

	cases := []struct {
		description string
		uuid        string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when announcement is not found",
			uuid:        "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Announcement)
			},
			expected: Expected{
				ann: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when announcement is found",
			uuid:        "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Announcement)
			},
			expected: Expected{
				ann: &models.Announcement{
					Date:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					UUID:    "00000000-0000-4000-0000-000000000000",
					Title:   "title0",
					Content: "content0",
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			ann, err := mongostore.AnnouncementGet(ctx, tc.uuid)
			assert.Equal(t, tc.expected, Expected{ann: ann, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestAnnouncementCreate(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	store := NewStore(db.Client().Database("test"), cache.NewNullCache())

	cases := []struct {
		description  string
		announcement *models.Announcement
		setup        func() error
		expected     error
	}{
		{
			description: "succeeds when data is valid",
			announcement: &models.Announcement{
				UUID:    "00000000-0000-4000-0000-000000000000",
				Title:   "title",
				Content: "content",
			},
			setup: func() error {
				return nil
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = store.AnnouncementCreate(ctx, tc.announcement)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestAnnouncementUpdate(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		ann         *models.Announcement
		setup       func() error
		expected    error
	}{
		{
			description: "fails when announcement is not found",
			ann: &models.Announcement{
				UUID:    "nonexistent",
				Title:   "edited title",
				Content: "edited content",
			},
			setup: func() error {
				return mongotest.UseFixture(fixtures.Announcement)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when announcement is found",
			ann: &models.Announcement{
				UUID:    "00000000-0000-4000-0000-000000000000",
				Title:   "edited title",
				Content: "edited content",
			},
			setup: func() error {
				return mongotest.UseFixture(fixtures.Announcement)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.AnnouncementUpdate(ctx, tc.ann)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestAnnouncementDelete(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		uuid        string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when announcement is not found",
			uuid:        "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Announcement)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when announcement is found",
			uuid:        "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Announcement)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.AnnouncementDelete(ctx, tc.uuid)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
