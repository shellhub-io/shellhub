package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestPublicKeyPushTag(t *testing.T) {
	cases := []struct {
		description string
		fingerprint string
		tenant      string
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "new-tag",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "",
			tag:         "new-tag",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "new-tag",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    nil,
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("public_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.fingerprint != "" && tc.tenant != "" {
				doc := bson.M{
					"fingerprint": tc.fingerprint,
					"tenant_id":   tc.tenant,
					"data":        []byte("test"),
					"created_at":  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					"name":        "public_key",
					"filter": bson.M{
						"hostname": ".*",
						"tags":     []string{},
					},
				}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			err := mongostore.PublicKeyPushTag(ctx, tc.tenant, tc.fingerprint, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPublicKeyPullTag(t *testing.T) {
	cases := []struct {
		description string
		fingerprint string
		tenant      string
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tag",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    nil,
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("public_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.fingerprint != "" && tc.tenant != "" && tc.tag != "" {
				doc := bson.M{
					"fingerprint": tc.fingerprint,
					"tenant_id":   tc.tenant,
					"data":        []byte("test"),
					"created_at":  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					"name":        "public_key",
					"filter": bson.M{
						"hostname": ".*",
						"tags":     []string{tc.tag},
					},
				}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			err := mongostore.PublicKeyPullTag(ctx, tc.tenant, tc.fingerprint, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPublicKeySetTags(t *testing.T) {
	type Expected struct {
		matchedCount int64
		updatedCount int64
		err          error
	}

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		tags        []string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tags:        []string{"tag-1"},
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				matchedCount: 0,
				updatedCount: 0,
				err:          nil,
			},
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "",
			tags:        []string{"tag-1"},
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				matchedCount: 0,
				updatedCount: 0,
				err:          nil,
			},
		},
		{
			description: "succeeds when tags public key is found and tags are equal to current public key tags",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tags:        []string{"tag-1"},
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				matchedCount: 1,
				updatedCount: 0,
				err:          nil,
			},
		},
		{
			description: "succeeds when tags public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tags:        []string{"new-tag"},
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				matchedCount: 1,
				updatedCount: 1,
				err:          nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("public_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.fingerprint != "" && tc.tenant != "" {
				doc := bson.M{
					"fingerprint": tc.fingerprint,
					"tenant_id":   tc.tenant,
					"data":        []byte("test"),
					"created_at":  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					"name":        "public_key",
					"filter": bson.M{
						"hostname": ".*",
						"tags":     tc.tags,
					},
				}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			matchedCount, updatedCount, err := mongostore.PublicKeySetTags(ctx, tc.tenant, tc.fingerprint, tc.tags)
			assert.Equal(t, tc.expected, Expected{matchedCount, updatedCount, err})
		})
	}
}

func TestPublicKeyBulkRenameTag(t *testing.T) {
	type Expected struct {
		count int64
		err   error
	}

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		oldTag      string
		newTag      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when public key is not found due to tenant",
			tenant:      "",
			oldTag:      "tag-1",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "fails when public key is not found due to tag",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "succeeds when public key is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag-1",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				count: 1,
				err:   nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("public_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.oldTag != "" && tc.tenant != "" {
				doc := bson.M{
					"fingerprint": "fingerprint",
					"tenant_id":   tc.tenant,
					"data":        []byte("test"),
					"created_at":  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					"name":        "public_key",
					"filter": bson.M{
						"hostname": ".*",
						"tags":     []string{tc.oldTag},
					},
				}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}
			count, err := mongostore.PublicKeyBulkRenameTag(ctx, tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, Expected{count, err})
		})
	}
}

func TestPublicKeyBulkDeleteTag(t *testing.T) {
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
			description: "fails when public key is not found due to tenant",
			tenant:      "",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "fails when public key is not found due to tag",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "succeeds when public key is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				count: 1,
				err:   nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("public_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.tag != "" && tc.tenant != "" {
				doc := bson.M{
					"fingerprint": "fingerprint",
					"tenant_id":   tc.tenant,
					"data":        []byte("test"),
					"created_at":  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					"name":        "public_key",
					"filter": bson.M{
						"hostname": ".*",
						"tags":     []string{"tag-1"},
					},
				}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}
			count, err := mongostore.PublicKeyBulkDeleteTag(context.TODO(), tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, Expected{count, err})
		})
	}
}

func TestPublicKeyGetTags(t *testing.T) {
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
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				tags: []string{"tag-1"},
				len:  1,
				err:  nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("public_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			doc := bson.M{
				"fingerprint": "fingerprint",
				"tenant_id":   tc.tenant,
				"data":        []byte("test"),
				"created_at":  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				"name":        "public_key",
				"filter": bson.M{
					"hostname": ".*",
					"tags":     []string{"tag-1"},
				},
			}
			testData = append(testData, doc)

			if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}

			tags, count, err := mongostore.PublicKeyGetTags(context.TODO(), tc.tenant)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})
		})
	}
}
