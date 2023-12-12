package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestFirewallRuleAddTag(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	cases := []struct {
		description string
		id          string
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when firewall rule is not found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag1",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails to add a tag that already exists",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag1",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds to add a new tag when firewall rule is found and tag is not set yet",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag4",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.FirewallRuleAddTag(ctx, tc.id, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestFirewallRuleRemoveTag(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	cases := []struct {
		description string
		id          string
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when firewall rule is not found",
			id:          "6504b7bd9b6c4a63a9ccc054",
			tag:         "tag1",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when firewall rule but tag is not",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "nonexistent",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when firewall rule and tag is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag1",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.FirewallRuleRemoveTag(ctx, tc.id, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestFirewallRuleUpdateTag(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	cases := []struct {
		description string
		id          string
		tags        []string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when firewall rule is not found",
			id:          "6504b7bd9b6c4a63a9ccc054",
			tags:        []string{"tag1", "tag2"},
			fixtures:    []string{fixtures.FirewallRule},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when firewall rule and tag is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tags:        []string{"tag1", "tag2"},
			fixtures:    []string{fixtures.FirewallRule},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.FirewallRuleUpdateTags(ctx, tc.id, tc.tags)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestFirewallRuleRenameTags(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	cases := []struct {
		description string
		tenant      string
		oldTag      string
		newTag      string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			oldTag:      "tag1",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when tag is not found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "nonexistent",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when tenant and tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag1",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.FirewallRuleRenameTag(ctx, tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestFirewallRuleDeleteTags(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	cases := []struct {
		description string
		tenant      string
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			tag:         "tag1",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when tag is not found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "nonexistent",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when tenant and tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag1",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.FirewallRuleDeleteTag(ctx, tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestFirewallRuleGetTags(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

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
			description: "succeeds when no one tag are found",
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
			fixtures:    []string{fixtures.FirewallRule},
			expected: Expected{
				tags: []string{"tag1"},
				len:  1,
				err:  nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			tags, count, err := mongostore.FirewallRuleGetTags(ctx, tc.tenant)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})
		})
	}
}
