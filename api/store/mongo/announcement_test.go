package mongo_test

import (
	"context"
	"testing"
	"time"

	shstore "github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
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
		paginator   query.Paginator
		sorter      query.Sorter
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when announcement list is empty",
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			sorter:      query.Sorter{Order: query.OrderAsc},
			fixtures:    []string{},
			expected: Expected{
				ann: nil,
				len: 0,
				err: nil,
			},
		},
		{
			description: "succeeds when announcement list is not empty",
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			sorter:      query.Sorter{Order: query.OrderAsc},
			fixtures:    []string{fixtureAnnouncements},
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
			description: "succeeds when announcement list is not empty and paginator and paginator size is limited",
			paginator:   query.Paginator{Page: 2, PerPage: 2},
			sorter:      query.Sorter{Order: query.OrderAsc},
			fixtures:    []string{fixtureAnnouncements},
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
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			sorter:      query.Sorter{Order: query.OrderDesc},
			fixtures:    []string{fixtureAnnouncements},
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			ann, count, err := store.AnnouncementList(ctx, tc.paginator, tc.sorter)
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
			fixtures:    []string{fixtureAnnouncements},
			expected: Expected{
				ann: nil,
				err: shstore.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when announcement is found",
			uuid:        "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtureAnnouncements},
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			ann, err := store.AnnouncementGet(ctx, tc.uuid)
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := store.AnnouncementCreate(ctx, tc.announcement)
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
			fixtures: []string{fixtureAnnouncements},
			expected: shstore.ErrNoDocuments,
		},
		{
			description: "succeeds when announcement is found",
			ann: &models.Announcement{
				UUID:    "00000000-0000-4000-0000-000000000000",
				Title:   "edited title",
				Content: "edited content",
			},
			fixtures: []string{fixtureAnnouncements},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := store.AnnouncementUpdate(ctx, tc.ann)
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
			fixtures:    []string{fixtureAnnouncements},
			expected:    shstore.ErrNoDocuments,
		},
		{
			description: "succeeds when announcement is found",
			uuid:        "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtureAnnouncements},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := store.AnnouncementDelete(ctx, tc.uuid)
			assert.Equal(t, tc.expected, err)
		})
	}
}
