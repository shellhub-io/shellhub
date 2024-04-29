package mongo_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

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

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.AnnouncementCreate(ctx, tc.announcement)
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
			expected: store.ErrNoDocuments,
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

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.AnnouncementUpdate(ctx, tc.ann)
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
			expected:    store.ErrNoDocuments,
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

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.AnnouncementDelete(ctx, tc.uuid)
			assert.Equal(t, tc.expected, err)
		})
	}
}
