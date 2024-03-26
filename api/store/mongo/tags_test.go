package mongo

import (
	"context"
	"sort"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestTagsGet(t *testing.T) {
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
			description: "succeeds when tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixturePublicKeys, fixtures.FixtureFirewallRules, fixtures.FixtureDevices},
			expected: Expected{
				tags: []string{"tag-1"},
				len:  1,
				err:  nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()

	collectionDevices := mongostore.db.Collection("devices")
	collectionPK := mongostore.db.Collection("public_keys")
	collectionfirewall := mongostore.db.Collection("firewall_rules")

	// Due to the non-deterministic order of applying fixtures when dealing with multiple datasets,
	// we ensure that both the expected and result arrays are correctly sorted.
	sort := func(tags []string) {
		sort.Slice(tags, func(i, j int) bool {
			return tags[i] < tags[j]
		})
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			// assert.NoError(t, fixtures.Apply(tc.fixtures...))
			// defer fixtures.Teardown() // nolint: errcheck
			session, err := mongoClient.StartSession()
			assert.NoError(t, err)
			defer session.EndSession(ctx)

			err = session.StartTransaction()
			assert.NoError(t, err)

			devDoc := bson.M{"tenant_id": tc.tenant, "tags": tc.expected.tags}
			if err := dbtest.InsertMockData(ctx, collectionDevices, []interface{}{devDoc}); err != nil {
				t.Fatalf("failed to insert documents into devices collection: %v", err)
			}

			pkDoc := bson.M{"tenant_id": tc.tenant, "filter.tags": tc.expected.tags}
			if err := dbtest.InsertMockData(ctx, collectionPK, []interface{}{pkDoc}); err != nil {
				t.Fatalf("failed to insert documents into public keys collection: %v", err)
			}

			firewallDoc := bson.M{"tenant_id": tc.tenant, "filter.tags": tc.expected.tags}
			if err := dbtest.InsertMockData(ctx, collectionfirewall, []interface{}{firewallDoc}); err != nil {
				t.Fatalf("failed to insert documents into firewall rules collection: %v", err)
			}

			tags, count, err := mongostore.TagsGet(context.TODO(), tc.tenant)
			sort(tc.expected.tags)
			sort(tags)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})
		})
	}
}

func TestTagsRename(t *testing.T) {
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
			description: "succeeds when tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag-1",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FixturePublicKeys, fixtures.FixtureFirewallRules, fixtures.FixtureDevices},
			expected: Expected{
				count: 1,
				err:   nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collectionDevices := mongostore.db.Collection("devices")
	collectionPK := mongostore.db.Collection("public_keys")
	collectionfirewall := mongostore.db.Collection("firewall_rules")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			devDoc := bson.M{"tenant_id": tc.tenant, "tags": []string{tc.oldTag}}
			if err := dbtest.InsertMockData(ctx, collectionDevices, []interface{}{devDoc}); err != nil {
				t.Fatalf("failed to insert documents into devices collection: %v", err)
			}

			pkDoc := bson.M{"tenant_id": tc.tenant, "tags": tc.oldTag}
			if err := dbtest.InsertMockData(ctx, collectionPK, []interface{}{pkDoc}); err != nil {
				t.Fatalf("failed to insert documents into public keys collection: %v", err)
			}

			firewallDoc := bson.M{"tenant_id": tc.tenant, "tags": tc.oldTag}
			if err := dbtest.InsertMockData(ctx, collectionfirewall, []interface{}{firewallDoc}); err != nil {
				t.Fatalf("failed to insert documents into firewall rules collection: %v", err)
			}

			count, err := mongostore.TagsRename(context.TODO(), tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, Expected{count, err})
		})
	}
}

func TestTagsDelete(t *testing.T) {
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
			description: "succeeds when tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixturePublicKeys, fixtures.FixtureFirewallRules, fixtures.FixtureDevices},
			expected: Expected{
				count: 1,
				err:   nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collectionDevices := mongostore.db.Collection("devices")
	collectionPK := mongostore.db.Collection("public_keys")
	collectionfirewall := mongostore.db.Collection("firewall_rules")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			devDoc := bson.M{"tenant_id": tc.tenant, "tags": []string{tc.tag}}
			if err := dbtest.InsertMockData(ctx, collectionDevices, []interface{}{devDoc}); err != nil {
				t.Fatalf("failed to insert documents into devices collection: %v", err)
			}

			pkDoc := bson.M{"tenant_id": tc.tenant, "tags": tc.tag}
			if err := dbtest.InsertMockData(ctx, collectionPK, []interface{}{pkDoc}); err != nil {
				t.Fatalf("failed to insert documents into public keys collection: %v", err)
			}

			firewallDoc := bson.M{"tenant_id": tc.tenant, "tags": tc.tag}
			if err := dbtest.InsertMockData(ctx, collectionfirewall, []interface{}{firewallDoc}); err != nil {
				t.Fatalf("failed to insert documents into firewall rules collection: %v", err)
			}

			count, err := mongostore.TagsDelete(context.TODO(), tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, Expected{count, err})
		})
	}
}
