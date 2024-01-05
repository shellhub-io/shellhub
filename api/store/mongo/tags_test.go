package mongo

import (
	"context"
	"sort"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/stretchr/testify/assert"
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

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	// Due to the non-deterministic order of applying fixtures when dealing with multiple datasets,
	// we ensure that both the expected and result arrays are correctly sorted.
	sort := func(tags []string) {
		sort.Slice(tags, func(i, j int) bool {
			return tags[i] < tags[j]
		})
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			tags, count, err := mongostore.TagsGet(context.TODO(), tc.tenant)
			sort(tc.expected.tags)
			sort(tags)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})
		})
	}
}

func TestTagsRename(t *testing.T) {
	cases := []struct {
		description string
		tenant      string
		oldTag      string
		newTag      string
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag-1",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FixturePublicKeys, fixtures.FixtureFirewallRules, fixtures.FixtureDevices},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.TagsRename(context.TODO(), tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestTagsDelete(t *testing.T) {
	cases := []struct {
		description string
		tenant      string
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixturePublicKeys, fixtures.FixtureFirewallRules, fixtures.FixtureDevices},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.TagsDelete(context.TODO(), tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}
