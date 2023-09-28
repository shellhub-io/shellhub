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

func TestPublicKeyAddTag(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		tag         string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag0",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			tag:         "tag0",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag0",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.PublicKeyAddTag(ctx, tc.tenant, tc.fingerprint, tc.tag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestPublicKeyRemoveTag(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		tag         string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tag",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag0",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.PublicKeyRemoveTag(ctx, tc.tenant, tc.fingerprint, tc.tag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestPublicKeyUpdateTags(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		tags        []string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tags:        []string{"tag1"},
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			tags:        []string{"tag1"},
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tags:        []string{"tag1"},
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.PublicKeyUpdateTags(ctx, tc.tenant, tc.fingerprint, tc.tags)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestPublicKeyRenameTag(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		oldTag      string
		newTag      string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			oldTag:      "tag1",
			newTag:      "edited-tag",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tag",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag0",
			newTag:      "edited-tag",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag1",
			newTag:      "edited-tag",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.PublicKeyRenameTag(ctx, tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestPublicKeyDeleteTag(t *testing.T) {
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
			description: "fails when public key is not found due to tenant",
			tenant:      "nonexistent",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tag",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag0",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.PublicKeyDeleteTag(ctx, tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestPublicKeyGetTags(t *testing.T) {
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
			description: "succeeds when tags list is greater than 1",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
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

			tags, count, err := mongostore.PublicKeyGetTags(ctx, tc.tenant)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
