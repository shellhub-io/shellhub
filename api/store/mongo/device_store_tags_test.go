package mongo

import (
	"context"
	"testing"

	"github.com/pinzolo/mongotest"
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
			tag:         "device1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "successfully creates single tag for an existing device",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tag:         "device1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
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
			tag:         "device1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceWithTag)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when device's tag doesn't exist",
			uid:         models.UID("4500450g5dc4g859858d6e247f3457380136087fd7426d8f337598eef0a120e"),
			tag:         "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceWithTag)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "successfully delete single tag for an existing device",
			uid:         models.UID("4500450g5dc4g859858d6e247f3457380136087fd7426d8f337598eef0a120e"),
			tag:         "device1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceWithTag)
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
			tags:        []string{"device1", "device2"},
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceWithTag)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "successfully update tags for an existing device",
			uid:         models.UID("4500450g5dc4g859858d6e247f3457380136087fd7426d8f337598eef0a120e"),
			tags:        []string{"device1", "device2"},
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceWithTag)
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
		})
	}
}
