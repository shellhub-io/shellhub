package mongo

import (
	"context"
	"testing"
	"time"

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
	type Expected struct {
		ann []models.AnnouncementShort
		len int
		err error
	}

	cases := []struct {
		description string
		page        paginator.Query
		order       order.Query
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when announcement list is empty",
			page:        paginator.Query{Page: -1, PerPage: -1},
			order:       order.Query{OrderBy: order.Asc},
			fixtures:    []string{},
			expected: Expected{
				ann: nil,
				len: 0,
				err: nil,
			},
		},
		{
			description: "succeeds when announcement list is not empty",
			page:        paginator.Query{Page: -1, PerPage: -1},
			order:       order.Query{OrderBy: order.Asc},
			fixtures:    []string{fixtures.FixtureAnnouncements},
			expected: Expected{
				ann: []models.AnnouncementShort{
					{
						Date:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UUID:  "00000000-0000-4000-0000-000000000000",
						Title: "title-0",
					},
					{
						Date:  time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UUID:  "00000000-0000-4001-0000-000000000000",
						Title: "title-1",
					},
					{
						Date:  time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UUID:  "00000000-0000-4002-0000-000000000000",
						Title: "title-2",
					},
					{
						Date:  time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UUID:  "00000000-0000-4003-0000-000000000000",
						Title: "title-3",
					},
				},
				len: 4,
				err: nil,
			},
		},
		{
			description: "succeeds when announcement list is not empty and page and page size is limited",
			page:        paginator.Query{Page: 2, PerPage: 2},
			order:       order.Query{OrderBy: order.Asc},
			fixtures:    []string{fixtures.FixtureAnnouncements},
			expected: Expected{
				ann: []models.AnnouncementShort{
					{
						Date:  time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UUID:  "00000000-0000-4002-0000-000000000000",
						Title: "title-2",
					},
					{
						Date:  time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UUID:  "00000000-0000-4003-0000-000000000000",
						Title: "title-3",
					},
				},
				len: 4,
				err: nil,
			},
		},
		{
			description: "succeeds when announcement list is not empty and order is desc",
			page:        paginator.Query{Page: -1, PerPage: -1},
			order:       order.Query{OrderBy: order.Desc},
			fixtures:    []string{fixtures.FixtureAnnouncements},
			expected: Expected{
				ann: []models.AnnouncementShort{
					{
						Date:  time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UUID:  "00000000-0000-4003-0000-000000000000",
						Title: "title-3",
					},
					{
						Date:  time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UUID:  "00000000-0000-4002-0000-000000000000",
						Title: "title-2",
					},
					{
						Date:  time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UUID:  "00000000-0000-4001-0000-000000000000",
						Title: "title-1",
					},
					{
						Date:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UUID:  "00000000-0000-4000-0000-000000000000",
						Title: "title-0",
					},
				},
				len: 4,
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			ann, count, err := mongostore.AnnouncementList(context.TODO(), tc.page, tc.order)
			assert.Equal(t, tc.expected, Expected{ann: ann, len: count, err: err})
		})
	}
}

func TestAnnouncementGet(t *testing.T) {
	type Expected struct {
		ann *models.Announcement
		err error
	}

	cases := []struct {
		description string
		uuid        string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when announcement is not found",
			uuid:        "nonexistent",
			fixtures:    []string{fixtures.FixtureAnnouncements},
			expected: Expected{
				ann: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when announcement is found",
			uuid:        "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixtureAnnouncements},
			expected: Expected{
				ann: &models.Announcement{
					Date:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					UUID:    "00000000-0000-4000-0000-000000000000",
					Title:   "title-0",
					Content: "content-0",
				},
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			ann, err := mongostore.AnnouncementGet(context.TODO(), tc.uuid)
			assert.Equal(t, tc.expected, Expected{ann: ann, err: err})
		})
	}
}

func TestAnnouncementCreate(t *testing.T) {
	cases := []struct {
		description  string
		announcement *models.Announcement
		fixtures     []string
		expected     error
	}{
		{
			description: "succeeds when data is valid",
			announcement: &models.Announcement{
				UUID:    "00000000-0000-40004-0000-000000000000",
				Title:   "title",
				Content: "content",
			},
			fixtures: []string{},
			expected: nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	store := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := store.AnnouncementCreate(context.TODO(), tc.announcement)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestAnnouncementUpdate(t *testing.T) {
	cases := []struct {
		description string
		ann         *models.Announcement
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when announcement is not found",
			ann: &models.Announcement{
				UUID:    "nonexistent",
				Title:   "edited title",
				Content: "edited content",
			},
			fixtures: []string{fixtures.FixtureAnnouncements},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when announcement is found",
			ann: &models.Announcement{
				UUID:    "00000000-0000-4000-0000-000000000000",
				Title:   "edited title",
				Content: "edited content",
			},
			fixtures: []string{fixtures.FixtureAnnouncements},
			expected: nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.AnnouncementUpdate(context.TODO(), tc.ann)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestAnnouncementDelete(t *testing.T) {
	cases := []struct {
		description string
		uuid        string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when announcement is not found",
			uuid:        "nonexistent",
			fixtures:    []string{fixtures.FixtureAnnouncements},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when announcement is found",
			uuid:        "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixtureAnnouncements},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.AnnouncementDelete(context.TODO(), tc.uuid)
			assert.Equal(t, tc.expected, err)
		})
	}
}
