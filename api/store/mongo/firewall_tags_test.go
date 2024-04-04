package mongo_test

import (
	"context"
	"testing"

	shstore "github.com/shellhub-io/shellhub/api/store"
	"github.com/stretchr/testify/assert"
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
			description: "fails when firewall rule is not found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag-1",
			fixtures:    []string{fixtureFirewallRules},
			expected:    shstore.ErrNoDocuments,
		},
		{
			description: "fails to add a tag that already exists",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag-1",
			fixtures:    []string{fixtureFirewallRules},
			expected:    shstore.ErrNoDocuments,
		},
		{
			description: "succeeds to add a new tag when firewall rule is found and tag is not set yet",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag4",
			fixtures:    []string{fixtureFirewallRules},
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

			err := store.FirewallRulePushTag(ctx, tc.id, tc.tag)
			assert.Equal(t, tc.expected, err)
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
			description: "fails when firewall rule is not found",
			id:          "6504b7bd9b6c4a63a9ccc054",
			tag:         "tag-1",
			fixtures:    []string{fixtureFirewallRules},
			expected:    shstore.ErrNoDocuments,
		},
		{
			description: "fails when firewall rule but tag is not",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "nonexistent",
			fixtures:    []string{fixtureFirewallRules},
			expected:    shstore.ErrNoDocuments,
		},
		{
			description: "succeeds when firewall rule and tag is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tag:         "tag-1",
			fixtures:    []string{fixtureFirewallRules},
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

			err := store.FirewallRulePullTag(ctx, tc.id, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestFirewallRuleSetTags(t *testing.T) {
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
			tags:        []string{"tag-1", "tag2"},
			fixtures:    []string{fixtureFirewallRules},
			expected:    shstore.ErrNoDocuments,
		},
		{
			description: "succeeds when firewall rule and tag is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			tags:        []string{"tag-1", "tag2"},
			fixtures:    []string{fixtureFirewallRules},
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

			err := store.FirewallRuleSetTags(ctx, tc.id, tc.tags)
			assert.Equal(t, tc.expected, err)
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
			tenant:      "nonexistent",
			oldTag:      "tag-1",
			newTag:      "edited-tag",
			fixtures:    []string{fixtureFirewallRules},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "fails when tag is not found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "nonexistent",
			newTag:      "edited-tag",
			fixtures:    []string{fixtureFirewallRules},
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
			fixtures:    []string{fixtureFirewallRules},
			expected: Expected{
				count: 3,
				err:   nil,
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

			count, err := store.FirewallRuleBulkRenameTag(ctx, tc.tenant, tc.oldTag, tc.newTag)
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
			tenant:      "nonexistent",
			tag:         "tag-1",
			fixtures:    []string{fixtureFirewallRules},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "fails when tag is not found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "nonexistent",
			fixtures:    []string{fixtureFirewallRules},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "succeeds when tenant and tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixtureFirewallRules},
			expected: Expected{
				count: 3,
				err:   nil,
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

			count, err := store.FirewallRuleBulkDeleteTag(ctx, tc.tenant, tc.tag)
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
			fixtures:    []string{fixtureFirewallRules},
			expected: Expected{
				tags: []string{"tag-1"},
				len:  1,
				err:  nil,
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

			tags, count, err := store.FirewallRuleGetTags(ctx, tc.tenant)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})
		})
	}
}
