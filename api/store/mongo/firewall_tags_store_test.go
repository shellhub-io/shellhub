package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/mongotest"
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
	fixtures.Configure(&db)

	cases := []struct {
		description string
		id          string
		tag         string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when firewall rule is not found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails to add a tag that already exists",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds to add a new tag when firewall rule is found and tag is not set yet",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag4",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.FirewallRuleAddTag(ctx, tc.id, tc.tag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestFirewallRuleRemoveTag(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		id          string
		tag         string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when firewall rule is not found",
			id:          "6504b7bd9b6c4a63a9ccc054",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when firewall rule but tag is not",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when firewall rule and tag is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.FirewallRuleRemoveTag(ctx, tc.id, tc.tag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestFirewallRuleUpdateTag(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		id          string
		tags        []string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when firewall rule is not found",
			id:          "6504b7bd9b6c4a63a9ccc054",
			tags:        []string{"tag1", "tag2"},
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when firewall rule and tag is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tags:        []string{"tag1", "tag2"},
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.FirewallRuleUpdateTags(ctx, tc.id, tc.tags)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestFirewallRuleRenameTags(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		tenant      string
		oldTag      string
		newTag      string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			oldTag:      "tag1",
			newTag:      "edited-tag",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when tag is not found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "nonexistent",
			newTag:      "edited-tag",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when tenant and tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag1",
			newTag:      "edited-tag",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.FirewallRuleRenameTag(ctx, tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestFirewallRuleDeleteTags(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		tenant      string
		tag         string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when tag is not found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when tenant and tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.FirewallRuleDeleteTag(ctx, tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestFirewallRuleGetTags(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		tags []string
		len  int
		err  error
	}

	cases := []struct {
		description string
		tenant      string
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds when no one tag are found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return nil
			},
			expected: Expected{
				tags: []string{},
				len:  0,
				err:  nil,
			},
		},
		{
			description: "succeeds when one or more tags are found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.FirewallRule)
			},
			expected: Expected{
				tags: []string{"tag1"},
				len:  1,
				err:  nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			tags, count, err := mongostore.FirewallRuleGetTags(ctx, tc.tenant)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
