package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestFirewallRulePushTag(t *testing.T) {
	cases := []struct {
		description string
		id          string
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds to add a new tag when firewall rule is found and tag is not set yet",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag4",
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected:    nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("firewall_rules")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.id != "" && tc.tag != "" {
				objID, err := primitive.ObjectIDFromHex(tc.id)
				if err != nil {
					t.Fatalf("failed to convert ID to ObjectID: %v", err)
				}

				doc := bson.M{"_id": objID, "tag": "old"}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			err := mongostore.FirewallRulePushTag(context.TODO(), tc.id, tc.tag)
			assert.Equal(t, tc.expected, err)

			if err := dbtest.DeleteMockData(ctx, collection); err != nil {
				t.Fatalf("failed to clean database: %v", err)
			}
		})
	}
}

func TestFirewallRulePullTag(t *testing.T) {
	cases := []struct {
		description string
		id          string
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "successfully removes tag from an existing firewall rule",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected:    nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("firewall_rules")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown()

			var testData []interface{}
			if tc.id != "" && tc.tag != "" {
				doc := bson.M{"_id": tc.id, "tag": "tag-1"}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			err := mongostore.FirewallRulePullTag(context.TODO(), tc.id, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestFirewallRuleSetTags(t *testing.T) {
	type Expected struct {
		err error
	}
	cases := []struct {
		description string
		id          string
		tags        []string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "successfully set tags for an existing firewall rule",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tags:        []string{"tag-1", "tag-2"},
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				err: nil,
			},
		},
		{
			description: "fails when firewall rule is not found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				err: store.ErrNoDocuments,
			},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("firewall_rules")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if len(tc.tags) > 0 {
				for _, tag := range tc.tags {
					doc := bson.M{"_id": tc.id, "filter": bson.M{"tags": []string{tag}}}
					testData = append(testData, doc)
				}

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			err := mongostore.FirewallRuleSetTags(context.TODO(), tc.id, tc.tags)
			assert.Equal(t, tc.expected.err, err)
		})
	}
}

func TestFirewallRuleBulkRenameTags(t *testing.T) {
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
			description: "fails when tenant is not found",
			tenant:      "",
			oldTag:      "tag-1",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "fails when tag is not found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "succeeds when tenant and tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag-1",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("firewall_rules")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.tenant != "" && tc.oldTag != "" {
				var testData []interface{}
				doc := bson.M{"tenant": tc.tenant, "old_tag": tc.oldTag, "tag": tc.newTag}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			count, err := mongostore.FirewallRuleBulkRenameTag(context.TODO(), tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, Expected{count, err})
		})
	}
}

func TestFirewallRuleBulkDeleteTags(t *testing.T) {
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
			description: "fails when tenant is not found",
			tenant:      "",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "fails when tag is not found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "",
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "succeeds when tenant and tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("firewall_rules")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.tenant != "" && tc.tag != "" {
				var testData []interface{}
				doc := bson.M{"tenant": tc.tenant, "tag": tc.tag}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			count, err := mongostore.FirewallRuleBulkDeleteTag(context.TODO(), tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, Expected{count, err})
		})
	}
}

func TestFirewallRuleGetTags(t *testing.T) {
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
			description: "succeeds when no tags are found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{},
			expected: Expected{
				tags: []string{},
				len:  0,
				err:  nil,
			},
		},
		{
			description: "succeeds when one or more tags are found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixtureFirewallRules},
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

	collection := mongostore.db.Collection("firewall_rules")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if len(tc.expected.tags) > 0 {
				doc := bson.M{"filter.tags": tc.expected.tags, "tenant_id": tc.tenant}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			tags, count, err := mongostore.FirewallRuleGetTags(context.TODO(), tc.tenant)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})
		})
	}
}
