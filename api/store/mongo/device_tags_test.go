package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestDevicePushTag(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when device doesn't exist",
			uid:         models.UID(""),
			tag:         "tag4",
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "successfully creates single tag for an existing device",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tag:         "tag4",
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    nil,
		},
	}

	ctx := context.TODO()
	mongostore := NewStore(mongoClient.Database("test"), cache.NewNullCache())
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("devices")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.uid != "" {
				doc := bson.M{"uid": tc.uid}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			err := mongostore.DevicePushTag(context.TODO(), tc.uid, tc.tag)
			if tc.expected != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expected.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if err := dbtest.DeleteMockData(ctx, collection); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}
		})
	}
}

func TestDevicePullTag(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when device doesn't exist",
			uid:         models.UID(""),
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when device's tag doesn't exist",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tag:         "",
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "successfully remove a single tag for an existing device",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    nil,
		},
	}

	ctx := context.TODO()
	mongostore := NewStore(mongoClient.Database("test"), cache.NewNullCache())
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("devices")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.uid != "" && tc.tag != "" {
				doc := bson.M{"uid": tc.uid, "tags": []string{tc.tag}}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			err := mongostore.DevicePullTag(context.TODO(), tc.uid, tc.tag)
			if tc.expected != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expected, err)
			} else {
				assert.NoError(t, err)
			}

			if err := dbtest.DeleteMockData(ctx, collection); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}
		})
	}
}

func TestDeviceSetTags(t *testing.T) {
	type Expected struct {
		matchedCount int64
		updatedCount int64
		err          error
	}
	cases := []struct {
		description string
		uid         models.UID
		tags        []string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "successfully when device doesn't exist",
			uid:         models.UID(""),
			tags:        []string{"new-tag"},
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				matchedCount: 0,
				updatedCount: 0,
				err:          nil,
			},
		},
		{
			description: "successfully when tags are equal to current device's tags",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tags:        []string{"tag-1"},
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				matchedCount: 1,
				updatedCount: 0,
				err:          nil,
			},
		},
		{
			description: "successfully update tags for an existing device",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tags:        []string{"new-tag"},
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				matchedCount: 1,
				updatedCount: 0,
				err:          nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := NewStore(mongoClient.Database("test"), cache.NewNullCache())
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("devices")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.uid != "" && len(tc.tags) > 0 {
				for i := 0; i < len(tc.tags); i++ {
					doc := bson.M{"uid": tc.uid, "tags": tc.tags}
					testData = append(testData, doc)
				}

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			matchedCount, updatedCount, err := mongostore.DeviceSetTags(context.TODO(), tc.uid, tc.tags)
			assert.Equal(t, tc.expected, Expected{matchedCount, updatedCount, err})

			if err := dbtest.DeleteMockData(ctx, collection); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}
		})
	}
}

func TestDeviceBulkRenameTag(t *testing.T) {
	type Expected struct {
		count int64
		err   error
	}

	cases := []struct {
		description string
		tenant      string
		oldTag      string
		newTag      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant doesn't exist",
			tenant:      "",
			oldTag:      "tag-1",
			newTag:      "newtag",
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "fails when device's tag doesn't exist",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "",
			newTag:      "newtag",
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "successfully rename tag for an existing device",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag-1",
			newTag:      "newtag",
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				count: 2,
				err:   nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := NewStore(mongoClient.Database("test"), cache.NewNullCache())
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("devices")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.oldTag != "" && tc.tenant != "" {
				var testData []interface{}
				for i := 0; i < int(tc.expected.count); i++ {
					doc := bson.M{"tenant_id": tc.tenant, "tags": []string{tc.oldTag}}
					testData = append(testData, doc)

				}

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			count, err := mongostore.DeviceBulkRenameTag(context.TODO(), tc.tenant, tc.oldTag, tc.newTag)
			if err != nil {
				t.Fatalf("failed to bulk rename tags: %v", err)
			}

			assert.Equal(t, tc.expected.count, count)

			if err := dbtest.DeleteMockData(ctx, collection); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}
		})
	}
}

func TestDeviceBulkDeleteTag(t *testing.T) {
	type Expected struct {
		count int64
		err   error
	}

	cases := []struct {
		description string
		tenant      string
		tag         string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant doesn't exist",
			tenant:      "",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				count: 0,
				err:   FromMongoError(nil),
			},
		},
		{
			description: "fails when device's tag doesn't exist",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "",
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "successfully delete single tag for an existing device",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				count: 1,
				err:   nil,
			},
		},
	}
	ctx := context.TODO()
	mongostore := NewStore(mongoClient.Database("test"), cache.NewNullCache())
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("devices")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.tenant != "" && tc.tag != "" {
				doc := bson.M{"tenant_id": tc.tenant, "tags": []string{tc.tag}}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			count, err := mongostore.DeviceBulkDeleteTag(context.TODO(), tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, Expected{count, err})

			if err := dbtest.DeleteMockData(ctx, collection); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}
		})
	}
}

func TestDeviceGetTags(t *testing.T) {
	type Expected struct {
		tags []string
		len  int
		err  error
	}

	cases := []struct {
		description string
		tenant      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when tags list is greater than 1",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				tags: []string{"tag-1"},
				len:  1,
				err:  nil,
			},
		},
	}
	ctx := context.TODO()
	mongostore := NewStore(mongoClient.Database("test"), cache.NewNullCache())
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("devices")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			doc := bson.M{
				"tags":      tc.expected.tags,
				"tenant_id": tc.tenant,
			}
			testData = append(testData, doc)

			if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}

			tags, count, err := mongostore.DeviceGetTags(context.TODO(), tc.tenant)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})

			if err := dbtest.DeleteMockData(ctx, collection); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}
		})
	}
}
