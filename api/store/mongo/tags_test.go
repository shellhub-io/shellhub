package mongo_test

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTagsGet(t *testing.T) {
	type Expected struct {
		err  error
		tags []string
		len  int
	}

	cases := []struct {
		expected    Expected
		description string
		tenant      string
		fixtures    []string
	}{
		{
			description: "succeeds when tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixturePublicKeys, fixtureFirewallRules, fixtureDevices},
			expected: Expected{
				tags: []string{"tag-1"},
				len:  1,
				err:  nil,
			},
		},
	}

	// Due to the non-deterministic order of applying fixtures when dealing with multiple datasets,
	// we ensure that both the expected and result arrays are correctly sorted.
	sort := func(tags []string) {
		sort.Slice(tags, func(i, j int) bool {
			return tags[i] < tags[j]
		})
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			tags, count, err := s.TagsGet(ctx, tc.tenant)

			sort(tc.expected.tags)
			sort(tags)

			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})
		})
	}
}

func TestTagsRename(t *testing.T) {
	type Expected struct {
		err   error
		count int64
	}

	cases := []struct {
		expected    Expected
		description string
		tenant      string
		oldTag      string
		newTag      string
		fixtures    []string
	}{
		{
			description: "succeeds when tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag-1",
			newTag:      "edited-tag",
			fixtures:    []string{fixturePublicKeys, fixtureFirewallRules, fixtureDevices},
			expected: Expected{
				count: 6,
				err:   nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			count, err := s.TagsRename(ctx, tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, Expected{count, err})
		})
	}
}

func TestTagsDelete(t *testing.T) {
	type Expected struct {
		err   error
		count int64
	}

	cases := []struct {
		expected    Expected
		description string
		tenant      string
		tag         string
		fixtures    []string
	}{
		{
			description: "succeeds when tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixturePublicKeys, fixtureFirewallRules, fixtureDevices},
			expected: Expected{
				count: 6,
				err:   nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			count, err := s.TagsDelete(ctx, tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, Expected{count, err})
		})
	}
}
