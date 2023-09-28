package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestDeviceCreateTag(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		uid         models.UID
		tag         string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when device doesn't exist",
			uid:         models.UID("nonexistent"),
			tag:         "tag4",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "successfully creates single tag for an existing device",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tag:         "tag4",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceCreateTag(ctx, tc.uid, tc.tag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestDeviceRemoveTag(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		uid         models.UID
		tag         string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when device doesn't exist",
			uid:         models.UID("nonexistent"),
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when device's tag doesn't exist",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tag:         "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "successfully remove a single tag for an existing device",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceRemoveTag(ctx, tc.uid, tc.tag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestDeviceUpdateTag(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		uid         models.UID
		tags        []string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when device doesn't exist",
			uid:         models.UID("nonexistent"),
			tags:        []string{"tag0"},
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "successfully update tags for an existing device",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tags:        []string{"tag0"},
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceUpdateTag(ctx, tc.uid, tc.tags)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestDeviceRenameTag(t *testing.T) {
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
			description: "fails when tenant doesn't exist",
			tenant:      "nonexistent",
			oldTag:      "tag1",
			newTag:      "newtag",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when device's tag doesn't exist",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag0",
			newTag:      "newtag",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "successfully rename tag for an existing device",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag1",
			newTag:      "newtag",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceRenameTag(ctx, tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestDeviceDeleteTag(t *testing.T) {
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
			description: "fails when tenant doesn't exist",
			tenant:      "nonexistent",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when device's tag doesn't exist",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag0",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "successfully delete single tag for an existing device",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Device)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceDeleteTag(ctx, tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
