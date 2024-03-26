package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
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
		expected    Expected
	}{
		{
			description: "succeeds when announcement list is empty",
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			sorter:      query.Sorter{Order: query.OrderAsc},
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
			description: "succeeds when announcement list is not empty and order is desc",
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			sorter:      query.Sorter{Order: query.OrderDesc},
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

	ctx := context.TODO()

	mongostore := GetMongoStore()
	collection := mongostore.db.Collection("announcements")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var testData []interface{}
			if tc.expected.ann != nil {
				for _, item := range tc.expected.ann {
					testData = append(testData, item)
				}

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			ann, count, err := mongostore.AnnouncementList(context.TODO(), tc.paginator, tc.sorter)
			assert.Equal(t, tc.expected, Expected{ann: ann, len: count, err: err})

			if err := dbtest.DeleteMockData(ctx, collection); err != nil {
				t.Fatalf("failed to clean database: %v", err)
			}

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
		expected    Expected
	}{
		{
			description: "fails when announcement is not found",
			uuid:        "",
			expected: Expected{
				ann: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when announcement is found",
			uuid:        "00000000-0000-4000-0000-000000000000",
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

	ctx := context.TODO()
	mongostore := GetMongoStore()

	collection := mongostore.db.Collection("announcements")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.uuid != "" {
				var testData []interface{}

				testData = append(testData, tc.expected.ann)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			ann, err := mongostore.AnnouncementGet(context.TODO(), tc.uuid)
			assert.Equal(t, tc.expected, Expected{ann: ann, err: err})
		})
	}
}

func TestAnnouncementCreate(t *testing.T) {
	cases := []struct {
		description  string
		announcement *models.Announcement
		expected     error
	}{
		{
			description: "succeeds when data is valid",
			announcement: &models.Announcement{
				UUID:    "00000000-0000-40004-0000-000000000000",
				Title:   "title",
				Content: "content",
			},
			expected: nil,
		},
	}
	ctx := context.TODO()
	mongostore := GetMongoStore()

	collection := mongostore.db.Collection("announcements")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {

			var testData []interface{}
			testData = append(testData, tc.announcement)

			if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}

			err := mongostore.AnnouncementCreate(context.TODO(), tc.announcement)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestAnnouncementUpdate(t *testing.T) {
	cases := []struct {
		description string
		ann         *models.Announcement
		expected    error
	}{
		{
			description: "fails when announcement is not found",
			ann:         &models.Announcement{},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when announcement is found",
			ann: &models.Announcement{
				UUID:    "00000000-0000-4000-0000-000000000000",
				Title:   "edited title",
				Content: "edited content",
				Date:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: nil,
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()

	collection := mongostore.db.Collection("announcements")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {

			if tc.ann.UUID != "" {
				var testData []interface{}

				testData = append(testData, tc.ann)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			err := mongostore.AnnouncementUpdate(context.TODO(), tc.ann)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestAnnouncementDelete(t *testing.T) {
	cases := []struct {
		description string
		uuid        string
		expected    error
	}{
		{
			description: "fails when announcement is not found",
			uuid:        "",
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when announcement is found",
			uuid:        "00000000-0000-4000-0000-000000000000",
			expected:    nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()

	collection := mongostore.db.Collection("announcements")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {

			if tc.uuid != "" {
				var testData []interface{}
				doc := bson.M{"uuid": tc.uuid}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}
			err := mongostore.AnnouncementDelete(context.TODO(), tc.uuid)
			assert.Equal(t, tc.expected, err)
		})
	}
}
