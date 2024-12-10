package mongo_test

import (
	"context"
	"sort"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestTagsGet(t *testing.T) {
	type Expected struct {
		tags []models.Tags
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
			fixtures:    []string{fixtureTags, fixturePublicKeys, fixtureFirewallRules, fixtureDevices},
			expected: Expected{
				tags: []models.Tags{
					{
						ID:     "67519c0c31490629a1fc612c",
						Name:   "red",
						Color:  "",
						Tenant: "00000000-0000-4000-0000-000000000000",
					},
					{
						ID:     "67519e4231490629a1fc6130",
						Name:   "blue",
						Color:  "#0000ff",
						Tenant: "00000000-0000-4000-0000-000000000000",
					},
					{
						ID:     "6751a03431490629a1fc6131",
						Name:   "tag-1",
						Color:  "#a25f36",
						Tenant: "00000000-0000-4000-0000-000000000000",
					},
					{
						ID:     "6751b1a93592db0deea3fd97",
						Name:   "green",
						Tenant: "00000000-0000-4000-0000-000000000000",
						Color:  "green",
					},
				},
				len: 4,
				err: nil,
			},
		},
	}

	// Due to the non-deterministic order of applying fixtures when dealing with multiple datasets,
	// we ensure that both the expected and result arrays are correctly sorted.
	sort := func(tags []models.Tags) {
		sort.Slice(tags, func(i, j int) bool {
			return tags[i].Name < tags[j].Name
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
			fixtures:    []string{fixtureTags, fixturePublicKeys, fixtureFirewallRules, fixtureDevices},
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
