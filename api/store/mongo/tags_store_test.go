package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestTagsGet(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
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
			description: "succeeds when tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(
					fixtures.PublicKey,
					fixtures.FirewallRule,
					fixtures.Device,
				)
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

			tags, count, err := mongostore.TagsGet(ctx, tc.tenant)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestTagRename(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
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
			description: "succeeds when tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag1",
			newTag:      "edited-tag",
			setup: func() error {
				return mongotest.UseFixture(
					fixtures.PublicKey,
					fixtures.FirewallRule,
					fixtures.Device,
				)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.TagRename(ctx, tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestTagDelete(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
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
			description: "succeeds when tag is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(
					fixtures.PublicKey,
					fixtures.FirewallRule,
					fixtures.Device,
				)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.TagDelete(ctx, tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
